package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	sdk "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/course/courseParticipation"
	"github.com/prompt-edu/prompt/servers/core/student"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

func GetSubjectIdentifiers(ctx *gin.Context) (sdk.SubjectIdentifiers, error) {
	userID, errUserUUID := utils.GetUserUUIDFromContext(ctx)
	if errUserUUID != nil {
		return sdk.SubjectIdentifiers{}, errUserUUID
	}

	studentID, errStudentUUID := getStudentID(ctx)

	if errors.Is(errStudentUUID, sql.ErrNoRows) {
		return sdk.SubjectIdentifiers{UserID: userID}, nil
	}

	if errStudentUUID != nil {
		return sdk.SubjectIdentifiers{}, errStudentUUID
	}

	courseParticipationIDs, errCourseParts := getCourseParticipations(ctx, studentID)
	if errCourseParts != nil {
		return sdk.SubjectIdentifiers{}, errCourseParts
	}

	return sdk.SubjectIdentifiers{
		UserID:                 userID,
		StudentID:              studentID,
		CourseParticipationIDs: courseParticipationIDs,
	}, nil
}

func AssembleSubjectIdentifiers(ctx context.Context, userID uuid.UUID, studentID *uuid.UUID) (sdk.SubjectIdentifiers, error) {
	identifiers := sdk.SubjectIdentifiers{UserID: userID}
	if studentID == nil {
		return identifiers, nil
	}
	identifiers.StudentID = *studentID

	cpIDs, err := getCourseParticipations(ctx, *studentID)
	if err != nil {
		return sdk.SubjectIdentifiers{}, err
	}
	identifiers.CourseParticipationIDs = cpIDs
	return identifiers, nil
}

func getStudentID(ctx *gin.Context) (uuid.UUID, error) {
	matrNr := utils.GetMatriculationNumberFromContext(ctx)
	universityLogin := utils.GetUniversityLoginFromContext(ctx)

	student, err := student.ResolveStudentByUniversityCredentials(ctx, &AuthServiceSingleton.queries, matrNr, universityLogin)
	if err != nil {
		return uuid.UUID{}, err
	}

	return student.ID, nil
}

func getCourseParticipations(ctx context.Context, studentID uuid.UUID) ([]uuid.UUID, error) {
	cps, err := courseParticipation.GetAllCourseParticipationsForStudent(ctx, studentID)
	if err != nil {
		return []uuid.UUID{}, err
	}

	uuids := []uuid.UUID{}

	for _, cp := range cps {
		uuids = append(uuids, cp.ID)
	}

	return uuids, nil
}
