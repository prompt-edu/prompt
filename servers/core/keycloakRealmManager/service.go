package keycloakRealmManager

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/keycloakRealmManager/keycloakRealmDTO"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
	log "github.com/sirupsen/logrus"
)

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
