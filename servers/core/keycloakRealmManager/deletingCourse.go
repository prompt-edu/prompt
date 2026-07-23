package keycloakRealmManager

import (
	"context"
	"fmt"
	"strings"

	"github.com/Nerzal/gocloak/v14"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

// DeleteCourseGroupsAndRoles removes the Keycloak artifacts created by
// CreateCourseGroupsAndRoles: the course group (Keycloak cascades its
// Lecturer/Editor/CustomGroups subgroups) and the client roles named after the
// course. It must run before the course row is removed from the database, since
// the group name is derived from it. A group that is already gone is treated as
// success so the operation stays idempotent.
func DeleteCourseGroupsAndRoles(ctx context.Context, courseID uuid.UUID) error {
	token, err := LoginClient(ctx)
	if err != nil {
		return err
	}

	courseGroupName, err := GetCourseGroupName(ctx, courseID)
	if err != nil {
		return fmt.Errorf("failed to get course group name: %w", err)
	}

	// 1. Delete the course group; Keycloak cascades subgroups and role mappings.
	group, err := GetCourseGroup(ctx, token.AccessToken, courseID)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to resolve course group: %w", err)
		}
		log.Warnf("course %s group not found in Keycloak; skipping group deletion", courseID)
	} else if err := KeycloakRealmSingleton.client.DeleteGroup(ctx, token.AccessToken, KeycloakRealmSingleton.Realm, *group.ID); err != nil {
		return fmt.Errorf("failed to delete course group: %w", err)
	}

	// 2. Delete the client roles. Group deletion only drops the group-role
	// mapping, not the role definitions themselves.
	return deleteCourseClientRoles(ctx, token.AccessToken, courseGroupName)
}

func deleteCourseClientRoles(ctx context.Context, accessToken, courseGroupName string) error {
	roles, err := KeycloakRealmSingleton.client.GetClientRoles(ctx, accessToken, KeycloakRealmSingleton.Realm, KeycloakRealmSingleton.idOfClient, gocloak.GetRoleParams{})
	if err != nil {
		return fmt.Errorf("failed to list client roles: %w", err)
	}

	for _, role := range roles {
		if role == nil || role.Name == nil || !isCourseClientRole(*role.Name, courseGroupName) {
			continue
		}
		if err := KeycloakRealmSingleton.client.DeleteClientRole(ctx, accessToken, KeycloakRealmSingleton.Realm, KeycloakRealmSingleton.idOfClient, *role.Name); err != nil {
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
