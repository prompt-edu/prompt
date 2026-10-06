package keycloakRealmManager

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/Nerzal/gocloak/v14"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
)

// CreateOrgGroupsAndRoles provisions the Keycloak artifacts of an org: the group
// /Orgs/<slug> with its Admin and Member subgroups, and the client roles
// org-<slug>-Admin and org-<slug>-Member mapped onto those subgroups. Every step
// reuses what already exists, so a retry after a partial failure completes the
// setup instead of failing on a leftover group. No users are added.
func (s *KeycloakRealmService) CreateOrgGroupsAndRoles(ctx context.Context, slug string) error {
	token, err := s.LoginClient(ctx)
	if err != nil {
		return err
	}

	orgsGroupID, err := s.getOrCreateGroupAtPath(ctx, token.AccessToken, ORGS_GROUP_NAME, "")
	if err != nil {
		return err
	}

	orgGroupPath := ORGS_GROUP_NAME + "/" + slug
	orgGroupID, err := s.getOrCreateGroupAtPath(ctx, token.AccessToken, orgGroupPath, orgsGroupID)
	if err != nil {
		return err
	}

	for _, orgRole := range []string{permissionValidation.OrgAdmin, permissionValidation.OrgMember} {
		role, err := s.GetOrCreateRealmRole(ctx, token.AccessToken, permissionValidation.OrgRoleName(slug, orgRole))
		if err != nil {
			return fmt.Errorf("failed to get or create the %s role: %w", orgRole, err)
		}

		subGroupID, err := s.getOrCreateGroupAtPath(ctx, token.AccessToken, orgGroupPath+"/"+orgRole, orgGroupID)
		if err != nil {
			return err
		}

		if err := s.AddRoleToGroup(ctx, token.AccessToken, subGroupID, role); err != nil {
			return fmt.Errorf("failed to map the %s role to its group: %w", orgRole, err)
		}
	}
	return nil
}

// getOrCreateGroupAtPath returns the ID of the group at groupPath, creating it under
// parentGroupID, or as a top-level group when parentGroupID is empty, if it does not
// exist. A creation that loses the race against a concurrent one, such as two first
// orgs both creating /Orgs, picks up the group that won.
func (s *KeycloakRealmService) getOrCreateGroupAtPath(ctx context.Context, accessToken, groupPath, parentGroupID string) (string, error) {
	groupID, found, err := s.findGroupAtPath(ctx, accessToken, groupPath)
	if err != nil || found {
		return groupID, err
	}

	groupName := path.Base(groupPath)
	newGroup := gocloak.Group{Name: &groupName}
	if parentGroupID == "" {
		groupID, err = s.client.CreateGroup(ctx, accessToken, s.Realm, newGroup)
	} else {
		groupID, err = s.client.CreateChildGroup(ctx, accessToken, s.Realm, parentGroupID, newGroup)
	}
	if hasStatus(err, http.StatusConflict) {
		if existingID, found, findErr := s.findGroupAtPath(ctx, accessToken, groupPath); findErr == nil && found {
			return existingID, nil
		}
	}
	if err != nil {
		return "", fmt.Errorf("failed to create group at path %s: %w", groupPath, err)
	}
	return groupID, nil
}

// findGroupAtPath returns the ID of the group at groupPath (without a leading slash, see
// GetGroupByPath), or found=false if there is none. The lookup goes by path because a
// name search, as in GetOrCreatePromptGroup, also matches subgroups elsewhere in the realm.
func (s *KeycloakRealmService) findGroupAtPath(ctx context.Context, accessToken, groupPath string) (groupID string, found bool, err error) {
	group, err := s.client.GetGroupByPath(ctx, accessToken, s.Realm, groupPath)
	if hasStatus(err, http.StatusNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("failed to get group at path %s: %w", groupPath, err)
	}
	if group.ID == nil || group.Name == nil || *group.Name != path.Base(groupPath) {
		return "", false, fmt.Errorf("unexpected group at path %s", groupPath)
	}
	return *group.ID, true, nil
}

// hasStatus reports whether a gocloak call failed with the given HTTP status. gocloak sets
// the status code on every HTTP error, so the code is matched, never a "404" substring: a
// transport error embeds the request URL, which can contain those digits in a slug, the
// Keycloak host, or the client ID.
func hasStatus(err error, status int) bool {
	var apiErr *gocloak.APIError
	return errors.As(err, &apiErr) && apiErr.Code == status
}
