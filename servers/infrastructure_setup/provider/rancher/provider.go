// Package rancher implements the Provider interface for Rancher project management.
// It uses the Rancher v3 REST API with Basic Authentication (access key / secret key).
package rancher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
)

// Config holds credentials for the Rancher provider.
type Config struct {
	RancherURL string `json:"rancher_url"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
	ClusterID  string `json:"cluster_id"`
}

// Provider implements provider.Provider for Rancher.
type Provider struct {
	cfg    Config
	client *http.Client
}

// New creates a Rancher provider from decoded credentials.
func New(cfg Config) *Provider {
	return &Provider{cfg: cfg, client: &http.Client{}}
}

func (p *Provider) GetType() string { return "rancher" }

func (p *Provider) GetAuthFields() []provider.AuthField {
	return []provider.AuthField{
		{Name: "rancher_url", Label: "Rancher URL", Type: "text", Required: true,
			Description: "Base URL of your Rancher instance (e.g. https://rancher.example.com)"},
		{Name: "access_key", Label: "Access Key", Type: "text", Required: true,
			Description: "Rancher API access key"},
		{Name: "secret_key", Label: "Secret Key", Type: "password", Required: true,
			Description: "Rancher API secret key"},
		{Name: "cluster_id", Label: "Cluster ID", Type: "text", Required: true,
			Description: "Target cluster ID (e.g. c-xxxxx)"},
	}
}

func (p *Provider) SupportedResourceTypes() []string { return []string{"project"} }

func (p *Provider) ValidateCredentials(ctx context.Context) error {
	_, err := p.get(ctx, "/v3")
	return err
}

// CreateResource creates a Rancher project and adds members.
// The roleTemplateId (e.g. "project-member") is read from ExtraConfig.
func (p *Provider) CreateResource(ctx context.Context, input provider.CreateResourceInput) (*provider.Resource, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, fmt.Errorf("rancher: resource name is empty")
	}

	projectID, projectURL, err := p.findOrCreateProject(ctx, input.Name)
	if err != nil {
		return nil, err
	}

	roleTemplateID := "project-member"
	if rt, ok := input.ExtraConfig["roleTemplateId"].(string); ok && rt != "" {
		roleTemplateID = rt
	}

	var warnings []string
	for _, member := range input.Members {
		perm, ok := input.PermissionMapping[member.Role]
		if !ok {
			perm = roleTemplateID
		}
		if err := p.addMember(ctx, projectID, member.Email, perm); err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", member.Email, err))
		}
	}

	return &provider.Resource{ExternalID: projectID, ExternalURL: projectURL, Warnings: warnings}, nil
}

// findOrCreateProject returns the project ID and URL, creating the project if it does not exist.
func (p *Provider) findOrCreateProject(ctx context.Context, name string) (string, string, error) {
	path := fmt.Sprintf("/v3/projects?name=%s&clusterId=%s", url.QueryEscape(name), p.cfg.ClusterID)
	body, err := p.get(ctx, path)
	if err != nil {
		return "", "", err
	}

	var listResp struct {
		Data []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Links struct {
				Self string `json:"self"`
			} `json:"links"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &listResp); err != nil {
		return "", "", err
	}

	for _, proj := range listResp.Data {
		if strings.EqualFold(proj.Name, name) {
			return proj.ID, proj.Links.Self, nil
		}
	}

	// Create the project.
	payload := map[string]interface{}{
		"name":      name,
		"clusterId": p.cfg.ClusterID,
	}
	createBody, err := p.post(ctx, "/v3/projects", payload)
	if err != nil {
		return "", "", err
	}

	var created struct {
		ID    string `json:"id"`
		Links struct {
			Self string `json:"self"`
		} `json:"links"`
	}
	if err := json.Unmarshal(createBody, &created); err != nil {
		return "", "", err
	}

	return created.ID, created.Links.Self, nil
}

// addMember binds a user to a Rancher project with the given role template.
func (p *Provider) addMember(ctx context.Context, projectID, email, roleTemplateID string) error {
	principalID, err := p.lookupUserPrincipal(ctx, email)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"projectId":       projectID,
		"userPrincipalId": principalID,
		"roleTemplateId":  roleTemplateID,
	}
	_, err = p.post(ctx, "/v3/projectroletemplatebindings", payload)
	return err
}

// lookupUserPrincipal resolves the Rancher principal ID for an email address.
//
// The principals search endpoint is used rather than /v3/users: the v3 user schema has
// no email field, so ?email= is not a supported filter and Rancher answers with an
// unfiltered list. Taking the first entry would bind whichever user happens to come
// first - usually the local admin - into a student project.
//
// The returned principal ID is used as-is. Assembling "local://"+userID is wrong on any
// LDAP- or OIDC-backed Rancher.
func (p *Provider) lookupUserPrincipal(ctx context.Context, email string) (string, error) {
	body, err := p.post(ctx, "/v3/principals?action=search", map[string]interface{}{
		"name":          email,
		"principalType": "user",
	})
	if err != nil {
		return "", err
	}

	var resp struct {
		Data []struct {
			ID            string `json:"id"`
			LoginName     string `json:"loginName"`
			Name          string `json:"name"`
			PrincipalType string `json:"principalType"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("rancher principals parse error: %w", err)
	}

	// Search is a fuzzy match, so the result is confirmed against the address rather
	// than trusted positionally.
	for _, principal := range resp.Data {
		if principal.PrincipalType != "" && principal.PrincipalType != "user" {
			continue
		}
		if principalMatchesEmail(principal.LoginName, principal.Name, email) {
			return principal.ID, nil
		}
	}
	return "", fmt.Errorf("rancher user not found for email: %s", email)
}

// principalMatchesEmail reports whether a search hit really is the requested address.
func principalMatchesEmail(loginName, name, email string) bool {
	return strings.EqualFold(loginName, email) || strings.EqualFold(name, email)
}

// get performs an authenticated GET request.
func (p *Provider) get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.cfg.RancherURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(p.cfg.AccessKey, p.cfg.SecretKey)

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
		return nil, fmt.Errorf("rancher GET %s: HTTP %d: %s", path, resp.StatusCode, body)
	}
	return body, nil
}

// post performs an authenticated POST request with a JSON body.
func (p *Provider) post(ctx context.Context, path string, payload interface{}) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.RancherURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(p.cfg.AccessKey, p.cfg.SecretKey)
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
		return nil, fmt.Errorf("rancher POST %s: HTTP %d: %s", path, resp.StatusCode, body)
	}
	return body, nil
}
