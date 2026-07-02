package keycloakRealmManager

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Nerzal/gocloak/v14"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/keycloakRealmManager/keycloakRealmDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
	log "github.com/sirupsen/logrus"
)

// Sentinel errors for the course-staff management API. Handlers map these to
// HTTP status codes; everything else is surfaced as 500.
var (
	ErrInvalidGroupName = errors.New("invalid group name")
	ErrSelfRemoval      = errors.New("cannot remove yourself from a course group via this API; ask another instructor or use the Keycloak admin console")
	ErrInvalidQuery     = errors.New("search query must be at least 2 characters")
	ErrUserNotFound     = errors.New("keycloak user not found")
)

// maxStaffMembers caps how many users are fetched per Lecturer/Editor group.
// Keycloak's default page size is 100; we explicitly request a higher value to
// avoid silently truncating larger course staffs.
const maxStaffMembers = 200

// maxSearchResults caps a user search response. Larger values are clamped down.
const maxSearchResults = 50

// isAllowedCourseGroup returns true if name is one of the two course-staff
// subgroups we expose for management.
func isAllowedCourseGroup(name string) bool {
	return name == permissionValidation.CourseLecturer || name == permissionValidation.CourseEditor
}

func AddCustomGroup(ctx context.Context, courseID uuid.UUID, groupName string) (string, error) {
	courseGroupName, err := GetCourseGroupName(ctx, courseID)
	if err != nil {
		return "", fmt.Errorf("failed to get course group name: %w", err)
	}

	// 1. Log into keycloak
	token, err := LoginClient(ctx)
	if err != nil {
		return "", err
	}

	// 2. Get Custom subgroup or create
	customGroupID, err := GetOrCreateCustomGroup(ctx, token.AccessToken, groupName, courseID)
	if err != nil {
		log.Error("Failed to get or create custom group: ", err)
		return "", errors.New("failed to get or create custom top level group")
	}

	// 3. Create desired role
	roleName := courseGroupName + "-cg-" + groupName
	role, err := GetOrCreateRealmRole(ctx, token.AccessToken, roleName)
	if err != nil {
		log.Error("failed to create role: ", err)
		return "", errors.New("failed to create keycloak roles")
	}

	// 4. Associate role with group
	err = AddRoleToGroup(ctx, token.AccessToken, customGroupID, role)
	if err != nil {
		log.Error("failed to associate role with group: ", err)
		return "", errors.New("failed to associate role with group")
	}

	return customGroupID, nil
}

func AddStudentsToGroup(ctx context.Context, courseID uuid.UUID, studentIDs []uuid.UUID, groupName string) (keycloakRealmDTO.AddStudentsToGroupResponse, error) {
	// 1. Log into keycloak
	token, err := LoginClient(ctx)
	if err != nil {
		return keycloakRealmDTO.AddStudentsToGroupResponse{}, err
	}

	// 2. Get Custom Group Folder
	customGroupID, err := GetCustomGroupID(ctx, token.AccessToken, groupName, courseID)
	if err != nil {
		log.Error("Failed to get custom group: ", err)
		return keycloakRealmDTO.AddStudentsToGroupResponse{}, errors.New("failed to get custom group")
	}

	// 3. Get the keycloak userIDs of the students
	succeededStudents, failedStudentIDs, err := AddStudentIDsToKeycloakGroup(ctx, token.AccessToken, studentIDs, customGroupID)
	// some error occurred before adding students to group
	if err != nil {
		log.Error("Failed to add students to group: ", err)
		return keycloakRealmDTO.AddStudentsToGroupResponse{}, errors.New("failed to add students to group")
	}

	response := keycloakRealmDTO.AddStudentsToGroupResponse{
		SucceededToAddStudentIDs: succeededStudents,
		FailedToAddStudentIDs:    failedStudentIDs,
	}
	return response, nil
}

