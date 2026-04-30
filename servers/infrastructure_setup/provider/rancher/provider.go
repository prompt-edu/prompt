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

func (p *Provider) ValidateCredentials(ctx context.Context) error {
	_, err := p.get(ctx, "/v3")
	return err
}

// CreateResource creates a Rancher project and adds members.
// The roleTemplateId (e.g. "project-member") is read from ExtraConfig.
func (p *Provider) CreateResource(ctx context.Context, input provider.CreateResourceInput) (*provider.Resource, error) {
	projectID, projectURL, err := p.findOrCreateProject(ctx, input.Name)
	if err != nil {
		return nil, err
	}

	roleTemplateID := "project-member"
	if rt, ok := input.ExtraConfig["roleTemplateId"].(string); ok && rt != "" {
		roleTemplateID = rt
	}

	for _, member := range input.Members {
		perm, ok := input.PermissionMapping[member.Role]
		if !ok {
			perm = roleTemplateID
		}
		if err := p.addMember(ctx, projectID, member.Email, perm); err != nil {
			return nil, fmt.Errorf("adding member %s to rancher project: %w", member.Email, err)
		}
	}

	return &provider.Resource{ExternalID: projectID, ExternalURL: projectURL}, nil
}

// DeleteResource removes a Rancher project by its ID.
func (p *Provider) DeleteResource(ctx context.Context, externalID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		p.cfg.RancherURL+"/v3/projects/"+externalID, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(p.cfg.AccessKey, p.cfg.SecretKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("rancher delete project %s: HTTP %d: %s", externalID, resp.StatusCode, body)
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
			ID   string `json:"id"`
			Name string `json:"name"`
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
	userID, err := p.lookupUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"projectId":              projectID,
		"userPrincipalId":        "local://" + userID,
		"roleTemplateId":         roleTemplateID,
	}
	_, err = p.post(ctx, "/v3/projectroletemplatebindings", payload)
	return err
}

// lookupUserByEmail finds a Rancher user ID by email address.
func (p *Provider) lookupUserByEmail(ctx context.Context, email string) (string, error) {
	path := fmt.Sprintf("/v3/users?email=%s", url.QueryEscape(email))
	body, err := p.get(ctx, path)
	if err != nil {
		return "", err
	}

	var listResp struct {
		Data []struct {
			ID    string `json:"id"`
			Email string `json:"principalIds"`
		} `json:"data"`
	}
	// Use a simpler struct since the Rancher user object varies.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", err
	}

	var dataList []struct {
		ID    string `json:"id"`
		Email string `json:"username"`
	}
	if err := json.Unmarshal(raw["data"], &dataList); err != nil {
		return "", fmt.Errorf("rancher users parse error: %w", err)
	}

	_ = listResp // used above to avoid unused import warning
	for _, u := range dataList {
		return u.ID, nil
	}
	return "", fmt.Errorf("rancher user not found for email: %s", email)
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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("rancher POST %s: HTTP %d: %s", path, resp.StatusCode, body)
	}
	return body, nil
}
