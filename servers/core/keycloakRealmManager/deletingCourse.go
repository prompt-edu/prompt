package keycloakRealmManager

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Nerzal/gocloak/v14"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

// DeleteCourseGroupsAndRoles removes the Keycloak artifacts created by
// CreateCourseGroupsAndRoles: the course group (Keycloak cascades its
// Lecturer/Editor/CustomGroups subgroups) and the client roles of the course.
// It must run before the course row is removed from the database, since the
// group name is derived from it.
func (s *KeycloakRealmService) DeleteCourseGroupsAndRoles(ctx context.Context, courseID uuid.UUID) error {
	courseGroupName, err := s.GetCourseGroupName(ctx, courseID)
	if err != nil {
		return fmt.Errorf("failed to get course group name: %w", err)
	}
	return s.deleteCourseGroupsAndRoles(ctx, courseGroupName)
}

// DeleteCourseGroupsAndRolesByName does what DeleteCourseGroupsAndRoles does for a course that is
// identified by its name and semester instead, such as one whose creating transaction has not
// committed yet and whose row other connections therefore cannot read.
func (s *KeycloakRealmService) DeleteCourseGroupsAndRolesByName(ctx context.Context, courseName, iterationName string) error {
	return s.deleteCourseGroupsAndRoles(ctx, permissionValidation.CourseIdentifier(iterationName, courseName))
}

const childGroupsPageSize = 100

// deleteCourseGroupsAndRoles deletes the roles of the course's custom groups, then the course group,
// then the Lecturer and Editor roles. Anything already gone counts as deleted, so a retry is safe.
//
// A custom group role is found through the custom group it belongs to, never by its name alone:
// "ws24-ios-cg-team-Lecturer" can be a custom group role of the course "ios" as well as the Lecturer
// role of a course "ios-cg-team" created before course names were refused a "-". Such a role can
// even be both at once, so a role that a group of another course still maps is kept. Until the
// group is deleted, its Lecturer members keep the Lecturer role, so they can retry a deletion that
// failed.
func (s *KeycloakRealmService) deleteCourseGroupsAndRoles(ctx context.Context, courseGroupName string) error {
	if !permissionValidation.IsAddressableInKeycloak(courseGroupName) {
		log.Warnf("course %s cannot be addressed in Keycloak; skipping the deletion of its groups and roles", courseGroupName)
		return nil
	}

	token, err := s.LoginClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to log in to Keycloak: %w", err)
	}

	courseGroupPath := TOP_LEVEL_GROUP_NAME + "/" + courseGroupName
	groupID, found, err := s.findGroupAtPath(ctx, token.AccessToken, courseGroupPath)
	if err != nil {
		return err
	}
	if found {
		customRoleNames, err := s.customGroupRoleNames(ctx, token.AccessToken, courseGroupPath, courseGroupName)
		if err != nil {
			return err
		}
		for _, roleName := range customRoleNames {
			if err := s.deleteCourseRole(ctx, token.AccessToken, courseGroupPath, roleName); err != nil {
				return err
			}
		}
		if err := s.client.DeleteGroup(ctx, token.AccessToken, s.Realm, groupID); err != nil && !hasStatus(err, http.StatusNotFound) {
			return fmt.Errorf("failed to delete course group: %w", err)
		}
	} else {
		log.Debugf("course %s group not found in Keycloak; skipping group deletion", courseGroupName)
	}

	for _, staffRole := range []string{permissionValidation.CourseEditor, permissionValidation.CourseLecturer} {
		if err := s.deleteCourseRole(ctx, token.AccessToken, courseGroupPath, permissionValidation.CourseRoleName(courseGroupName, staffRole)); err != nil {
			return err
		}
	}
	return nil
}

func (s *KeycloakRealmService) customGroupRoleNames(ctx context.Context, accessToken, courseGroupPath, courseGroupName string) ([]string, error) {
	customGroupsID, found, err := s.findGroupAtPath(ctx, accessToken, courseGroupPath+"/"+CUSTOM_GROUPS_NAME)
	if err != nil || !found {
		return nil, err
	}

	var roleNames []string
	for first := 0; ; first += childGroupsPageSize {
		page, err := s.client.GetChildGroups(ctx, accessToken, s.Realm, customGroupsID, gocloak.GetChildGroupsParams{
			First: gocloak.IntP(first),
			Max:   gocloak.IntP(childGroupsPageSize),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list the custom groups: %w", err)
		}
		for _, customGroup := range page {
			if customGroup.Name == nil {
				continue
			}
			roleName := customGroupRoleName(courseGroupName, *customGroup.Name)
			if !permissionValidation.IsAddressableInKeycloak(roleName) {
				log.Warnf("custom group role %s cannot be addressed in Keycloak; skipping its deletion", roleName)
				continue
			}
			roleNames = append(roleNames, roleName)
		}
		if len(page) < childGroupsPageSize {
			return roleNames, nil
		}
	}
}

func (s *KeycloakRealmService) deleteCourseRole(ctx context.Context, accessToken, courseGroupPath, roleName string) error {
	otherCourseGroup, err := s.otherCourseGroupOfRole(ctx, accessToken, courseGroupPath, roleName)
	if err != nil {
		return err
	}
	if otherCourseGroup != "" {
		log.Warnf("client role %s is mapped to the group %s of another course; keeping it", roleName, otherCourseGroup)
		return nil
	}
	if err := s.client.DeleteClientRole(ctx, accessToken, s.Realm, s.idOfClient, roleName); err != nil && !hasStatus(err, http.StatusNotFound) {
		return fmt.Errorf("failed to delete client role %s: %w", roleName, err)
	}
	return nil
}