// GetStudentsInGroup retrieves the student IDs for a given course's group.
func GetStudentsInGroup(ctx context.Context, courseID uuid.UUID, groupName string) (keycloakRealmDTO.GroupMembers, error) {
	// 1. Log into Keycloak.
	token, err := LoginClient(ctx)
	if err != nil {
		return keycloakRealmDTO.GroupMembers{}, fmt.Errorf("failed to login to keycloak: %w", err)
	}

	// 2. Get Custom Group Folder.
	customGroupID, err := GetCustomGroupID(ctx, token.AccessToken, groupName, courseID)
	if err != nil {
		log.Error("Failed to get custom group", "error", err)
		return keycloakRealmDTO.GroupMembers{}, fmt.Errorf("failed to get custom group: %w", err)
	}

	// 3. Retrieve group members from Keycloak.
	members, err := GetGroupMembers(ctx, token.AccessToken, customGroupID)
	if err != nil {
		log.Error("Failed to get group members", "error", err)
		return keycloakRealmDTO.GroupMembers{}, fmt.Errorf("failed to get group members: %w", err)
	}

	// Build a slice of emails from the group members.
	// (Skip any members without an email.)
	var memberEmails []string
	for _, member := range members {
		if member.Email == nil || *member.Email == "" {
			log.Warn("Skipping member with missing email", "member", member)
			continue
		}
		memberEmails = append(memberEmails, *member.Email)
	}

	// 4. Get students from the database using the list of emails.
	studentsObjects, err := KeycloakRealmSingleton.queries.GetStudentsByEmail(ctx, memberEmails)
	if err != nil {
		log.Error("Failed to get students by email", "error", err)
		return keycloakRealmDTO.GroupMembers{}, fmt.Errorf("failed to get students by email: %w", err)
	}

	// Convert database models to student DTOs.
	studentDTOs := make([]studentDTO.Student, len(studentsObjects))
	for i, student := range studentsObjects {
		studentDTOs[i] = studentDTO.GetStudentDTOFromDBModel(student)
	}

	// Create a lookup map from email to student DTO for quick access.
	studentByEmail := make(map[string]studentDTO.Student)
	for _, student := range studentDTOs {
		// Assuming that studentDTO.Student has an Email field.
		studentByEmail[student.Email] = student
	}

	var notFoundUsers []keycloakRealmDTO.KeycloakUser

	// 5. Check each group member against the student lookup.
	for _, member := range members {
		// Skip members with missing email.
		if member.Email == nil || *member.Email == "" {
			continue
		}
		email := *member.Email
		if _, exists := studentByEmail[email]; !exists {
			// Add to notFoundUsers if the member email was not found among the students.
			notFoundUsers = append(notFoundUsers, keycloakRealmDTO.GetKeycloakUserDTO(*member))
		}
	}

	return keycloakRealmDTO.GroupMembers{
		Students:    studentDTOs,
		NonStudents: notFoundUsers,
	}, nil
}

// GetCourseStaff returns the Lecturer and Editor members of a course.
func GetCourseStaff(ctx context.Context, courseID uuid.UUID) (keycloakRealmDTO.CourseStaff, error) {
	token, err := LoginClient(ctx)
	if err != nil {
		return keycloakRealmDTO.CourseStaff{}, err
	}

	lecturers, err := getCourseGroupMembers(ctx, token.AccessToken, courseID, permissionValidation.CourseLecturer)
	if err != nil {
		return keycloakRealmDTO.CourseStaff{}, err
	}

	editors, err := getCourseGroupMembers(ctx, token.AccessToken, courseID, permissionValidation.CourseEditor)
	if err != nil {
		return keycloakRealmDTO.CourseStaff{}, err
	}

	return keycloakRealmDTO.CourseStaff{
		Lecturers: lecturers,
		Editors:   editors,
	}, nil
}

