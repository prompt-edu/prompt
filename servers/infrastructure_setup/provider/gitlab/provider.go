// Package gitlab implements the Provider interface for GitLab group management.
// It uses the GitLab REST API v4 with a personal access token.
package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
)

// Config holds the credentials for the GitLab provider.
type Config struct {
	BaseURL       string `json:"base_url"`
	PrivateToken  string `json:"private_token"`
	ParentGroupID *int   `json:"parent_group_id,omitempty"`
}

// Provider implements provider.Provider for GitLab.
type Provider struct {
	cfg    Config
	client *http.Client
}

// New creates a GitLab provider from decoded credentials.
func New(cfg Config) *Provider {
	return &Provider{cfg: cfg, client: &http.Client{}}
}

func (p *Provider) GetType() string { return "gitlab" }

func (p *Provider) GetAuthFields() []provider.AuthField {
	return []provider.AuthField{
		{Name: "base_url", Label: "GitLab URL", Type: "text", Required: true,
			Description: "Your GitLab instance URL (e.g. https://gitlab.com)"},
		{Name: "private_token", Label: "Personal Access Token", Type: "password", Required: true,
			Description: "PAT with api scope"},
		{Name: "parent_group_id", Label: "Parent Group ID", Type: "text", Required: false,
			Description: "Numeric ID of the parent group (optional)"},
	}
}

func (p *Provider) ValidateCredentials(ctx context.Context) error {
	_, err := p.get(ctx, "/api/v4/user")
	return err
}

// CreateResource creates a GitLab group with the given name and adds members.
// It is idempotent: if a group with the same full_path already exists, it is reused.
func (p *Provider) CreateResource(ctx context.Context, input provider.CreateResourceInput) (*provider.Resource, error) {
	name := sanitizeName(input.Name)
	slug := toSlug(name)

	groupID, groupURL, err := p.findOrCreateGroup(ctx, name, slug)
	if err != nil {
		return nil, err
	}

	for _, member := range input.Members {
		permission, ok := input.PermissionMapping[member.Role]
		if !ok {
			permission = "guest"
		}
		if err := p.addMember(ctx, groupID, member.Email, permission); err != nil {
			return nil, fmt.Errorf("adding member %s: %w", member.Email, err)
		}
	}

	return &provider.Resource{ExternalID: fmt.Sprintf("%d", groupID), ExternalURL: groupURL}, nil
}

// DeleteResource removes a GitLab group by its numeric ID.
func (p *Provider) DeleteResource(ctx context.Context, externalID string) error {
	path := fmt.Sprintf("/api/v4/groups/%s", externalID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, p.cfg.BaseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("PRIVATE-TOKEN", p.cfg.PrivateToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusAccepted {
		return nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil // already gone, treat as success
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("gitlab delete group %s: HTTP %d: %s", externalID, resp.StatusCode, body)
}

// findOrCreateGroup returns the group ID and URL for a group, creating it if necessary.
func (p *Provider) findOrCreateGroup(ctx context.Context, name, slug string) (int, string, error) {
	searchPath := fmt.Sprintf("/api/v4/groups?search=%s&min_access_level=50", url.QueryEscape(slug))
	body, err := p.get(ctx, searchPath)
	if err != nil {
		return 0, "", err
	}

	var groups []struct {
		ID       int    `json:"id"`
		FullPath string `json:"full_path"`
		WebURL   string `json:"web_url"`
	}
	if err := json.Unmarshal(body, &groups); err != nil {
		return 0, "", err
	}

	// Check for exact full_path match (idempotency).
	for _, g := range groups {
		basePath := slug
		if p.cfg.ParentGroupID != nil {
			// Group full paths include parent namespace; check suffix.
			if strings.HasSuffix(g.FullPath, "/"+slug) {
				return g.ID, g.WebURL, nil
			}
		} else if g.FullPath == basePath {
			return g.ID, g.WebURL, nil
		}
	}

	// Create the group.
	payload := map[string]interface{}{
		"name": name,
		"path": slug,
	}
	if p.cfg.ParentGroupID != nil {
		payload["parent_id"] = *p.cfg.ParentGroupID
	}

	createBody, err := p.post(ctx, "/api/v4/groups", payload)
	if err != nil {
		return 0, "", err
	}

	var created struct {
		ID     int    `json:"id"`
		WebURL string `json:"web_url"`
	}
	if err := json.Unmarshal(createBody, &created); err != nil {
		return 0, "", err
	}

	return created.ID, created.WebURL, nil
}

// addMember adds a user to a GitLab group by email and permission level string.
func (p *Provider) addMember(ctx context.Context, groupID int, email, permission string) error {
	userID, err := p.lookupUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	accessLevel := gitlabAccessLevel(permission)
	payload := map[string]interface{}{
		"user_id":      userID,
		"access_level": accessLevel,
	}

	path := fmt.Sprintf("/api/v4/groups/%d/members", groupID)
	_, err = p.post(ctx, path, payload)
	if err != nil {
		// HTTP 409 means "already a member", treat as success.
		if strings.Contains(err.Error(), "HTTP 409") {
			return nil
		}
		return err
	}
	return nil
}

// lookupUserByEmail searches for a GitLab user by email.
func (p *Provider) lookupUserByEmail(ctx context.Context, email string) (int, error) {
	path := fmt.Sprintf("/api/v4/users?search=%s", url.QueryEscape(email))
	body, err := p.get(ctx, path)
	if err != nil {
		return 0, err
	}

	var users []struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &users); err != nil {
		return 0, err
	}

	for _, u := range users {
		if strings.EqualFold(u.Email, email) {
			return u.ID, nil
		}
	}
	return 0, fmt.Errorf("gitlab user not found for email: %s", email)
}

// get performs an authenticated GET request and returns the response body.
func (p *Provider) get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.cfg.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", p.cfg.PrivateToken)

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
		return nil, fmt.Errorf("gitlab GET %s: HTTP %d: %s", path, resp.StatusCode, body)
	}
	return body, nil
}

// post performs an authenticated POST request with a JSON body.
func (p *Provider) post(ctx context.Context, path string, payload interface{}) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", p.cfg.PrivateToken)
	req.Header.Set("Content-Type", "application/json")

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
		return nil, fmt.Errorf("gitlab POST %s: HTTP %d: %s", path, resp.StatusCode, body)
	}
	return body, nil
}

// sanitizeName removes characters not allowed in GitLab group names.
var illegalChars = regexp.MustCompile(`[^a-zA-Z0-9 ._-]`)

func sanitizeName(name string) string {
	return strings.TrimSpace(illegalChars.ReplaceAllString(name, "-"))
}

// toSlug converts a name to a GitLab-compatible path slug.
var nonAlphanumDash = regexp.MustCompile(`[^a-z0-9-]`)
var multipleDash = regexp.MustCompile(`-+`)

func toSlug(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "-")
	s = nonAlphanumDash.ReplaceAllString(s, "")
	s = multipleDash.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// gitlabAccessLevel maps a permission string to a GitLab numeric access level.
func gitlabAccessLevel(permission string) int {
	switch strings.ToLower(permission) {
	case "guest":
		return 10
	case "reporter":
		return 20
	case "developer":
		return 30
	case "maintainer":
		return 40
	case "owner":
		return 50
	default:
		return 10 // guest as safe default
	}
}
