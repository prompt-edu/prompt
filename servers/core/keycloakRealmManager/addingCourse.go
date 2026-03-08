package keycloakRealmManager

import (
	"context"
	"fmt"

	"github.com/Nerzal/gocloak/v13"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

func CreateCourseGroupsAndRoles(ctx context.Context, courseName, iterationName, userID string) error {
	token, err := LoginClient(ctx)
	if err != nil {
		return err
	}

	promptGroupID, err := GetOrCreatePromptGroup(ctx, token.AccessToken)
	if err != nil {
		return err
	}

	courseGroupName := fmt.Sprintf("%s-%s", iterationName, courseName)
	courseGroupID, err := CreateChildGroup(ctx, token.AccessToken, courseGroupName, promptGroupID)
	if err != nil {
		return err
	}

	subGroupNames := []string{permissionValidation.CourseLecturer, permissionValidation.CourseEditor}
	for _, groupName := range subGroupNames {
		// create role for the group
		roleName := fmt.Sprintf("%s-%s-%s", iterationName, courseName, groupName)
		role, err := GetOrCreateRealmRole(ctx, token.AccessToken, roleName)
		if err != nil {
			return err
		}

		// Create Subgroup with courseGroup as parent
		subGroupID, err := CreateChildGroup(ctx, token.AccessToken, groupName, courseGroupID)
		if err != nil {
			return err
		}

		// Associate role with group
		err = KeycloakRealmSingleton.client.AddClientRolesToGroup(ctx, token.AccessToken, KeycloakRealmSingleton.Realm, KeycloakRealmSingleton.idOfClient, subGroupID, []gocloak.Role{*role})
		if err != nil {
			log.Error("failed to associate role with group: ", err)
			return err
		}

		// Add the requester to the lecturer group
		if groupName == permissionValidation.CourseLecturer {
			err = KeycloakRealmSingleton.client.AddUserToGroup(ctx, token.AccessToken, KeycloakRealmSingleton.Realm, userID, subGroupID)
			if err != nil {
				log.Error("failed to add user to group: ", err)
				return err
			}
		}
	}
	return nil
}
