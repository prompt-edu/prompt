package keycloakRealmManager

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Nerzal/gocloak/v13"
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	log "github.com/sirupsen/logrus"
)

// TODO: think about caching token at some point
// Get connection tokes
func LoginClient(ctx context.Context) (*gocloak.JWT, error) {
	token, err := KeycloakRealmSingleton.client.LoginClient(ctx, KeycloakRealmSingleton.ClientID, KeycloakRealmSingleton.ClientSecret, KeycloakRealmSingleton.Realm)
	if err != nil {
		log.Error("failed to authenticate to Keycloak: ", err)
		return nil, err
	}
	return token, nil
}

// GetGroupByPath wraps client.GetGroupByPath, returning an error if
// there is any Keycloak error or if the retrieved group name mismatches.
func GetGroupByPath(ctx context.Context, accessToken, groupPath, expectedName string) (*gocloak.Group, error) {
	group, err := KeycloakRealmSingleton.client.GetGroupByPath(ctx, accessToken, KeycloakRealmSingleton.Realm, groupPath)
	if err != nil {
		log.Errorf("failed to get group from Keycloak (path=%s): %v", groupPath, err)
		return nil, fmt.Errorf("failed to get group at path %s: %w", groupPath, err)
	}
	if group == nil || group.Name == nil || *group.Name != expectedName {
		log.Errorf("group name mismatch at path=%s, expected=%s, got=%v", groupPath, expectedName, group.Name)
		return nil, fmt.Errorf("group name mismatch or not found at path %s", groupPath)
	}
	return group, nil
}

// CreateChildGroup is a small wrapper to create a child group under parentGroupID.
func CreateChildGroup(ctx context.Context, accessToken, groupName, parentGroupID string) (string, error) {
	group := gocloak.Group{Name: &groupName}
	childGroupID, err := KeycloakRealmSingleton.client.CreateChildGroup(ctx, accessToken, KeycloakRealmSingleton.Realm, parentGroupID, group)
	if err != nil {
		log.Error("failed to create child group: ", err)
		return "", errors.New("failed to create Keycloak group")
	}
	return childGroupID, nil
}

