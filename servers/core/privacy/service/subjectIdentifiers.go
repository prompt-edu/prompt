package service

import (
	"database/sql"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	sdk "github.com/prompt-edu/prompt-sdk/promptTypes"
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
		courseParticipationIDs = []uuid.UUID{}
	}

	return sdk.SubjectIdentifiers{
		UserID:                 userID,
		StudentID:              studentID,
		CourseParticipationIDs: courseParticipationIDs,
	}, nil
}

func getStudentID(ctx *gin.Context) (uuid.UUID, error) {
	matrNr := utils.GetMatriculationNumberFromContext(ctx)
	universityLogin := utils.GetUniversityLoginFromContext(ctx)

	student, err := student.ResolveStudentByUniversityCredentials(ctx, &PrivacyServiceSingleton.queries, matrNr, universityLogin)
	if err != nil {
		return uuid.UUID{}, err
	}

	return student.ID, nil
}

func getCourseParticipations(ctx *gin.Context, studentID uuid.UUID) ([]uuid.UUID, error) {
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
