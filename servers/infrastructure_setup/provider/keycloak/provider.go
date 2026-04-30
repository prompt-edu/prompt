// Package keycloak implements the Provider interface for Keycloak group management.
// It uses the Keycloak Admin REST API with client_credentials authentication.
//
// Security note: The service account client must be granted ONLY the
// manage-groups and view-users realm roles — NOT realm-admin.
package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
)

// Config holds credentials for the Keycloak provider.
type Config struct {
	KeycloakURL  string `json:"keycloak_url"`
	Realm        string `json:"realm"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// tokenCache caches the access token to avoid requesting a new token for every call.
type tokenCache struct {
	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

// Provider implements provider.Provider for Keycloak.
type Provider struct {
	cfg    Config
	client *http.Client
	cache  tokenCache
}

// New creates a Keycloak provider from decoded credentials.
func New(cfg Config) *Provider {
	return &Provider{cfg: cfg, client: &http.Client{}}
}

func (p *Provider) GetType() string { return "keycloak" }

func (p *Provider) GetAuthFields() []provider.AuthField {
	return []provider.AuthField{
		{Name: "keycloak_url", Label: "Keycloak URL", Type: "text", Required: true,
			Description: "Keycloak server URL (e.g. https://auth.example.com)"},
		{Name: "realm", Label: "Realm", Type: "text", Required: true,
			Description: "Keycloak realm name"},
		{Name: "client_id", Label: "Client ID", Type: "text", Required: true,
			Description: "Service account client ID with manage-groups and view-users roles"},
		{Name: "client_secret", Label: "Client Secret", Type: "password", Required: true,
			Description: "Service account client secret"},
	}
}

func (p *Provider) ValidateCredentials(ctx context.Context) error {
	token, err := p.getAccessToken(ctx)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/admin/realms/%s", p.cfg.Realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.cfg.KeycloakURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}
	if resp.StatusCode == http.StatusForbidden {
		// Token is valid but lacks realm-admin; that is expected and acceptable.
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("keycloak validate: HTTP %d: %s", resp.StatusCode, body)
}

// CreateResource creates a Keycloak group and adds members.
// It is idempotent: if a group with the same name already exists, it is reused.
func (p *Provider) CreateResource(ctx context.Context, input provider.CreateResourceInput) (*provider.Resource, error) {
	token, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	groupID, err := p.findOrCreateGroup(ctx, token, input.Name)
	if err != nil {
		return nil, err
	}

	for _, member := range input.Members {
		userID, err := p.lookupUserByEmail(ctx, token, member.Email)
		if err != nil {
			// Skip users not found in Keycloak.
			continue
		}
		if err := p.addMemberToGroup(ctx, token, userID, groupID); err != nil {
			return nil, fmt.Errorf("adding user %s to keycloak group: %w", member.Email, err)
		}
	}

	groupURL := fmt.Sprintf("%s/admin/realms/%s/groups/%s", p.cfg.KeycloakURL, p.cfg.Realm, groupID)
	return &provider.Resource{ExternalID: groupID, ExternalURL: groupURL}, nil
}

// DeleteResource removes a Keycloak group by its ID.
func (p *Provider) DeleteResource(ctx context.Context, externalID string) error {
	token, err := p.getAccessToken(ctx)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/admin/realms/%s/groups/%s", p.cfg.Realm, externalID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, p.cfg.KeycloakURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusNotFound {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("keycloak delete group %s: HTTP %d: %s", externalID, resp.StatusCode, body)
}

// findOrCreateGroup returns the group ID for a given name, creating it if needed.
func (p *Provider) findOrCreateGroup(ctx context.Context, token, name string) (string, error) {
	path := fmt.Sprintf("/admin/realms/%s/groups?search=%s", p.cfg.Realm, url.QueryEscape(name))
	body, err := p.get(ctx, token, path)
	if err != nil {
		return "", err
	}

	var groups []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &groups); err != nil {
		return "", err
	}

	for _, g := range groups {
		if g.Name == name {
			return g.ID, nil
		}
	}

	// Create the group.
	payload := map[string]string{"name": name}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	createPath := fmt.Sprintf("/admin/realms/%s/groups", p.cfg.Realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.KeycloakURL+createPath, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("keycloak create group: HTTP %d: %s", resp.StatusCode, respBody)
	}

	// Keycloak returns the group location in the Location header.
	location := resp.Header.Get("Location")
	if location == "" {
		// Fallback: search for the just-created group.
		return p.findGroupByExactName(ctx, token, name)
	}

	// Extract ID from the Location URL path.
	parts := strings.Split(location, "/")
	return parts[len(parts)-1], nil
}

// findGroupByExactName searches for a group by exact name match.
func (p *Provider) findGroupByExactName(ctx context.Context, token, name string) (string, error) {
	path := fmt.Sprintf("/admin/realms/%s/groups?search=%s", p.cfg.Realm, url.QueryEscape(name))
	body, err := p.get(ctx, token, path)
	if err != nil {
		return "", err
	}

	var groups []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &groups); err != nil {
		return "", err
	}
	for _, g := range groups {
		if g.Name == name {
			return g.ID, nil
		}
	}
	return "", fmt.Errorf("keycloak group %q not found after creation", name)
}

// addMemberToGroup adds a user to a Keycloak group.
func (p *Provider) addMemberToGroup(ctx context.Context, token, userID, groupID string) error {
	path := fmt.Sprintf("/admin/realms/%s/users/%s/groups/%s", p.cfg.Realm, userID, groupID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, p.cfg.KeycloakURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("keycloak add user to group: HTTP %d: %s", resp.StatusCode, body)
}

// lookupUserByEmail finds a Keycloak user by email address.
func (p *Provider) lookupUserByEmail(ctx context.Context, token, email string) (string, error) {
	path := fmt.Sprintf("/admin/realms/%s/users?email=%s&exact=true", p.cfg.Realm, url.QueryEscape(email))
	body, err := p.get(ctx, token, path)
	if err != nil {
		return "", err
	}

	var users []struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &users); err != nil {
		return "", err
	}
	if len(users) == 0 {
		return "", fmt.Errorf("keycloak user not found for email: %s", email)
	}
	return users[0].ID, nil
}

// getAccessToken returns a valid access token, refreshing from Keycloak if needed.
func (p *Provider) getAccessToken(ctx context.Context) (string, error) {
	p.cache.mu.Lock()
	defer p.cache.mu.Unlock()

	if p.cache.token != "" && time.Now().Before(p.cache.expiresAt) {
		return p.cache.token, nil
	}

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", p.cfg.KeycloakURL, p.cfg.Realm)
	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {p.cfg.ClientID},
		"client_secret": {p.cfg.ClientSecret},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL,
		strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("keycloak token request failed: HTTP %d: %s", resp.StatusCode, body)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}

	p.cache.token = tokenResp.AccessToken
	// Refresh 30 seconds before expiry to avoid races.
	p.cache.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn-30) * time.Second)
	return p.cache.token, nil
}

// get performs an authenticated GET request to the Keycloak Admin API.
func (p *Provider) get(ctx context.Context, token, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.cfg.KeycloakURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("keycloak GET %s: HTTP %d: %s", path, resp.StatusCode, body)
	}
	return body, nil
}
