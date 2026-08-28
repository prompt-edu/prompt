// Package gitlab implements the Provider interface for GitLab group and project
// management. It uses the GitLab REST API v4 with a personal access token.
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

const (
	resourceTypeGroup   = "group"
	resourceTypeProject = "project"

	// extraKeyParentGroupTemplate names the team subgroup a project is created in.
	extraKeyParentGroupTemplate = "parent_group_template"
	// extraKeyVisibility overrides the visibility of a created project.
	extraKeyVisibility = "visibility"
	// extraKeyInitializeWithReadme controls whether a created project gets a first commit.
	extraKeyInitializeWithReadme = "initialize_with_readme"

	// defaultProjectVisibility is set explicitly rather than relying on the instance's
	// default_project_visibility setting, which an administrator may have set to public.
	defaultProjectVisibility = "private"
)

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

func (p *Provider) SupportedResourceTypes() []string { return []string{"group", "project"} }

func (p *Provider) TemplatedExtraConfigKeys() []string { return []string{"parent_group_template"} }

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

// CreateResource creates the requested GitLab resource and adds members.
// It is idempotent: an existing resource of the same path is reused.
func (p *Provider) CreateResource(ctx context.Context, input provider.CreateResourceInput) (*provider.Resource, error) {
	switch input.ResourceType {
	case "", resourceTypeGroup:
		return p.createGroup(ctx, input)
	case resourceTypeProject:
		return p.createProject(ctx, input)
	default:
		// Never fall through to group creation: a mis-typed resource type would create a
		// group named after the project template and report it as a success.
		return nil, fmt.Errorf("gitlab: unsupported resource type %q (supported: %s, %s)",
			input.ResourceType, resourceTypeGroup, resourceTypeProject)
	}
}

func (p *Provider) createGroup(ctx context.Context, input provider.CreateResourceInput) (*provider.Resource, error) {
	name := sanitizeName(input.Name)
	slug := toSlug(name)
	if slug == "" {
		return nil, fmt.Errorf("gitlab: resource name %q sanitizes to an empty group path", input.Name)
	}

	parentID, hasParent, err := p.parentGroupID()
	if err != nil {
		return nil, err
	}

	groupID, groupURL, err := p.findOrCreateGroup(ctx, name, slug, parentID, hasParent)
	if err != nil {
		return nil, err
	}

	return &provider.Resource{
		ExternalID:  strconv.Itoa(groupID),
		ExternalURL: groupURL,
		Warnings:    p.addMembers(ctx, groupID, input),
	}, nil
}

// createProject creates the team's subgroup and a project inside it.
//
// The subgroup is an explicit side effect of a project resource rather than a tracked
// dependency: instances carry no ordering, so a project cannot wait for a separately
// configured group row. Both paths adopt by exact path under the same parent, so a
// project config and a group config for the same team converge on one subgroup.
//
// Members are granted access on the subgroup only. GitLab group members inherit access
// to every project in the group, so per-project invitations would add nothing.
func (p *Provider) createProject(ctx context.Context, input provider.CreateResourceInput) (*provider.Resource, error) {
	projectName := sanitizeName(input.Name)
	projectSlug := toSlug(projectName)
	if projectSlug == "" {
		return nil, fmt.Errorf("gitlab: resource name %q sanitizes to an empty project path", input.Name)
	}

	parentID, hasParent, err := p.parentGroupID()
	if err != nil {
		return nil, err
	}
	// Without a configured root group the project would land in the token owner's
	// personal namespace, or create a top-level group most instances forbid.
	if !hasParent {
		return nil, fmt.Errorf("gitlab: a project resource requires the provider's parent_group_id to be set")
	}

	groupTemplate, _ := input.ExtraConfig[extraKeyParentGroupTemplate].(string)
	groupName := sanitizeName(groupTemplate)
	groupSlug := toSlug(groupName)
	if groupSlug == "" {
		return nil, fmt.Errorf("gitlab: a project resource requires extra config %q, naming the team subgroup to create the project in",
			extraKeyParentGroupTemplate)
	}

	groupID, _, err := p.findOrCreateGroup(ctx, groupName, groupSlug, parentID, hasParent)
	if err != nil {
		return nil, err
	}

	// Everything below happens after the subgroup exists, so a failure is reported as a
	// warning and the instance stays retryable rather than losing the group.
	warnings := p.addMembers(ctx, groupID, input)

	projectID, projectURL, err := p.findOrCreateProject(ctx, groupID, groupSlug, projectName, projectSlug, input)
	if err != nil {
		return nil, err
	}

	return &provider.Resource{
		ExternalID:  strconv.Itoa(projectID),
		ExternalURL: projectURL,
		Warnings:    warnings,
	}, nil
}

