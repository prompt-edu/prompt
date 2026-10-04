package keycloakRealmManager

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Nerzal/gocloak/v14"
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

// TODO: think about caching token at some point
// Get connection tokes
func (s *KeycloakRealmService) LoginClient(ctx context.Context) (*gocloak.JWT, error) {
	token, err := s.client.LoginClient(ctx, s.ClientID, s.ClientSecret, s.Realm)
	if err != nil {
		log.Error("failed to authenticate to Keycloak: ", err)
		return nil, err
	}
	return token, nil
}

// GetGroupByPath wraps client.GetGroupByPath, returning an error if
// there is any Keycloak error or if the retrieved group name mismatches.
//
// The leading slash is stripped before delegating to gocloak. gocloak builds the
// admin URL by joining path segments with "/", and a leading-slash groupPath
// produces a "group-by-path//..." double slash that recent Keycloak versions
// reject under strict path normalisation. Keycloak itself strips a leading slash
// from the path internally (see KeycloakModelUtils.findGroupByPath), so passing
// it without is equivalent at the business-logic layer.
func (s *KeycloakRealmService) GetGroupByPath(ctx context.Context, accessToken, groupPath, expectedName string) (*gocloak.Group, error) {
	normalisedPath := strings.TrimPrefix(groupPath, "/")
	group, err := s.client.GetGroupByPath(ctx, accessToken, s.Realm, normalisedPath)
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
func (s *KeycloakRealmService) CreateChildGroup(ctx context.Context, accessToken, groupName, parentGroupID string) (string, error) {
	group := gocloak.Group{Name: &groupName}
	childGroupID, err := s.client.CreateChildGroup(ctx, accessToken, s.Realm, parentGroupID, group)
	if err != nil {
		log.Error("failed to create child group: ", err)
		return "", errors.New("failed to create Keycloak group")
	}
	return childGroupID, nil
}

// GetOrCreatePromptGroup tries to find a top-level group exactly named TOP_LEVEL_GROUP_NAME.
// If it doesn’t exist, it creates it. Returns the group ID either way.
func (s *KeycloakRealmService) GetOrCreatePromptGroup(ctx context.Context, accessToken string) (string, error) {
	exact := true
	groups, err := s.client.GetGroups(ctx, accessToken, s.Realm, gocloak.GetGroupsParams{
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
	baseGroupID, err := s.client.CreateGroup(ctx, accessToken, s.Realm, group)
	if err != nil {
		log.Error("failed to create base group: ", err)
		return "", errors.New("failed to create keycloak group")
	}
	return baseGroupID, nil
}

// GetCourseGroupName fetches the course from DB and constructs its Keycloak group name.
func (s *KeycloakRealmService) GetCourseGroupName(ctx context.Context, courseID uuid.UUID) (string, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	course, err := s.queries.GetCourse(ctxWithTimeout, courseID)
	if err != nil {
		return "", fmt.Errorf("failed to get course: %w", err)
	}
	courseGroupName := permissionValidation.CourseIdentifier(course.SemesterTag.String, course.Name)
	return courseGroupName, nil
}

// GetCourseGroup returns the Keycloak group for a specific course, if it exists.
func (s *KeycloakRealmService) GetCourseGroup(ctx context.Context, accessToken string, courseID uuid.UUID) (*gocloak.Group, error) {
	courseGroupName, err := s.GetCourseGroupName(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get course group name: %w", err)
	}
	groupPath := "/" + TOP_LEVEL_GROUP_NAME + "/" + courseGroupName
	return s.GetGroupByPath(ctx, accessToken, groupPath, courseGroupName)
}

// GetCourseEditorGroup returns the Editor subgroup for a course.
func (s *KeycloakRealmService) GetCourseEditorGroup(ctx context.Context, accessToken string, courseID uuid.UUID) (*gocloak.Group, error) {
	return s.GetCourseSubgroup(ctx, accessToken, courseID, "Editor")
}

// GetCourseSubgroup returns the Keycloak group at /{TOP_LEVEL}/{courseGroupName}/{subgroup}.
// subgroup MUST be a compile-time constant from permissionValidation (CourseLecturer or
// CourseEditor); callers are responsible for validating any caller-supplied value against
// an allow-list before invoking this function. The subgroup value is passed to
// GetGroupByPath as expectedName so the existing name-mismatch guard fires if Keycloak
// returns a misdirected group.
func (s *KeycloakRealmService) GetCourseSubgroup(ctx context.Context, accessToken string, courseID uuid.UUID, subgroup string) (*gocloak.Group, error) {
	courseGroupName, err := s.GetCourseGroupName(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get course group name: %w", err)
	}
	groupPath := "/" + TOP_LEVEL_GROUP_NAME + "/" + courseGroupName + "/" + subgroup
	return s.GetGroupByPath(ctx, accessToken, groupPath, subgroup)
}

// GetOrCreateCustomTopLevelGroup returns the ID of the “CUSTOM_GROUPS_NAME” subgroup under the
// course group. If it doesn’t exist, it creates it.
func (s *KeycloakRealmService) GetOrCreateCustomTopLevelGroup(ctx context.Context, accessToken string, courseID uuid.UUID) (string, error) {
	courseGroupName, err := s.GetCourseGroupName(ctx, courseID)
	if err != nil {
		return "", fmt.Errorf("failed to get course group name: %w", err)
	}

	// Build path: TOP_LEVEL_GROUP_NAME/<courseGroupName>/CUSTOM_GROUPS_NAME (no leading
	// slash - see GetGroupByPath docstring for the URL-normalisation reason).
	groupPath := fmt.Sprintf("%s/%s/%s", TOP_LEVEL_GROUP_NAME, courseGroupName, CUSTOM_GROUPS_NAME)
	group, err := s.client.GetGroupByPath(ctx, accessToken, s.Realm, groupPath)
	if err == nil && group.Name != nil && *group.Name == CUSTOM_GROUPS_NAME {
		// Found existing group
		return *group.ID, nil
	} else if err != nil && !strings.Contains(err.Error(), "404") {
		// If we hit an error other than 404, it’s a real error
		log.Errorf("failed to get group from Keycloak for path [%s]: %v", groupPath, err)
		return "", fmt.Errorf("failed to get group from Keycloak: %w", err)
	}

	// Not found (404) – we must create the group
	courseGroup, err := s.GetCourseGroup(ctx, accessToken, courseID)
	if err != nil {
		log.Error("failed to get course group: ", err)
		return "", fmt.Errorf("failed to get course group: %w", err)
	}

	newGroupID, err := s.CreateChildGroup(ctx, accessToken, CUSTOM_GROUPS_NAME, *courseGroup.ID)
	if err != nil {
		log.Error("failed to create custom top-level group: ", err)
		return "", fmt.Errorf("failed to create custom top-level group: %w", err)
	}
	return newGroupID, nil
}

// GetOrCreateCustomGroup checks if a child group named groupName exists under parentGroupID,
// creates it otherwise.
func (s *KeycloakRealmService) GetOrCreateCustomGroup(ctx context.Context, accessToken, groupName string, courseID uuid.UUID) (string, error) {
	// Ensure the custom top-level group (e.g., "/.../CUSTOM_GROUPS_NAME") exists.
	customTopLevelGroupID, err := s.GetOrCreateCustomTopLevelGroup(ctx, accessToken, courseID)
	if err != nil {
		log.Error("failed to get or create custom top-level group: ", err)
		return "", fmt.Errorf("failed to get or create custom top-level group: %w", err)
	}

	courseGroupName, err := s.GetCourseGroupName(ctx, courseID)
	if err != nil {
		return "", fmt.Errorf("failed to get course group name: %w", err)
	}

	// Build path: TOP_LEVEL_GROUP_NAME/<courseGroupName>/CUSTOM_GROUPS_NAME/<groupName>
	// (no leading slash - see GetGroupByPath docstring).
	groupPath := fmt.Sprintf("%s/%s/%s/%s", TOP_LEVEL_GROUP_NAME, courseGroupName, CUSTOM_GROUPS_NAME, groupName)
	group, err := s.client.GetGroupByPath(ctx, accessToken, s.Realm, groupPath)
	if err == nil && group.Name != nil && *group.Name == groupName {
		// Found existing subgroup
		return *group.ID, nil
	} else if err != nil && !strings.Contains(err.Error(), "404") {
		log.Errorf("failed to get group from Keycloak for path [%s]: %v", groupPath, err)
		return "", fmt.Errorf("failed to get group: %w", err)
	}

	// Not found (404), create new child group under the custom top-level group
	newGroupID, err := s.CreateChildGroup(ctx, accessToken, groupName, customTopLevelGroupID)
	if err != nil {
		log.Error("failed to create child group: ", err)
		return "", fmt.Errorf("failed to create child group: %w", err)
	}

	return newGroupID, nil
}

// GetCustomGroupID returns the ID of customGroupName under the course’s “CUSTOM_GROUPS_NAME”.
func (s *KeycloakRealmService) GetCustomGroupID(ctx context.Context, accessToken, customGroupName string, courseID uuid.UUID) (string, error) {
	courseGroupName, err := s.GetCourseGroupName(ctx, courseID)
	if err != nil {
		return "", fmt.Errorf("failed to get course group name: %w", err)
	}
	groupPath := "/" + TOP_LEVEL_GROUP_NAME + "/" + courseGroupName + "/" + CUSTOM_GROUPS_NAME + "/" + customGroupName

	group, err := s.GetGroupByPath(ctx, accessToken, groupPath, customGroupName)
	if err != nil {
		log.Error("failed to get custom group from Keycloak: ", err)
		return "", errors.New("failed to get custom group")
	}
	return *group.ID, nil
}

// GetOrCreateRealmRole fetches or creates a role for the Keycloak client identified by idOfClient.
func (s *KeycloakRealmService) GetOrCreateRealmRole(ctx context.Context, accessToken, roleName string) (*gocloak.Role, error) {
	// Trying to get (check if exists)
	existingRole, err := s.client.GetClientRole(ctx, accessToken, s.Realm, s.idOfClient, roleName)
	if err == nil {
		// Role already exists
		log.Debug("Role already exists: ", existingRole.ID)
		return existingRole, nil
	} else if !strings.Contains(err.Error(), "404") {
		log.Error("failed to get role: ", err)
		return nil, err
	}

	// Creating realm role (only returns name)
	name, err := s.client.CreateClientRole(ctx, accessToken, s.Realm, s.idOfClient, gocloak.Role{Name: &roleName})
	if err != nil {
		log.Error("failed to create role: ", err)
		return nil, err
	}

	// Getting the just created role
	createdRole, err := s.client.GetClientRole(ctx, accessToken, s.Realm, s.idOfClient, name)
	if err != nil {
		log.Error("failed to get newly created role: ", err)
		return nil, err
	}
	return createdRole, nil
}

// AddRoleToGroup associates the provided role with the specified group (no-op if already associated).
func (s *KeycloakRealmService) AddRoleToGroup(ctx context.Context, accessToken, groupID string, role *gocloak.Role) error {
	err := s.client.AddClientRolesToGroup(ctx, accessToken, s.Realm, s.idOfClient, groupID, []gocloak.Role{*role})
	if err != nil {
		log.Error("failed to associate role with group: ", err)
		return err
	}
	return nil
}

// AddStudentIDsToKeycloakGroup puts each student into the given group.
// Returns slices of succeeded and failed student UUIDs.
func (s *KeycloakRealmService) AddStudentIDsToKeycloakGroup(ctx context.Context, accessToken string, studentIDs []uuid.UUID, groupID string) ([]uuid.UUID, []uuid.UUID, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	students, err := s.queries.GetStudentUniversityLogins(ctxWithTimeout, studentIDs)
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
		keycloakUser, err := s.client.GetUsers(ctxWithTimeout, accessToken, s.Realm, gocloak.GetUsersParams{
			Username: &student.UniversityLogin.String,
		})

		if err != nil || len(keycloakUser) != 1 {
			log.Error("failed to get keycloak user for student: ", err)
			failedStudents = append(failedStudents, student.ID)
			continue
		}

		// Add user to the group
		err = s.client.AddUserToGroup(ctxWithTimeout, accessToken, s.Realm, *keycloakUser[0].ID, groupID)
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
func (s *KeycloakRealmService) GetGroupMembers(ctx context.Context, accessToken, groupID string) ([]*gocloak.User, error) {
	members, err := s.client.GetGroupMembers(ctx, accessToken, s.Realm, groupID, gocloak.GetGroupsParams{})
	if err != nil {
		log.Error("failed to get group members: ", err)
		return nil, err
	}
	return members, nil
}
