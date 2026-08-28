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
	"strconv"
	"strings"

	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
)

// gitlabPageSize is GitLab's maximum per_page value.
const gitlabPageSize = 100

// Config holds the credentials for the GitLab provider.
type Config struct {
	BaseURL      string `json:"base_url"`
	PrivateToken string `json:"private_token"`
	// Credential values are stored and validated as strings, so a numeric ID has to
	// be decoded as a string and parsed on use.
	ParentGroupID string `json:"parent_group_id,omitempty"`
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

func (p *Provider) SupportedResourceTypes() []string { return []string{"group"} }

func (p *Provider) ValidateCredentials(ctx context.Context) error {
	if _, _, err := p.parentGroupID(); err != nil {
		return err
	}
	_, err := p.get(ctx, "/api/v4/user")
	return err
}

// parentGroupID parses the configured parent group ID. The second return value is
// false when no parent is configured.
func (p *Provider) parentGroupID() (int, bool, error) {
	raw := strings.TrimSpace(p.cfg.ParentGroupID)
	if raw == "" {
		return 0, false, nil
	}
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, false, fmt.Errorf("gitlab: parent_group_id must be a positive integer, got %q", raw)
	}
	return id, true, nil
}

// CreateResource creates a GitLab group with the given name and adds members.
// It is idempotent: if a group with the same full_path already exists, it is reused.
func (p *Provider) CreateResource(ctx context.Context, input provider.CreateResourceInput) (*provider.Resource, error) {
	name := sanitizeName(input.Name)
	slug := toSlug(name)
	if slug == "" {
		return nil, fmt.Errorf("gitlab: resource name %q sanitizes to an empty group path", input.Name)
	}

	groupID, groupURL, err := p.findOrCreateGroup(ctx, name, slug)
	if err != nil {
		return nil, err
	}

	var warnings []string
	for _, member := range input.Members {
		permission, ok := input.PermissionMapping[member.Role]
		if !ok {
			warnings = append(warnings, fmt.Sprintf("%s: no permission mapped for role %q", member.Email, member.Role))
			continue
		}
		accessLevel, err := gitlabAccessLevel(permission)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", member.Email, err))
			continue
		}
		if err := p.addMember(ctx, groupID, member.Email, accessLevel); err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", member.Email, err))
		}
	}

	return &provider.Resource{
		ExternalID:  fmt.Sprintf("%d", groupID),
		ExternalURL: groupURL,
		Warnings:    warnings,
	}, nil
}

// findOrCreateGroup returns the group ID and URL for a group, creating it if necessary.
//
// With a parent configured we search that parent's subgroups and compare the path
// exactly. Matching on a full_path suffix across all visible groups would also match
// another course's group of the same name and add students to it.
func (p *Provider) findOrCreateGroup(ctx context.Context, name, slug string) (int, string, error) {
	parentID, hasParent, err := p.parentGroupID()
	if err != nil {
		return 0, "", err
	}

	if id, webURL, found, err := p.findGroup(ctx, slug, parentID, hasParent); err != nil {
		return 0, "", err
	} else if found {
		return id, webURL, nil
	}

	// Create the group.
	payload := map[string]interface{}{
		"name": name,
		"path": slug,
	}
	if hasParent {
		payload["parent_id"] = parentID
	}

	createBody, err := p.post(ctx, "/api/v4/groups", payload)
	if err != nil {
		// A concurrent run, or a group the search could not see, may have taken the
		// path already. Re-resolve instead of failing the instance.
		if id, webURL, found, findErr := p.findGroup(ctx, slug, parentID, hasParent); findErr == nil && found {
			return id, webURL, nil
		}
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

// findGroup looks for a group whose path matches slug exactly.
//
// GitLab's search is a substring match, so "ios-team-1" also matches "ios-team-10".
// The result is paginated to make sure an exact match on a later page is still found;
// concluding "absent" would make the create fail on a taken path.
func (p *Provider) findGroup(ctx context.Context, slug string, parentID int, hasParent bool) (int, string, bool, error) {
	for page := 1; ; page++ {
		var searchPath string
		if hasParent {
			searchPath = fmt.Sprintf("/api/v4/groups/%d/subgroups?search=%s&per_page=%d&page=%d",
				parentID, url.QueryEscape(slug), gitlabPageSize, page)
		} else {
			searchPath = fmt.Sprintf("/api/v4/groups?search=%s&per_page=%d&page=%d",
				url.QueryEscape(slug), gitlabPageSize, page)
		}

		body, err := p.get(ctx, searchPath)
		if err != nil {
			return 0, "", false, err
		}

		var groups []struct {
			ID       int    `json:"id"`
			Path     string `json:"path"`
			FullPath string `json:"full_path"`
			WebURL   string `json:"web_url"`
		}
		if err := json.Unmarshal(body, &groups); err != nil {
			return 0, "", false, err
		}

		for _, g := range groups {
			if hasParent {
				if g.Path == slug {
					return g.ID, g.WebURL, true, nil
				}
			} else if g.FullPath == slug {
				return g.ID, g.WebURL, true, nil
			}
		}

		if len(groups) < gitlabPageSize {
			return 0, "", false, nil
		}
	}
}

// addMember invites a user to a GitLab group by email.
//
// The invitations endpoint is used rather than a user lookup followed by /members:
// /users?search=<email> only returns the email field to admin tokens, so resolving a
// user ID fails for the personal access token a course would normally use. Invitations
// also cover students who have not signed in to the GitLab instance yet.
func (p *Provider) addMember(ctx context.Context, groupID int, email string, accessLevel int) error {
	payload := map[string]interface{}{
		"email":        email,
		"access_level": accessLevel,
	}

	path := fmt.Sprintf("/api/v4/groups/%d/invitations", groupID)
	body, err := p.post(ctx, path, payload)
	if err != nil {
		if strings.Contains(err.Error(), "HTTP 409") {
			return nil
		}
		return err
	}

	// The endpoint answers 201 even when an individual invitation was rejected; the
	// per-email outcome is in the body.
	var resp struct {
		Status  string            `json:"status"`
		Message map[string]string `json:"message"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}
	if resp.Status == "success" {
		return nil
	}
	if reason, ok := resp.Message[email]; ok {
		if isAlreadyMember(reason) {
			return nil
		}
		return fmt.Errorf("gitlab invitation for %s rejected: %s", email, reason)
	}
	return nil
}

func isAlreadyMember(reason string) bool {
	lowered := strings.ToLower(reason)
	return strings.Contains(lowered, "already a member") ||
		strings.Contains(lowered, "already invited") ||
		strings.Contains(lowered, "member already exists")
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
	defer func() { _ = resp.Body.Close() }()

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
	defer func() { _ = resp.Body.Close() }()

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
func gitlabAccessLevel(permission string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(permission)) {
	case "guest":
		return 10, nil
	case "reporter":
		return 20, nil
	case "developer":
		return 30, nil
	case "maintainer":
		return 40, nil
	case "owner":
		return 50, nil
	default:
		return 0, fmt.Errorf("gitlab: unknown permission %q (expected guest, reporter, developer, maintainer or owner)", permission)
	}
}