// addMembers grants every member access to a group, reporting per-member failures as
// warnings so the instance becomes partial rather than failing outright.
func (p *Provider) addMembers(ctx context.Context, groupID int, input provider.CreateResourceInput) []string {
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
	return warnings
}

// findOrCreateGroup returns the group ID and URL for a group, creating it if necessary.
//
// With a parent configured we search that parent's subgroups and compare the path
// exactly. Matching on a full_path suffix across all visible groups would also match
// another course's group of the same name and add students to it.
func (p *Provider) findOrCreateGroup(ctx context.Context, name, slug string, parentID int, hasParent bool) (int, string, error) {
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

// findOrCreateProject returns the project ID and web URL, creating the project if it
// does not exist yet.
//
// Existence is decided by an exact lookup on the namespaced path rather than by parsing
// the create error. GitLab's duplicate-path response is the least specified part of the
// surface: create answers 400 with a field-keyed message map, while the fork endpoint
// answers 409 with a flat array of sentences. The create error is only used as a signal
// to re-run the exact lookup.
func (p *Provider) findOrCreateProject(ctx context.Context, groupID int, groupSlug, projectName, projectSlug string, input provider.CreateResourceInput) (int, string, error) {
	fullPath := groupSlug + "/" + projectSlug

	if id, webURL, found, err := p.findProject(ctx, fullPath); err != nil {
		return 0, "", err
	} else if found {
		return id, webURL, nil
	}

	payload := map[string]interface{}{
		"name":         projectName,
		"path":         projectSlug,
		"namespace_id": groupID,
		"visibility":   projectVisibility(input.ExtraConfig),
	}
	if initializeWithReadme(input.ExtraConfig) {
		payload["initialize_with_readme"] = true
	}

	body, err := p.post(ctx, "/api/v4/projects", payload)
	if err != nil {
		// A concurrent run may have taken the path. Re-resolve rather than reporting a
		// failure the retry would only repeat.
		if id, webURL, found, findErr := p.findProject(ctx, fullPath); findErr == nil && found {
			return id, webURL, nil
		}
		return 0, "", err
	}

	var created struct {
		ID     int    `json:"id"`
		WebURL string `json:"web_url"`
	}
	if err := json.Unmarshal(body, &created); err != nil {
		return 0, "", err
	}
	return created.ID, created.WebURL, nil
}

// findProject resolves a project by its namespaced path. The path has to be URL-encoded
// as a single segment, so every "/" becomes "%2F".
func (p *Provider) findProject(ctx context.Context, fullPath string) (int, string, bool, error) {
	body, err := p.get(ctx, "/api/v4/projects/"+url.PathEscape(fullPath))
	if err != nil {
		if strings.Contains(err.Error(), "HTTP 404") {
			return 0, "", false, nil
		}
		return 0, "", false, err
	}

	var found struct {
		ID     int    `json:"id"`
		WebURL string `json:"web_url"`
	}
	if err := json.Unmarshal(body, &found); err != nil {
		return 0, "", false, err
	}
	return found.ID, found.WebURL, true, nil
}

func projectVisibility(extra map[string]interface{}) string {
	if value, ok := extra[extraKeyVisibility].(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return defaultProjectVisibility
}

// initializeWithReadme defaults to true so the repository is usable straight away;
// GitLab's own default leaves it without a branch.
func initializeWithReadme(extra map[string]interface{}) bool {
	if value, ok := extra[extraKeyInitializeWithReadme].(bool); ok {
		return value
	}
	return true
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
		return nil, provider.HTTPError("gitlab", http.MethodGet, path, resp.StatusCode, body)
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
		return nil, provider.HTTPError("gitlab", http.MethodPost, path, resp.StatusCode, body)
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
