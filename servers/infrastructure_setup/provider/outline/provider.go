// Package outline implements the Provider interface for Outline collection management.
// It uses the Outline JSON-RPC-style API (POST requests) with an API key.
package outline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
)

// Config holds credentials for the Outline provider.
type Config struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"` // defaults to https://app.getoutline.com/api
}

// Provider implements provider.Provider for Outline.
type Provider struct {
	cfg    Config
	client *http.Client
}

// New creates an Outline provider from decoded credentials.
func New(cfg Config) *Provider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://app.getoutline.com/api"
	}
	return &Provider{cfg: cfg, client: &http.Client{}}
}

func (p *Provider) GetType() string { return "outline" }

func (p *Provider) GetAuthFields() []provider.AuthField {
	return []provider.AuthField{
		{Name: "api_key", Label: "API Key", Type: "password", Required: true,
			Description: "Outline API key starting with ol_api_"},
		{Name: "base_url", Label: "Outline API Base URL", Type: "text", Required: false,
			Description: "Defaults to https://app.getoutline.com/api for Outline Cloud"},
	}
}

func (p *Provider) ValidateCredentials(ctx context.Context) error {
	var resp struct {
		OK   bool `json:"ok"`
		Data struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := p.call(ctx, "auth.info", map[string]interface{}{}, &resp); err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("outline auth.info returned ok=false")
	}
	return nil
}

// CreateResource creates an Outline collection and adds members.
// If a collection with the same name already exists, it is reused.
func (p *Provider) CreateResource(ctx context.Context, input provider.CreateResourceInput) (*provider.Resource, error) {
	collectionID, collectionURL, err := p.findOrCreateCollection(ctx, input.Name)
	if err != nil {
		return nil, err
	}

	for _, member := range input.Members {
		permission, ok := input.PermissionMapping[member.Role]
		if !ok {
			permission = "read"
		}
		userID, err := p.lookupUserByEmail(ctx, member.Email)
		if err != nil {
			// Skip users not found in Outline.
			continue
		}
		if err := p.addMember(ctx, collectionID, userID, permission); err != nil {
			return nil, fmt.Errorf("adding member %s to outline collection: %w", member.Email, err)
		}
	}

	return &provider.Resource{ExternalID: collectionID, ExternalURL: collectionURL}, nil
}

// DeleteResource removes an Outline collection by its ID.
func (p *Provider) DeleteResource(ctx context.Context, externalID string) error {
	var resp struct {
		OK bool `json:"ok"`
	}
	return p.call(ctx, "collections.delete", map[string]interface{}{"id": externalID}, &resp)
}

// findOrCreateCollection returns the ID and URL of a collection, creating it if necessary.
func (p *Provider) findOrCreateCollection(ctx context.Context, name string) (string, string, error) {
	// Outline does not provide a search-by-name endpoint; list all and match.
	var listResp struct {
		OK   bool `json:"ok"`
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"data"`
	}
	if err := p.call(ctx, "collections.list", map[string]interface{}{"limit": 100}, &listResp); err != nil {
		return "", "", err
	}

	for _, c := range listResp.Data {
		if strings.EqualFold(c.Name, name) {
			return c.ID, c.URL, nil
		}
	}

	// Create the collection.
	var createResp struct {
		OK   bool `json:"ok"`
		Data struct {
			Collection struct {
				ID  string `json:"id"`
				URL string `json:"url"`
			} `json:"collection"`
		} `json:"data"`
	}
	if err := p.call(ctx, "collections.create", map[string]interface{}{
		"name":        name,
		"sharing":     false,
		"permission":  "read",
	}, &createResp); err != nil {
		return "", "", err
	}

	if !createResp.OK {
		return "", "", fmt.Errorf("outline collections.create returned ok=false")
	}

	return createResp.Data.Collection.ID, createResp.Data.Collection.URL, nil
}

// addMember adds a user to an Outline collection using memberships.create.
func (p *Provider) addMember(ctx context.Context, collectionID, userID, permission string) error {
	var resp struct {
		OK bool `json:"ok"`
	}
	return p.call(ctx, "memberships.create", map[string]interface{}{
		"collectionId": collectionID,
		"userId":       userID,
		"permission":   permission, // "read", "read_write", or "admin"
	}, &resp)
}

// lookupUserByEmail finds an Outline user by email.
func (p *Provider) lookupUserByEmail(ctx context.Context, email string) (string, error) {
	var listResp struct {
		OK   bool `json:"ok"`
		Data []struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"data"`
		Pagination struct {
			NextPath string `json:"nextPath"`
		} `json:"pagination"`
	}

	offset := 0
	limit := 100
	for {
		if err := p.call(ctx, "users.list", map[string]interface{}{
			"offset": offset,
			"limit":  limit,
		}, &listResp); err != nil {
			return "", err
		}

		for _, u := range listResp.Data {
			if strings.EqualFold(u.Email, email) {
				return u.ID, nil
			}
		}

		if len(listResp.Data) < limit {
			break
		}
		offset += limit
	}

	return "", fmt.Errorf("outline user not found for email: %s", email)
}

// call performs an authenticated POST request to the Outline API.
func (p *Provider) call(ctx context.Context, method string, params map[string]interface{}, result interface{}) error {
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/%s", p.cfg.BaseURL, method), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("outline %s: HTTP %d: %s", method, resp.StatusCode, body)
	}

	return json.Unmarshal(body, result)
}
