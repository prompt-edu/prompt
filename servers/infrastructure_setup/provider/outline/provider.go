// Package outline implements the Provider interface for Outline collection management.
// It uses the Outline JSON-RPC-style API (POST requests) with an API key.
package outline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"slices"
	"strings"

	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
)

const outlinePageSize = 100

// extraKeyGroupNameTemplate overrides the display name of the Outline group bound to the
// collection; it defaults to the collection name.
const extraKeyGroupNameTemplate = "group_name_template"

// ownershipMarkerPrefix tags a collection description so a later run can tell a
// collection this course phase created from an unrelated one of the same name.
const ownershipMarkerPrefix = "prompt-key: "

// validPermissions is Outline's CollectionPermission enum.
var validPermissions = map[string]bool{"read": true, "read_write": true, "admin": true}

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

func (p *Provider) SupportedResourceTypes() []string { return []string{"collection"} }

func (p *Provider) TemplatedExtraConfigKeys() []string { return []string{"group_name_template"} }

// The group name template is optional: it defaults to the collection name.
func (p *Provider) RequiredExtraConfigKeys(string) []string { return nil }

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

// CreateResource creates an Outline collection, provisions a group per distinct
// permission, puts the members in those groups and grants each group access to the
// collection.
//
// Note on what this does and does not do. Outline has no OIDC group synchronisation:
// its OIDC plugin reads no groups claim, so a collection's visibility can never be
// decided from a token. Keycloak authenticates the user; Outline authorises against its
// own group membership. PROMPT puts the same snapshot of team members into both the
// Keycloak group and the Outline group. Nothing is synchronised afterwards, and no
// membership is ever removed.
func (p *Provider) CreateResource(ctx context.Context, input provider.CreateResourceInput) (*provider.Resource, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, fmt.Errorf("outline: resource name is empty")
	}

	collectionID, collectionURL, err := p.resolveCollection(ctx, input)
	if err != nil {
		return nil, err
	}

	// Everything below runs after the collection exists, so a failure is a warning and
	// the instance stays partial with its external ID rather than failing and later
	// re-adopting the collection by name.
	buckets, warnings := permissionBuckets(input)
	warnings = append(warnings, p.bindGroups(ctx, collectionID, input, buckets)...)

	return &provider.Resource{ExternalID: collectionID, ExternalURL: collectionURL, Warnings: warnings}, nil
}

// permissionBuckets groups the members by the Outline permission their role maps to.
// A role with no mapping, or one mapping to a permission Outline does not accept, is
// reported instead of silently becoming read access.
func permissionBuckets(input provider.CreateResourceInput) (map[string][]provider.Member, []string) {
	buckets := map[string][]provider.Member{}
	var warnings []string

	for _, member := range input.Members {
		permission, ok := input.PermissionMapping[member.Role]
		if !ok {
			warnings = append(warnings, fmt.Sprintf("%s: no permission mapped for role %q", member.Email, member.Role))
			continue
		}
		normalized := strings.ToLower(strings.TrimSpace(permission))
		if !validPermissions[normalized] {
			warnings = append(warnings, fmt.Sprintf("%s: unknown permission %q (expected read, read_write or admin)", member.Email, permission))
			continue
		}
		buckets[normalized] = append(buckets[normalized], member)
	}
	return buckets, warnings
}

// bindGroups provisions one group per distinct permission and grants it collection
// access. One group can hold one permission, so a team of students at read plus a tutor
// at read_write becomes two groups and two bindings rather than a silent choice between
// them.
func (p *Provider) bindGroups(ctx context.Context, collectionID string, input provider.CreateResourceInput, buckets map[string][]provider.Member) []string {
	var warnings []string
	if len(buckets) == 0 {
		return warnings
	}

	baseName := strings.TrimSpace(input.Name)
	if template, ok := input.ExtraConfig[extraKeyGroupNameTemplate].(string); ok && strings.TrimSpace(template) != "" {
		baseName = strings.TrimSpace(template)
	}
	split := len(buckets) > 1

	emails := make([]string, 0, len(input.Members))
	for _, members := range buckets {
		for _, member := range members {
			emails = append(emails, member.Email)
		}
	}
	userIDs, err := p.lookupUsersByEmail(ctx, emails)
	if err != nil {
		return append(warnings, fmt.Sprintf("could not look up Outline users: %v", err))
	}

	for _, permission := range slices.Sorted(maps.Keys(buckets)) {
		// The key always carries the permission, even for a single group. Keying the
		// lone group on the bare stable key would orphan it as soon as a second
		// permission appeared and the key gained a suffix.
		externalID := input.StableKey + ":" + permission

		groupName := baseName
		if split {
			// A distinguishable display name, so Outline's group list stays readable
			// when one collection is served by several groups.
			groupName = fmt.Sprintf("%s (%s)", baseName, permission)
		}

		groupID, err := p.findOrCreateGroup(ctx, groupName, externalID)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("group %q: %v", groupName, err))
			continue
		}

		for _, member := range buckets[permission] {
			userID, ok := userIDs[strings.ToLower(member.Email)]
			if !ok {
				warnings = append(warnings, fmt.Sprintf("%s: no Outline user with this email", member.Email))
				continue
			}
			if err := p.addUserToGroup(ctx, groupID, userID); err != nil {
				warnings = append(warnings, fmt.Sprintf("%s: %v", member.Email, err))
			}
		}

		if err := p.addGroupToCollection(ctx, collectionID, groupID, permission); err != nil {
			warnings = append(warnings, fmt.Sprintf("group %q: %v", groupName, err))
		}
	}
	return warnings
}

