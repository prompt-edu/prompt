package keycloakRealmManager

import (
	"context"

	"github.com/Nerzal/gocloak/v14"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

func (s *KeycloakRealmService) CreateCourseGroupsAndRoles(ctx context.Context, courseName, iterationName, userID string) error {
	token, err := s.LoginClient(ctx)
	if err != nil {
		return err
	}

	promptGroupID, err := s.GetOrCreatePromptGroup(ctx, token.AccessToken)
	if err != nil {
		return err
	}

	courseGroupName := permissionValidation.CourseIdentifier(iterationName, courseName)
	courseGroupID, err := s.CreateChildGroup(ctx, token.AccessToken, courseGroupName, promptGroupID)
	if err != nil {
		return err
	}

	subGroupNames := []string{permissionValidation.CourseLecturer, permissionValidation.CourseEditor}
	for _, groupName := range subGroupNames {
		// create role for the group
		roleName := permissionValidation.CourseRoleName(courseGroupName, groupName)
		role, err := s.GetOrCreateRealmRole(ctx, token.AccessToken, roleName)
		if err != nil {
			return err
		}

		// Create Subgroup with courseGroup as parent
		subGroupID, err := s.CreateChildGroup(ctx, token.AccessToken, groupName, courseGroupID)
		if err != nil {
			return err
		}

		// Associate role with group
		err = s.client.AddClientRolesToGroup(ctx, token.AccessToken, s.Realm, s.idOfClient, subGroupID, []gocloak.Role{*role})
		if err != nil {
			log.Error("failed to associate role with group: ", err)
			return err
		}

		// Add the requester to the lecturer group
		if groupName == permissionValidation.CourseLecturer {
			err = s.client.AddUserToGroup(ctx, token.AccessToken, s.Realm, userID, subGroupID)
			if err != nil {
				log.Error("failed to add user to group: ", err)
				return err
			}
		}
	}
	return nil
}