func getCourseGroupMembers(ctx context.Context, accessToken string, courseID uuid.UUID, groupName string) ([]keycloakRealmDTO.StaffMember, error) {
	group, err := GetCourseSubgroup(ctx, accessToken, courseID, groupName)
	if err != nil {
		// A missing Lecturer/Editor subgroup (e.g. legacy course, manual cleanup)
		// should not blank out the other group's table. Treat 404 as "no members
		// yet" and let the caller render the other group normally. Genuine
		// permission/transport failures still bubble up as 500.
		if strings.Contains(err.Error(), "404") {
			log.Warnf("course %s %s group not found in Keycloak; treating as empty", courseID, groupName)
			return []keycloakRealmDTO.StaffMember{}, nil
		}
		return nil, fmt.Errorf("failed to resolve %s group: %w", groupName, err)
	}

	first := 0
	max := maxStaffMembers
	members, err := KeycloakRealmSingleton.client.GetGroupMembers(ctx, accessToken, KeycloakRealmSingleton.Realm, *group.ID, gocloak.GetGroupsParams{
		First: &first,
		Max:   &max,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch members of %s group: %w", groupName, err)
	}

	if len(members) == maxStaffMembers {
		log.Warnf("course %s %s group hit the %d-member fetch cap; some members are not shown", courseID, groupName, maxStaffMembers)
	}

	result := make([]keycloakRealmDTO.StaffMember, 0, len(members))
	for _, m := range members {
		result = append(result, keycloakRealmDTO.GetStaffMemberFromKeycloakUser(m))
	}
	return result, nil
}

// AddUserToCourseGroup adds a Keycloak user to the Lecturer or Editor group of
// a course. The groupName MUST come from the URL allow-list (the router
// enforces this). The target user is verified to exist before any group
// mutation. AddUserToGroup is idempotent at the Keycloak level.
func AddUserToCourseGroup(ctx context.Context, courseID uuid.UUID, groupName, targetUserID, callerUserID string) error {
	if !isAllowedCourseGroup(groupName) {
		return ErrInvalidGroupName
	}

	token, err := LoginClient(ctx)
	if err != nil {
		return err
	}

	if _, err := KeycloakRealmSingleton.client.GetUserByID(ctx, token.AccessToken, KeycloakRealmSingleton.Realm, targetUserID); err != nil {
		// gocloak v13 sometimes returns an APIError with Code=0 and the status text
		// in Message, so check both the structured code and the string as a fallback
		// (same pattern used elsewhere in realmManagement.go).
		var apiErr *gocloak.APIError
		if (errors.As(err, &apiErr) && apiErr.Code == 404) || strings.Contains(err.Error(), "404") {
			return ErrUserNotFound
		}
		log.Error("failed to verify target keycloak user: ", err)
		return fmt.Errorf("failed to verify keycloak user: %w", err)
	}

	group, err := GetCourseSubgroup(ctx, token.AccessToken, courseID, groupName)
	if err != nil {
		return fmt.Errorf("failed to resolve %s group: %w", groupName, err)
	}

	if err := KeycloakRealmSingleton.client.AddUserToGroup(ctx, token.AccessToken, KeycloakRealmSingleton.Realm, targetUserID, *group.ID); err != nil {
		log.Error("failed to add user to group: ", err)
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	log.Infof("course-staff audit: caller=%s action=add target=%s group=%s course=%s", callerUserID, targetUserID, groupName, courseID)
	return nil
}

// RemoveUserFromCourseGroup removes a Keycloak user from the Lecturer or
// Editor group of a course. Self-removal is rejected: with this rule, the
// final lecturer can never delete themselves, which makes a count check or
// advisory lock unnecessary.
func RemoveUserFromCourseGroup(ctx context.Context, courseID uuid.UUID, groupName, targetUserID, callerUserID string) error {
	if !isAllowedCourseGroup(groupName) {
		return ErrInvalidGroupName
	}
	if targetUserID == callerUserID {
		return ErrSelfRemoval
	}

	token, err := LoginClient(ctx)
	if err != nil {
		return err
	}

	group, err := GetCourseSubgroup(ctx, token.AccessToken, courseID, groupName)
	if err != nil {
		return fmt.Errorf("failed to resolve %s group: %w", groupName, err)
	}

	if err := KeycloakRealmSingleton.client.DeleteUserFromGroup(ctx, token.AccessToken, KeycloakRealmSingleton.Realm, targetUserID, *group.ID); err != nil {
		log.Error("failed to remove user from group: ", err)
		return fmt.Errorf("failed to remove user from group: %w", err)
	}

	log.Infof("course-staff audit: caller=%s action=remove target=%s group=%s course=%s", callerUserID, targetUserID, groupName, courseID)
	return nil
}

// GetKeycloakStatus runs three independent probes against the configured
// Keycloak service account: client-credentials login, read-users, and
// read-groups. Each probe is performed even if a previous one failed, so the
// caller can see exactly which permission is missing. Healthy is true only
// when all probes pass.
//
// The read-users and read-groups probes only check that the API call returns
// without error; an empty result set is treated as a pass (a freshly-seeded
// realm with no users or no top-level Prompt group yet should still report
// the service account as correctly configured).
func GetKeycloakStatus(ctx context.Context) keycloakRealmDTO.KeycloakStatus {
	capabilities := map[string]bool{
		keycloakRealmDTO.CapabilityKeycloakLogin:      false,
		keycloakRealmDTO.CapabilityKeycloakReadUsers:  false,
		keycloakRealmDTO.CapabilityKeycloakReadGroups: false,
	}

	token, err := LoginClient(ctx)
	if err != nil {
		log.Warn("keycloak status: login probe failed: ", err)
		return keycloakRealmDTO.KeycloakStatus{
			Healthy:      false,
			Capabilities: capabilities,
		}
	}
	capabilities[keycloakRealmDTO.CapabilityKeycloakLogin] = true

	first := 0
	max := 1
	brief := true
	if _, err := KeycloakRealmSingleton.client.GetUsers(ctx, token.AccessToken, KeycloakRealmSingleton.Realm, gocloak.GetUsersParams{
		First:               &first,
		Max:                 &max,
		BriefRepresentation: &brief,
	}); err != nil {
		log.Warn("keycloak status: read-users probe failed: ", err)
	} else {
		capabilities[keycloakRealmDTO.CapabilityKeycloakReadUsers] = true
	}

	groupMax := 1
	if _, err := KeycloakRealmSingleton.client.GetGroups(ctx, token.AccessToken, KeycloakRealmSingleton.Realm, gocloak.GetGroupsParams{
		First: &first,
		Max:   &groupMax,
	}); err != nil {
		log.Warn("keycloak status: read-groups probe failed: ", err)
	} else {
		capabilities[keycloakRealmDTO.CapabilityKeycloakReadGroups] = true
	}

	healthy := true
	for _, ok := range capabilities {
		if !ok {
			healthy = false
			break
		}
	}

	return keycloakRealmDTO.KeycloakStatus{
		Healthy:      healthy,
		Capabilities: capabilities,
	}
}

// SearchKeycloakUsers performs a realm-wide search by username/email/name.
// Authorization for the route is "is a lecturer of any course" - see the
// comment in router.go.
func SearchKeycloakUsers(ctx context.Context, query string, limit int) (keycloakRealmDTO.UserSearchResults, error) {
	q := strings.TrimSpace(query)
	if len(q) < 2 {
		return keycloakRealmDTO.UserSearchResults{}, ErrInvalidQuery
	}

	if limit <= 0 || limit > maxSearchResults {
		limit = maxSearchResults
	}

	token, err := LoginClient(ctx)
	if err != nil {
		return keycloakRealmDTO.UserSearchResults{}, err
	}

	// Fetch one more than the caller asked for so we can tell "exactly N matches"
	// from "more than N matches" without an extra round-trip.
	fetchMax := limit + 1
	brief := true
	users, err := KeycloakRealmSingleton.client.GetUsers(ctx, token.AccessToken, KeycloakRealmSingleton.Realm, gocloak.GetUsersParams{
		Search:              &q,
		Max:                 &fetchMax,
		BriefRepresentation: &brief,
	})
	if err != nil {
		log.Error("failed to search keycloak users: ", err)
		return keycloakRealmDTO.UserSearchResults{}, fmt.Errorf("failed to search keycloak users: %w", err)
	}

	truncated := len(users) > limit
	if truncated {
		users = users[:limit]
	}

	results := make([]keycloakRealmDTO.StaffMember, 0, len(users))
	for _, u := range users {
		results = append(results, keycloakRealmDTO.GetStaffMemberFromKeycloakUser(u))
	}

	return keycloakRealmDTO.UserSearchResults{
		Results:   results,
		Truncated: truncated,
	}, nil
}

func AddStudentsToEditorGroup(ctx context.Context, courseID uuid.UUID, studentIDs []uuid.UUID) (keycloakRealmDTO.AddStudentsToGroupResponse, error) {
	// 1. Log into keycloak
	token, err := LoginClient(ctx)
	if err != nil {
		log.Error("Failed to login to keycloak: ", err)
		return keycloakRealmDTO.AddStudentsToGroupResponse{}, err
	}

	// 2. Get Course Group
	editorGroup, err := GetCourseEditorGroup(ctx, token.AccessToken, courseID)
	if err != nil {
		log.Error("Failed to get course group: ", err)
		return keycloakRealmDTO.AddStudentsToGroupResponse{}, errors.New("failed to get course group")
	}

	// 3. Get the keycloak userIDs of the students
	succeededStudents, failedStudentIDs, err := AddStudentIDsToKeycloakGroup(ctx, token.AccessToken, studentIDs, *editorGroup.ID)
	// some error occurred before adding students to group
	if err != nil {
		log.Error("Failed to add students to group: ", err)
		return keycloakRealmDTO.AddStudentsToGroupResponse{}, errors.New("failed to add students to group")
	}

	response := keycloakRealmDTO.AddStudentsToGroupResponse{
		SucceededToAddStudentIDs: succeededStudents,
		FailedToAddStudentIDs:    failedStudentIDs,
	}
	return response, nil
}