// resolveCollection returns the collection for this instance, creating it when needed.
//
// A single listing pass looks for the ID already recorded on the instance and for a
// name match. The recorded ID wins: on a retry the collection is re-attached by ID and
// the name never decides anything. A name match is adopted only when the collection
// carries this instance's ownership marker, because Outline collections share one flat
// workspace namespace and adopting a stranger's collection would hand a team read access
// to unrelated documents.
func (p *Provider) resolveCollection(ctx context.Context, input provider.CreateResourceInput) (string, string, error) {
	type collection struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		URL         string `json:"url"`
		Description string `json:"description"`
	}

	var nameMatch *collection
	offset := 0
	for {
		var listResp struct {
			OK   bool         `json:"ok"`
			Data []collection `json:"data"`
		}
		if err := p.call(ctx, "collections.list", map[string]interface{}{
			"offset": offset,
			"limit":  outlinePageSize,
		}, &listResp); err != nil {
			return "", "", err
		}
		if !listResp.OK {
			return "", "", fmt.Errorf("outline collections.list returned ok=false")
		}

		for _, c := range listResp.Data {
			if input.ExistingExternalID != "" && c.ID == input.ExistingExternalID {
				return c.ID, c.URL, nil
			}
			if nameMatch == nil && strings.EqualFold(c.Name, input.Name) {
				match := c
				nameMatch = &match
			}
		}

		if len(listResp.Data) < outlinePageSize {
			break
		}
		offset += outlinePageSize
	}

	if nameMatch != nil {
		if !strings.Contains(nameMatch.Description, ownershipMarker(input.StableKey)) {
			return "", "", fmt.Errorf(
				"outline: a collection named %q already exists and was not created by this course phase; rename it or rename the resource template",
				input.Name)
		}
		return nameMatch.ID, nameMatch.URL, nil
	}

	return p.createCollection(ctx, input)
}

