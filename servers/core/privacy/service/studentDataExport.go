package service

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	sdk "github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/core/applicationAdministration"
	"github.com/prompt-edu/prompt/servers/core/instructorNote"
	"github.com/prompt-edu/prompt/servers/core/student"
)

func GetSubjectData(c *gin.Context, subjectIdentifiers sdk.SubjectIdentifiers) []any {

  res := []any{}

  getSubjectDataForUser(c, subjectIdentifiers.UserID, &res)

  if (subjectIdentifiers.StudentID != uuid.Nil) {
    getSubjectDataForStudent(c, subjectIdentifiers.StudentID, subjectIdentifiers.CourseParticipationIDs, &res)
  }

  return res
}

func getSubjectDataForUser(c *gin.Context, userUUID uuid.UUID, res *[]any ) {

  addToResult("Instructor Notes as Author", func () (any, error) { 
    return instructorNote.GetStudentNotesForAuthor(c, userUUID)
  }, res)

}

func getSubjectDataForStudent(c *gin.Context, studentUUID uuid.UUID, courseParticipationUUIDs []uuid.UUID, res *[]any) {

  addToResult("Student record", func () (any, error) { 
    return student.GetStudentByID(c, studentUUID)
  }, res)

  addToResult("Enrollments", func () (any, error) { 
    return student.GetStudentEnrollmentsByID(c, studentUUID)
  }, res)

  addToResult("Instructor Notes as Receiver", func () (any, error) {
    return instructorNote.GetStudentNotesByIDWithoutAuthor(c, studentUUID)
  }, res)

  addToResult("Application Data", func () (any, error) {
    return applicationAdministration.GetAllApplicationAnswers(c, courseParticipationUUIDs)
  }, res)

}

type DataSourceResult struct {
	Name  string
	Value any
	Error string
}

func addToResult(name string, serviceMethod func() (any, error), res *[]any) {
  data, err := serviceMethod()
  if err != nil { 
    *res = append(*res, DataSourceResult{
      Name: name,
      Error: err.Error(),
    })
    return
  }
  *res = append(*res, DataSourceResult{
    Name: name,
    Value: data,
  })
}


