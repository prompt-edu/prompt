package keycloakRealmManager

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/Nerzal/gocloak/v14"
	"github.com/prompt-edu/prompt/servers/core/keycloakRealmManager/keycloakRealmDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

var (
	ErrUniversityLoginNotFound  = errors.New("no user with this university login exists")
	ErrAmbiguousUniversityLogin = errors.New("several users have this university login")
	ErrInvalidOrgRole           = errors.New("org role must be Admin or Member")
)

// maxUniversityLoginMatches bounds a user lookup by university login, which should find one user.
const maxUniversityLoginMatches = 10

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
		users, err := s.groupMembers(ctx, token.AccessToken, groupID)
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

// FindUserByUniversityLogin resolves a university login to a Keycloak user: through the
// university_login attribute that feeds the token claim, or, for users without it, through an equal
// username.
func (s *KeycloakRealmService) FindUserByUniversityLogin(ctx context.Context, universityLogin string) (keycloakRealmDTO.StaffMember, error) {
	token, err := s.LoginClient(ctx)
	if err != nil {
		return keycloakRealmDTO.StaffMember{}, err
	}

	matches, err := s.findUsers(ctx, token.AccessToken,
		gocloak.GetUsersParams{Q: gocloak.StringP(keycloakRealmDTO.UniversityLoginAttribute + ":" + universityLogin)},
		func(user keycloakRealmDTO.StaffMember) bool {
			return strings.EqualFold(user.UniversityLogin, universityLogin)
		})
	if err != nil {
		return keycloakRealmDTO.StaffMember{}, err
	}
	if len(matches) == 0 {
		matches, err = s.findUsers(ctx, token.AccessToken,
			gocloak.GetUsersParams{Username: &universityLogin, Exact: gocloak.BoolP(true)},
			func(user keycloakRealmDTO.StaffMember) bool { return strings.EqualFold(user.Username, universityLogin) })
		if err != nil {
			return keycloakRealmDTO.StaffMember{}, err
		}
	}

	switch len(matches) {
	case 0:
		return keycloakRealmDTO.StaffMember{}, ErrUniversityLoginNotFound
	case 1:
		return matches[0], nil
	default:
		return keycloakRealmDTO.StaffMember{}, ErrAmbiguousUniversityLogin
	}
}

// findUsers runs a user search and keeps the results that match exactly, since Keycloak may match
// a search value as a substring.
func (s *KeycloakRealmService) findUsers(ctx context.Context, accessToken string, params gocloak.GetUsersParams, matches func(keycloakRealmDTO.StaffMember) bool) ([]keycloakRealmDTO.StaffMember, error) {
	// The full representation carries the user attributes, which hold the university login.
	params.BriefRepresentation = gocloak.BoolP(false)
	params.Max = gocloak.IntP(maxUniversityLoginMatches)
	users, err := s.client.GetUsers(ctx, accessToken, s.Realm, params)
	if err != nil {
		return nil, fmt.Errorf("failed to search keycloak users: %w", err)
	}

	var result []keycloakRealmDTO.StaffMember
	for _, user := range users {
		member := keycloakRealmDTO.GetStaffMemberFromKeycloakUser(user)
		if member.KeycloakUserID != "" && matches(member) {
			result = append(result, member)
		}
	}
	return result, nil
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

// groupMembers returns every user in the group, page by page.
func (s *KeycloakRealmService) groupMembers(ctx context.Context, accessToken, groupID string) ([]*gocloak.User, error) {
	var users []*gocloak.User
	for first := 0; ; first += groupMembersPageSize {
		page, err := s.client.GetGroupMembers(ctx, accessToken, s.Realm, groupID, gocloak.GetGroupsParams{
			First: gocloak.IntP(first),
			Max:   gocloak.IntP(groupMembersPageSize),
			// The full representation carries the user attributes, which hold the university login.
			BriefRepresentation: gocloak.BoolP(false),
		})
		if err != nil {
			return nil, err
		}
		users = append(users, page...)
		if len(page) < groupMembersPageSize {
			return users, nil
		}
	}
}
