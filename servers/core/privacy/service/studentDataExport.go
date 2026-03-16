package service

import (
	"github.com/google/uuid"
	sdk "github.com/prompt-edu/prompt-sdk/promptTypes"
)

func GetSubjectData(subjectIdentifiers sdk.SubjectIdentifiers) {

  getSubjectDataForUser(subjectIdentifiers.UserID)

  if (subjectIdentifiers.StudentID != uuid.Nil) {
    getSubjectDataForStudent(subjectIdentifiers.StudentID, subjectIdentifiers.CourseParticipationIDs)
  }
}

func getSubjectDataForUser(userUUID uuid.UUID) {

}
func getSubjectDataForStudent(studentUUID uuid.UUID, courseParticipationUUIDs []uuid.UUID) {

}



