package keycloakRealmManager

import (
	"context"
	"fmt"

	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
)

// CreateCourseGroupsAndRoles creates the course's groups and client roles, maps each staff role to
// its group, and makes the requester the course's only Lecturer.
//
// The course service lets no two courses share an identifier, so a group or role already named
// after it is left over from a creation whose commit failed, or from a course deleted before
// deletions cleaned up Keycloak. They are deleted first, which drops every grant of them to a user
// or another group, so nobody keeps access to the new course through them and running it again
// ends in the same state.
func (s *KeycloakRealmService) CreateCourseGroupsAndRoles(ctx context.Context, courseName, iterationName, userID string) error {
	courseGroupName := permissionValidation.CourseIdentifier(iterationName, courseName)
	if err := s.deleteCourseGroupsAndRoles(ctx, courseGroupName); err != nil {
		return fmt.Errorf("failed to delete the leftover groups and roles: %w", err)
	}

	token, err := s.LoginClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to log in to Keycloak: %w", err)
	}

	promptGroupID, err := s.getOrCreateGroupAtPath(ctx, token.AccessToken, TOP_LEVEL_GROUP_NAME, "")
	if err != nil {
		return err
	}

	courseGroupID, err := s.CreateChildGroup(ctx, token.AccessToken, courseGroupName, promptGroupID)
	if err != nil {
		return err
	}

	for _, staffRole := range []string{permissionValidation.CourseLecturer, permissionValidation.CourseEditor} {
		role, err := s.GetOrCreateRealmRole(ctx, token.AccessToken, permissionValidation.CourseRoleName(courseGroupName, staffRole))
		if err != nil {
			return fmt.Errorf("failed to create the %s role: %w", staffRole, err)
		}

		subGroupID, err := s.CreateChildGroup(ctx, token.AccessToken, staffRole, courseGroupID)
		if err != nil {
			return err
		}

		if err := s.AddRoleToGroup(ctx, token.AccessToken, subGroupID, role); err != nil {
			return fmt.Errorf("failed to map the %s role to its group: %w", staffRole, err)
		}

		if staffRole == permissionValidation.CourseLecturer {
			if err := s.client.AddUserToGroup(ctx, token.AccessToken, s.Realm, userID, subGroupID); err != nil {
				return fmt.Errorf("failed to add the requester to the Lecturer group: %w", err)
			}
		}
	}
	return nil
}
