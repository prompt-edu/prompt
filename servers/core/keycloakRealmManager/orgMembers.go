package keycloakRealmManager

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/prompt-edu/prompt/servers/core/keycloakRealmManager/keycloakRealmDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

var ErrInvalidOrgRole = errors.New("org role must be Admin or Member")

var orgRoles = []string{permissionValidation.OrgAdmin, permissionValidation.OrgMember}

func orgRoleGroupPath(slug, orgRole string) string {
	return ORGS_GROUP_NAME + "/" + slug + "/" + orgRole
}

// GetOrgMembers lists the users in the org's Admin and Member groups, Admins first. A user in both
// groups is listed once, as an Admin. A missing group counts as empty.
func (s *KeycloakRealmService) GetOrgMembers(ctx context.Context, slug string) ([]keycloakRealmDTO.OrgMember, error) {
	token, err := s.LoginClient(ctx)
	if err != nil {
		return nil, err
	}

	members := []keycloakRealmDTO.OrgMember{}
	listed := map[string]bool{}
	for _, orgRole := range orgRoles {
		groupID, found, err := s.findGroupAtPath(ctx, token.AccessToken, orgRoleGroupPath(slug, orgRole))
		if err != nil {
			return nil, err
		}
		if !found {
			log.Warnf("org %s %s group not found in Keycloak; treating as empty", slug, orgRole)
			continue
		}
		users, err := s.GetGroupMembers(ctx, token.AccessToken, groupID)
		if err != nil {
			return nil, fmt.Errorf("failed to get the members of the %s group: %w", orgRole, err)
		}
		for _, user := range users {
			if user.ID == nil || listed[*user.ID] {
				continue
			}
			listed[*user.ID] = true
			members = append(members, keycloakRealmDTO.OrgMember{
				StaffMember: keycloakRealmDTO.GetStaffMemberFromKeycloakUser(user),
				Role:        orgRole,
			})
		}
	}
	return members, nil
}

// SetOrgMemberRole gives the user exactly the given org role. It adds them to that role's group
// before taking them out of the other one, so a role change never drops them from the org on the
// way. Both steps are idempotent.
func (s *KeycloakRealmService) SetOrgMemberRole(ctx context.Context, slug, userID, orgRole string) error {
	if !slices.Contains(orgRoles, orgRole) {
		return ErrInvalidOrgRole
	}

	token, err := s.LoginClient(ctx)
	if err != nil {
		return err
	}

	groupIDs := map[string]string{}
	for _, role := range orgRoles {
		groupID, found, err := s.findGroupAtPath(ctx, token.AccessToken, orgRoleGroupPath(slug, role))
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("the %s group of org %s is missing in Keycloak", role, slug)
		}
		groupIDs[role] = groupID
	}

	if err := s.client.AddUserToGroup(ctx, token.AccessToken, s.Realm, userID, groupIDs[orgRole]); err != nil {
		if hasStatus(err, http.StatusNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to add the user to the %s group: %w", orgRole, err)
	}
	for _, role := range orgRoles {
		if role == orgRole {
			continue
		}
		if err := s.client.DeleteUserFromGroup(ctx, token.AccessToken, s.Realm, userID, groupIDs[role]); err != nil {
			return fmt.Errorf("failed to remove the user from the %s group: %w", role, err)
		}
	}
	return nil
}

// RemoveOrgMember takes the user out of both org role groups. Removing a user who is not a member
// changes nothing.
func (s *KeycloakRealmService) RemoveOrgMember(ctx context.Context, slug, userID string) error {
	token, err := s.LoginClient(ctx)
	if err != nil {
		return err
	}

	for _, orgRole := range orgRoles {
		groupID, found, err := s.findGroupAtPath(ctx, token.AccessToken, orgRoleGroupPath(slug, orgRole))
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		if err := s.client.DeleteUserFromGroup(ctx, token.AccessToken, s.Realm, userID, groupID); err != nil {
			if hasStatus(err, http.StatusNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("failed to remove the user from the %s group: %w", orgRole, err)
		}
	}
	return nil
}
