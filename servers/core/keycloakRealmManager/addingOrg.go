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

func hasStatus(err error, status int) bool {
	var apiErr *gocloak.APIError
	return errors.As(err, &apiErr) && apiErr.Code == status
}