// GetOrCreatePromptGroup tries to find a top-level group exactly named TOP_LEVEL_GROUP_NAME.
// If it doesn’t exist, it creates it. Returns the group ID either way.
func GetOrCreatePromptGroup(ctx context.Context, accessToken string) (string, error) {
	exact := true
	groups, err := KeycloakRealmSingleton.client.GetGroups(ctx, accessToken, KeycloakRealmSingleton.Realm, gocloak.GetGroupsParams{
		Search: &TOP_LEVEL_GROUP_NAME,
		Exact:  &exact,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get groups from Keycloak: %w", err)
	}

	if len(groups) == 1 && groups[0].Name != nil && *groups[0].Name == TOP_LEVEL_GROUP_NAME {
		return *groups[0].ID, nil
	}

	// If not found, create the group
	group := gocloak.Group{Name: &TOP_LEVEL_GROUP_NAME}
	baseGroupID, err := KeycloakRealmSingleton.client.CreateGroup(ctx, accessToken, KeycloakRealmSingleton.Realm, group)
	if err != nil {
		log.Error("failed to create base group: ", err)
		return "", errors.New("failed to create keycloak group")
	}
	return baseGroupID, nil
}

// GetCourseGroupName fetches the course from DB and constructs its Keycloak group name.
func GetCourseGroupName(ctx context.Context, courseID uuid.UUID) (string, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	course, err := KeycloakRealmSingleton.queries.GetCourse(ctxWithTimeout, courseID)
	if err != nil {
		return "", fmt.Errorf("failed to get course: %w", err)
	}
	courseGroupName := course.SemesterTag.String + "-" + course.Name
	return courseGroupName, nil
}

// GetCourseGroup returns the Keycloak group for a specific course, if it exists.
func GetCourseGroup(ctx context.Context, accessToken string, courseID uuid.UUID) (*gocloak.Group, error) {
	courseGroupName, err := GetCourseGroupName(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get course group name: %w", err)
	}
	groupPath := "/" + TOP_LEVEL_GROUP_NAME + "/" + courseGroupName
	return GetGroupByPath(ctx, accessToken, groupPath, courseGroupName)
}

func GetCourseEditorGroup(ctx context.Context, accessToken string, courseID uuid.UUID) (*gocloak.Group, error) {
	courseGroupName, err := GetCourseGroupName(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get course group name: %w", err)
	}
	groupPath := "/" + TOP_LEVEL_GROUP_NAME + "/" + courseGroupName + "/" + "Editor"

	return GetGroupByPath(ctx, accessToken, groupPath, "Editor")
}

// GetOrCreateCustomTopLevelGroup returns the ID of the “CUSTOM_GROUPS_NAME” subgroup under the
// course group. If it doesn’t exist, it creates it.
func GetOrCreateCustomTopLevelGroup(ctx context.Context, accessToken string, courseID uuid.UUID) (string, error) {
	courseGroupName, err := GetCourseGroupName(ctx, courseID)
	if err != nil {
		return "", fmt.Errorf("failed to get course group name: %w", err)
	}

	// Build path: /TOP_LEVEL_GROUP_NAME/<courseGroupName>/CUSTOM_GROUPS_NAME
	groupPath := fmt.Sprintf("/%s/%s/%s", TOP_LEVEL_GROUP_NAME, courseGroupName, CUSTOM_GROUPS_NAME)
	group, err := KeycloakRealmSingleton.client.GetGroupByPath(ctx, accessToken, KeycloakRealmSingleton.Realm, groupPath)
	if err == nil && group.Name != nil && *group.Name == CUSTOM_GROUPS_NAME {
		// Found existing group
		return *group.ID, nil
	} else if err != nil && !strings.Contains(err.Error(), "404") {
		// If we hit an error other than 404, it’s a real error
		log.Errorf("failed to get group from Keycloak for path [%s]: %v", groupPath, err)
		return "", fmt.Errorf("failed to get group from Keycloak: %w", err)
	}

	// Not found (404) – we must create the group
	courseGroup, err := GetCourseGroup(ctx, accessToken, courseID)
	if err != nil {
		log.Error("failed to get course group: ", err)
		return "", fmt.Errorf("failed to get course group: %w", err)
	}

	newGroupID, err := CreateChildGroup(ctx, accessToken, CUSTOM_GROUPS_NAME, *courseGroup.ID)
	if err != nil {
		log.Error("failed to create custom top-level group: ", err)
		return "", fmt.Errorf("failed to create custom top-level group: %w", err)
	}
	return newGroupID, nil
}

// GetOrCreateCustomGroup checks if a child group named groupName exists under parentGroupID,
// creates it otherwise.
func GetOrCreateCustomGroup(ctx context.Context, accessToken, groupName string, courseID uuid.UUID) (string, error) {
	// Ensure the custom top-level group (e.g., "/.../CUSTOM_GROUPS_NAME") exists.
	customTopLevelGroupID, err := GetOrCreateCustomTopLevelGroup(ctx, accessToken, courseID)
	if err != nil {
		log.Error("failed to get or create custom top-level group: ", err)
		return "", fmt.Errorf("failed to get or create custom top-level group: %w", err)
	}

	courseGroupName, err := GetCourseGroupName(ctx, courseID)
	if err != nil {
		return "", fmt.Errorf("failed to get course group name: %w", err)
	}

	// Build path: /TOP_LEVEL_GROUP_NAME/<courseGroupName>/CUSTOM_GROUPS_NAME/<groupName>
	groupPath := fmt.Sprintf("/%s/%s/%s/%s", TOP_LEVEL_GROUP_NAME, courseGroupName, CUSTOM_GROUPS_NAME, groupName)
	group, err := KeycloakRealmSingleton.client.GetGroupByPath(ctx, accessToken, KeycloakRealmSingleton.Realm, groupPath)
	if err == nil && group.Name != nil && *group.Name == groupName {
		// Found existing subgroup
		return *group.ID, nil
	} else if err != nil && !strings.Contains(err.Error(), "404") {
		log.Errorf("failed to get group from Keycloak for path [%s]: %v", groupPath, err)
		return "", fmt.Errorf("failed to get group: %w", err)
	}

	// Not found (404), create new child group under the custom top-level group
	newGroupID, err := CreateChildGroup(ctx, accessToken, groupName, customTopLevelGroupID)
	if err != nil {
		log.Error("failed to create child group: ", err)
		return "", fmt.Errorf("failed to create child group: %w", err)
	}

	return newGroupID, nil
}

// GetCustomGroupID returns the ID of customGroupName under the course’s “CUSTOM_GROUPS_NAME”.
func GetCustomGroupID(ctx context.Context, accessToken, customGroupName string, courseID uuid.UUID) (string, error) {
	courseGroupName, err := GetCourseGroupName(ctx, courseID)
	if err != nil {
		return "", fmt.Errorf("failed to get course group name: %w", err)
	}
	groupPath := "/" + TOP_LEVEL_GROUP_NAME + "/" + courseGroupName + "/" + CUSTOM_GROUPS_NAME + "/" + customGroupName

	group, err := GetGroupByPath(ctx, accessToken, groupPath, customGroupName)
	if err != nil {
		log.Error("failed to get custom group from Keycloak: ", err)
		return "", errors.New("failed to get custom group")
	}
	return *group.ID, nil
}

// GetOrCreateRealmRole fetches or creates a role for the Keycloak client identified by idOfClient.
func GetOrCreateRealmRole(ctx context.Context, accessToken, roleName string) (*gocloak.Role, error) {
	// Trying to get (check if exists)
	existingRole, err := KeycloakRealmSingleton.client.GetClientRole(ctx, accessToken, KeycloakRealmSingleton.Realm, KeycloakRealmSingleton.idOfClient, roleName)
	if err == nil {
		// Role already exists
		log.Debug("Role already exists: ", existingRole.ID)
		return existingRole, nil
	} else if !strings.Contains(err.Error(), "404") {
		log.Error("failed to get role: ", err)
		return nil, err
	}

	// Creating realm role (only returns name)
	name, err := KeycloakRealmSingleton.client.CreateClientRole(ctx, accessToken, KeycloakRealmSingleton.Realm, KeycloakRealmSingleton.idOfClient, gocloak.Role{Name: &roleName})
	if err != nil {
		log.Error("failed to create role: ", err)
		return nil, err
	}

	// Getting the just created role
	createdRole, err := KeycloakRealmSingleton.client.GetClientRole(ctx, accessToken, KeycloakRealmSingleton.Realm, KeycloakRealmSingleton.idOfClient, name)
	if err != nil {
		log.Error("failed to get newly created role: ", err)
		return nil, err
	}
	return createdRole, nil
}

// AddRoleToGroup associates the provided role with the specified group (no-op if already associated).
func AddRoleToGroup(ctx context.Context, accessToken, groupID string, role *gocloak.Role) error {
	err := KeycloakRealmSingleton.client.AddClientRolesToGroup(ctx, accessToken, KeycloakRealmSingleton.Realm, KeycloakRealmSingleton.idOfClient, groupID, []gocloak.Role{*role})
	if err != nil {
		log.Error("failed to associate role with group: ", err)
		return err
	}
	return nil
}

// AddStudentIDsToKeycloakGroup puts each student into the given group.
// Returns slices of succeeded and failed student UUIDs.
func AddStudentIDsToKeycloakGroup(ctx context.Context, accessToken string, studentIDs []uuid.UUID, groupID string) ([]uuid.UUID, []uuid.UUID, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	students, err := KeycloakRealmSingleton.queries.GetStudentUniversityLogins(ctxWithTimeout, studentIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get student emails: %w", err)
	}
	if len(students) != len(studentIDs) {
		return nil, nil, errors.New("not all students found in DB")
	}

	var failedStudents []uuid.UUID
	var succeededStudents []uuid.UUID
	for _, student := range students {
		// Get the keycloak user to the student email
		keycloakUser, err := KeycloakRealmSingleton.client.GetUsers(ctxWithTimeout, accessToken, KeycloakRealmSingleton.Realm, gocloak.GetUsersParams{
			Username: &student.UniversityLogin.String,
		})

		if err != nil || len(keycloakUser) != 1 {
			log.Error("failed to get keycloak user for student: ", err)
			failedStudents = append(failedStudents, student.ID)
			continue
		}

		// Add user to the group
		err = KeycloakRealmSingleton.client.AddUserToGroup(ctxWithTimeout, accessToken, KeycloakRealmSingleton.Realm, *keycloakUser[0].ID, groupID)
		if err != nil {
			log.Error("failed to get keycloak user for student: ", err)
			failedStudents = append(failedStudents, student.ID)
			continue
		}
		succeededStudents = append(succeededStudents, student.ID)
	}

	return succeededStudents, failedStudents, nil
}

// GetGroupMembers returns the users that belong to the given group.
func GetGroupMembers(ctx context.Context, accessToken, groupID string) ([]*gocloak.User, error) {
	members, err := KeycloakRealmSingleton.client.GetGroupMembers(ctx, accessToken, KeycloakRealmSingleton.Realm, groupID, gocloak.GetGroupsParams{})
	if err != nil {
		log.Error("failed to get group members: ", err)
		return nil, err
	}
	return members, nil
}