// createCollection creates a private collection carrying the ownership marker.
//
// permission is deliberately omitted. Outline treats a collection with no permission as
// private, requiring a membership to read; setting it to "read" would let every member of
// the workspace read the collection, which is the opposite of per-team visibility.
func (p *Provider) createCollection(ctx context.Context, input provider.CreateResourceInput) (string, string, error) {
	var createResp struct {
		OK   bool `json:"ok"`
		Data struct {
			Collection struct {
				ID  string `json:"id"`
				URL string `json:"url"`
			} `json:"collection"`
			ID  string `json:"id"`
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := p.call(ctx, "collections.create", map[string]interface{}{
		"name":        input.Name,
		"sharing":     false,
		"description": collectionDescription(input.StableKey),
	}, &createResp); err != nil {
		return "", "", err
	}
	if !createResp.OK {
		return "", "", fmt.Errorf("outline collections.create returned ok=false")
	}

	// Outline nests the created collection under data.collection; some versions return
	// it flat under data.
	if createResp.Data.Collection.ID != "" {
		return createResp.Data.Collection.ID, createResp.Data.Collection.URL, nil
	}
	if createResp.Data.ID == "" {
		return "", "", fmt.Errorf("outline collections.create returned no collection id")
	}
	return createResp.Data.ID, createResp.Data.URL, nil
}

func ownershipMarker(stableKey string) string { return ownershipMarkerPrefix + stableKey }

func collectionDescription(stableKey string) string {
	return "Provisioned by PROMPT. Access is granted through the Outline group bound to this collection.\n\n" +
		ownershipMarker(stableKey)
}

// findOrCreateGroup resolves the Outline group paired with this instance.
//
// The group is keyed on externalId, the field Outline provides for linking a group to an
// external source, rather than on its display name. A display name follows the lecturer's
// template and can change; the key must not.
func (p *Provider) findOrCreateGroup(ctx context.Context, name, externalID string) (string, error) {
	if id, found, err := p.findGroupByExternalID(ctx, externalID); err != nil {
		return "", err
	} else if found {
		return id, nil
	}

	var createResp struct {
		OK   bool `json:"ok"`
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := p.call(ctx, "groups.create", map[string]interface{}{
		"name":       name,
		"externalId": externalID,
	}, &createResp); err != nil {
		// Losing a race, or an Outline version that rejects a duplicate name, must not
		// fail an instance whose group already exists.
		if id, found, findErr := p.findGroupByExternalID(ctx, externalID); findErr == nil && found {
			return id, nil
		}
		return "", err
	}
	if !createResp.OK || createResp.Data.ID == "" {
		return "", fmt.Errorf("outline groups.create returned no group id")
	}
	return createResp.Data.ID, nil
}

// findGroupByExternalID looks a group up by its external key. groups.list nests its
// payload under data.groups, unlike users.list which returns a flat array.
func (p *Provider) findGroupByExternalID(ctx context.Context, externalID string) (string, bool, error) {
	var listResp struct {
		OK   bool `json:"ok"`
		Data struct {
			Groups []struct {
				ID         string `json:"id"`
				ExternalID string `json:"externalId"`
			} `json:"groups"`
		} `json:"data"`
	}
	if err := p.call(ctx, "groups.list", map[string]interface{}{
		"externalId": externalID,
		"limit":      outlinePageSize,
	}, &listResp); err != nil {
		return "", false, err
	}
	if !listResp.OK {
		return "", false, fmt.Errorf("outline groups.list returned ok=false")
	}
	for _, g := range listResp.Data.Groups {
		if g.ExternalID == externalID {
			return g.ID, true, nil
		}
	}
	return "", false, nil
}

func (p *Provider) addUserToGroup(ctx context.Context, groupID, userID string) error {
	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := p.call(ctx, "groups.add_user", map[string]interface{}{
		"id":     groupID,
		"userId": userID,
	}, &resp); err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("outline groups.add_user failed: %s", provider.UpstreamReason(resp.Error))
	}
	return nil
}

// addGroupToCollection grants a group access to the collection. permission is always
// sent: Outline defaults collections.add_group to read_write, which would hand students
// write access to their own collection by omission.
func (p *Provider) addGroupToCollection(ctx context.Context, collectionID, groupID, permission string) error {
	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := p.call(ctx, "collections.add_group", map[string]interface{}{
		"id":         collectionID,
		"groupId":    groupID,
		"permission": permission,
	}, &resp); err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("outline collections.add_group failed: %s", provider.UpstreamReason(resp.Error))
	}
	return nil
}

// lookupUsersByEmail resolves every email in one or two calls instead of paging the
// workspace once per member. The email filter is tried first; anything it leaves
// unresolved is filled in by a single listing pass, because the filter parameter has
// changed shape across Outline versions.
func (p *Provider) lookupUsersByEmail(ctx context.Context, emails []string) (map[string]string, error) {
	wanted := map[string]bool{}
	for _, email := range emails {
		wanted[strings.ToLower(email)] = true
	}
	found := map[string]string{}
	if len(wanted) == 0 {
		return found, nil
	}

	var filtered struct {
		OK   bool `json:"ok"`
		Data []struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"data"`
	}
	if err := p.call(ctx, "users.list", map[string]interface{}{
		"emails": slices.Sorted(maps.Keys(wanted)),
		"limit":  outlinePageSize,
	}, &filtered); err == nil && filtered.OK {
		for _, u := range filtered.Data {
			if key := strings.ToLower(u.Email); wanted[key] {
				found[key] = u.ID
			}
		}
	}
	if len(found) == len(wanted) {
		return found, nil
	}

	offset := 0
	for {
		var listResp struct {
			OK   bool `json:"ok"`
			Data []struct {
				ID    string `json:"id"`
				Email string `json:"email"`
			} `json:"data"`
		}
		if err := p.call(ctx, "users.list", map[string]interface{}{
			"offset": offset,
			"limit":  outlinePageSize,
		}, &listResp); err != nil {
			return found, err
		}
		// Without this check a workspace-wide permission or rate-limit failure returns
		// HTTP 200 with ok=false and an empty page, which would be reported as "this
		// student does not exist".
		if !listResp.OK {
			return found, fmt.Errorf("outline users.list returned ok=false")
		}

		for _, u := range listResp.Data {
			if key := strings.ToLower(u.Email); wanted[key] {
				found[key] = u.ID
			}
		}
		if len(found) == len(wanted) || len(listResp.Data) < outlinePageSize {
			return found, nil
		}
		offset += outlinePageSize
	}
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
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return provider.HTTPError("outline", http.MethodPost, method, resp.StatusCode, body)
	}

	return json.Unmarshal(body, result)
}
