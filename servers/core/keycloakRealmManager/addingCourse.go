package keycloakRealmManager

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
)

// CreateCourseGroupsAndRoles gets or creates the course's groups, client roles, and role mappings,
// and makes the requester the course's only Lecturer, so running it again changes nothing.
//
// The course service lets no two courses share an identifier, so a course group that already exists
// is left over from a creation whose commit failed, or from a course deleted before deletions
// cleaned up Keycloak. Its Lecturer and Editor members are removed and its custom groups deleted,
// so nobody keeps access to the new course through it.
func (s *KeycloakRealmService) CreateCourseGroupsAndRoles(ctx context.Context, courseName, iterationName, userID string) error {
	token, err := s.LoginClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to log in to Keycloak: %w", err)
	}

	promptGroupID, err := s.getOrCreateGroupAtPath(ctx, token.AccessToken, TOP_LEVEL_GROUP_NAME, "")
	if err != nil {
		return err
	}

	courseGroupName := permissionValidation.CourseIdentifier(iterationName, courseName)
	courseGroupPath := TOP_LEVEL_GROUP_NAME + "/" + courseGroupName
	courseGroupID, err := s.getOrCreateGroupAtPath(ctx, token.AccessToken, courseGroupPath, promptGroupID)
	if err != nil {
		return err
	}

	var lecturerGroupID string
	for _, staffRole := range []string{permissionValidation.CourseLecturer, permissionValidation.CourseEditor} {
		role, err := s.GetOrCreateRealmRole(ctx, token.AccessToken, permissionValidation.CourseRoleName(courseGroupName, staffRole))
		if err != nil {
			return fmt.Errorf("failed to get or create the %s role: %w", staffRole, err)
		}

		subGroupID, err := s.getOrCreateGroupAtPath(ctx, token.AccessToken, courseGroupPath+"/"+staffRole, courseGroupID)
		if err != nil {
			return err
		}

		if err := s.AddRoleToGroup(ctx, token.AccessToken, subGroupID, role); err != nil {
			return fmt.Errorf("failed to map the %s role to its group: %w", staffRole, err)
		}

		keptUserID := ""
		if staffRole == permissionValidation.CourseLecturer {
			lecturerGroupID = subGroupID
			keptUserID = userID
		}
		if err := s.removeGroupMembersExcept(ctx, token.AccessToken, subGroupID, keptUserID); err != nil {
			return fmt.Errorf("failed to empty the %s group: %w", staffRole, err)
		}
	}

	if err := s.deleteGroupAtPath(ctx, token.AccessToken, courseGroupPath+"/"+CUSTOM_GROUPS_NAME); err != nil {
		return err
	}

	if err := s.client.AddUserToGroup(ctx, token.AccessToken, s.Realm, userID, lecturerGroupID); err != nil {
		return fmt.Errorf("failed to add the requester to the Lecturer group: %w", err)
	}
	return nil
}

// removeGroupMembersExcept removes every member of the group but keptUserID.
func (s *KeycloakRealmService) removeGroupMembersExcept(ctx context.Context, accessToken, groupID, keptUserID string) error {
	members, err := s.GetGroupMembers(ctx, accessToken, groupID)
	if err != nil {
		return fmt.Errorf("failed to get the group members: %w", err)
	}
	for _, member := range members {
		if member.ID == nil || *member.ID == keptUserID {
			continue
		}
		if err := s.client.DeleteUserFromGroup(ctx, accessToken, s.Realm, *member.ID, groupID); err != nil {
			return fmt.Errorf("failed to remove user %s: %w", *member.ID, err)
		}
	}
	return nil
}

// deleteGroupAtPath deletes the group at groupPath together with its subgroups and memberships. A
// group that does not exist counts as deleted.
func (s *KeycloakRealmService) deleteGroupAtPath(ctx context.Context, accessToken, groupPath string) error {
	groupID, found, err := s.findGroupAtPath(ctx, accessToken, groupPath)
	if err != nil || !found {
		return err
	}
	if err := s.client.DeleteGroup(ctx, accessToken, s.Realm, groupID); err != nil && !hasStatus(err, http.StatusNotFound) {
		return fmt.Errorf("failed to delete group at path %s: %w", groupPath, err)
	}
	return nil
}
