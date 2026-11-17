package keycloakRealmManager

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Nerzal/gocloak/v14"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

// DeleteCourseGroupsAndRoles removes the Keycloak artifacts created by
// CreateCourseGroupsAndRoles: the client roles named after the course and the
// course group (Keycloak cascades its Lecturer/Editor/CustomGroups subgroups).
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

// deleteCourseGroupsAndRoles deletes the roles first and the group last, so a failure part way
// through keeps the group and its memberships. Anything already gone counts as deleted, so a
// retry is safe.
func (s *KeycloakRealmService) deleteCourseGroupsAndRoles(ctx context.Context, courseGroupName string) error {
	token, err := s.LoginClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to log in to Keycloak: %w", err)
	}

	groupPath := TOP_LEVEL_GROUP_NAME + "/" + courseGroupName
	groupID, found, err := s.findGroupAtPath(ctx, token.AccessToken, groupPath)
	if err != nil {
		return err
	}

	// Group deletion only drops the group-role mappings, not the role definitions themselves.
	if err := s.deleteCourseClientRoles(ctx, token.AccessToken, courseGroupName); err != nil {
		return err
	}

	if !found {
		log.Warnf("course %s group not found in Keycloak; skipping group deletion", courseGroupName)
		return nil
	}
	if err := s.client.DeleteGroup(ctx, token.AccessToken, s.Realm, groupID); err != nil && !hasStatus(err, http.StatusNotFound) {
		return fmt.Errorf("failed to delete course group: %w", err)
	}
	return nil
}

func (s *KeycloakRealmService) deleteCourseClientRoles(ctx context.Context, accessToken, courseGroupName string) error {
	roles, err := s.client.GetClientRoles(ctx, accessToken, s.Realm, s.idOfClient, gocloak.GetRoleParams{})
	if err != nil {
		return fmt.Errorf("failed to list client roles: %w", err)
	}

	for _, role := range roles {
		if role == nil || role.Name == nil || !isCourseClientRole(*role.Name, courseGroupName) {
			continue
		}
		if err := s.client.DeleteClientRole(ctx, accessToken, s.Realm, s.idOfClient, *role.Name); err != nil && !hasStatus(err, http.StatusNotFound) {
			return fmt.Errorf("failed to delete client role %s: %w", *role.Name, err)
		}
	}
	return nil
}

// isCourseClientRole reports whether a client role was created for the given
// course. Creation produces "<group>-Lecturer", "<group>-Editor" and custom
// "<group>-cg-<name>" roles. The staff roles are matched exactly and custom
// roles by the "-cg-" prefix so a course whose group name is a prefix of
// another's (e.g. "ws24-ios" vs "ws24-ios-advanced") is not caught by mistake.
func isCourseClientRole(roleName, courseGroupName string) bool {
	if roleName == courseGroupName+"-"+permissionValidation.CourseLecturer ||
		roleName == courseGroupName+"-"+permissionValidation.CourseEditor {
		return true
	}
	return strings.HasPrefix(roleName, courseGroupName+"-cg-")
}
