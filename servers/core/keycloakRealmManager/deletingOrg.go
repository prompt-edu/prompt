package keycloakRealmManager

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Nerzal/gocloak/v14"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

const groupMembersPageSize = 100

// DeleteOrgGroupsAndRoles deletes the org's roles and groups, and returns the Keycloak IDs of the
// users who were in its Admin and Member groups, keyed by role, so that a deletion the database
// rolls back can put them back. It changes nothing when the members cannot be read.
func (s *KeycloakRealmService) DeleteOrgGroupsAndRoles(ctx context.Context, slug string) (map[string][]string, error) {
	token, err := s.LoginClient(ctx)
	if err != nil {
		return nil, err
	}

	groupPath := ORGS_GROUP_NAME + "/" + slug
	groupID, found, err := s.findGroupAtPath(ctx, token.AccessToken, groupPath)
	if err != nil {
		return nil, err
	}

	members, err := s.orgMemberIDs(ctx, token.AccessToken, groupPath)
	if err != nil {
		return nil, err
	}

	for _, orgRole := range []string{permissionValidation.OrgAdmin, permissionValidation.OrgMember} {
		roleName := permissionValidation.OrgRoleName(slug, orgRole)
		if err := s.client.DeleteClientRole(ctx, token.AccessToken, s.Realm, s.idOfClient, roleName); err != nil && !hasStatus(err, http.StatusNotFound) {
			return members, fmt.Errorf("failed to delete client role %s: %w", roleName, err)
		}
	}

	if !found {
		log.Warnf("org %s group not found in Keycloak; skipping group deletion", slug)
		return members, nil
	}
	if err := s.client.DeleteGroup(ctx, token.AccessToken, s.Realm, groupID); err != nil && !hasStatus(err, http.StatusNotFound) {
		return members, fmt.Errorf("failed to delete org group: %w", err)
	}
	return members, nil
}

func (s *KeycloakRealmService) orgMemberIDs(ctx context.Context, accessToken, orgGroupPath string) (map[string][]string, error) {
	members := map[string][]string{}
	for _, orgRole := range []string{permissionValidation.OrgAdmin, permissionValidation.OrgMember} {
		groupID, found, err := s.findGroupAtPath(ctx, accessToken, orgGroupPath+"/"+orgRole)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		userIDs, err := s.groupMemberIDs(ctx, accessToken, groupID)
		if err != nil {
			return nil, fmt.Errorf("failed to get the members of the %s group: %w", orgRole, err)
		}
		members[orgRole] = userIDs
	}
	return members, nil
}

func (s *KeycloakRealmService) groupMemberIDs(ctx context.Context, accessToken, groupID string) ([]string, error) {
	var userIDs []string
	for first := 0; ; first += groupMembersPageSize {
		pageSize := groupMembersPageSize
		page, err := s.client.GetGroupMembers(ctx, accessToken, s.Realm, groupID, gocloak.GetGroupsParams{First: &first, Max: &pageSize})
		if err != nil {
			return nil, err
		}
		for _, user := range page {
			if user.ID != nil {
				userIDs = append(userIDs, *user.ID)
			}
		}
		if len(page) < groupMembersPageSize {
			return userIDs, nil
		}
	}
}

// AddOrgMembers puts users back into the org's role groups, keyed by role as DeleteOrgGroupsAndRoles
// returns them. It tries every user and reports each one it could not add.
func (s *KeycloakRealmService) AddOrgMembers(ctx context.Context, slug string, members map[string][]string) error {
	if len(members[permissionValidation.OrgAdmin])+len(members[permissionValidation.OrgMember]) == 0 {
		return nil
	}
	token, err := s.LoginClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to log in to re-add the org's members %v: %w", members, err)
	}

	var errs []error
	for _, orgRole := range []string{permissionValidation.OrgAdmin, permissionValidation.OrgMember} {
		userIDs := members[orgRole]
		if len(userIDs) == 0 {
			continue
		}
		groupPath := ORGS_GROUP_NAME + "/" + slug + "/" + orgRole
		groupID, found, err := s.findGroupAtPath(ctx, token.AccessToken, groupPath)
		if err == nil && !found {
			err = fmt.Errorf("group %s not found", groupPath)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to re-add the %s members %v: %w", orgRole, userIDs, err))
			continue
		}
		for _, userID := range userIDs {
			if err := s.client.AddUserToGroup(ctx, token.AccessToken, s.Realm, userID, groupID); err != nil {
				errs = append(errs, fmt.Errorf("failed to re-add user %s to the %s group: %w", userID, orgRole, err))
			}
		}
	}
	return errors.Join(errs...)
}

func (s *KeycloakRealmService) OrgGroupsOrRolesExist(ctx context.Context, slug string) (bool, error) {
	token, err := s.LoginClient(ctx)
	if err != nil {
		return false, err
	}

	_, found, err := s.findGroupAtPath(ctx, token.AccessToken, ORGS_GROUP_NAME+"/"+slug)
	if err != nil || found {
		return found, err
	}

	for _, orgRole := range []string{permissionValidation.OrgAdmin, permissionValidation.OrgMember} {
		roleName := permissionValidation.OrgRoleName(slug, orgRole)
		_, err := s.client.GetClientRole(ctx, token.AccessToken, s.Realm, s.idOfClient, roleName)
		if err == nil {
			return true, nil
		}
		if !hasStatus(err, http.StatusNotFound) {
			return false, fmt.Errorf("failed to get client role %s: %w", roleName, err)
		}
	}
	return false, nil
}
