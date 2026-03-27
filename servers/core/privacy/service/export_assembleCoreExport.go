package service

import (
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	sdk "github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/core/applicationAdministration"
	"github.com/prompt-edu/prompt/servers/core/instructorNote"
	"github.com/prompt-edu/prompt/servers/core/storage/files"
	"github.com/prompt-edu/prompt/servers/core/student"
)

func AggregateSubjectDataFromCore(c *gin.Context, doc ServiceExportRequest, subjectIdentifiers sdk.SubjectIdentifiers) (err error) {
  defer func() { UpdateExportDocStatus(err, c, doc.ExportDoc.ID) }()

  ex, err := utils.NewExport()
  if err != nil { return; }

  defer ex.Close()

  getSubjectDataForUser(c, ex, subjectIdentifiers.UserID)

  if subjectIdentifiers.StudentID != uuid.Nil {
    getSubjectDataForStudent(c, ex, subjectIdentifiers.StudentID, subjectIdentifiers.CourseParticipationIDs)
  }

  err = ex.UploadTo(c, doc.PresignedUploadURL)
  return
}

func getSubjectDataForUser(c *gin.Context, ex *utils.Export, userUUID uuid.UUID) {

  ex.AddJSON("Instructor Notes as Author", "user/instructor_notes.json", func () (any, error) {
    return instructorNote.GetStudentNotesForAuthorWithoutStudent(c, userUUID)
  })

}

func getSubjectDataForStudent(c *gin.Context, ex *utils.Export, studentUUID uuid.UUID, courseParticipationUUIDs []uuid.UUID) {

  ex.AddJSON("Student record", "student/student_record.json", func () (any, error) {
    return student.GetStudentByID(c, studentUUID)
  })

  ex.AddJSON("Enrollments", "student/enrollments.json", func () (any, error) {
    return student.GetStudentEnrollmentsByID(c, studentUUID)
  })

  ex.AddJSON("Instructor Notes as Receiver", "student/instructor_notes.json", func () (any, error) {
    return instructorNote.GetStudentNotesByIDWithoutAuthor(c, studentUUID)
  })

  ex.AddJSON("Application Data", "student/application.json", func () (any, error) {
    return applicationAdministration.GetAllApplicationAnswers(c, courseParticipationUUIDs)
  })

  addApplicationFiles(c, ex, courseParticipationUUIDs)

}

func addApplicationFiles(c *gin.Context, ex *utils.Export, courseParticipationUUIDs []uuid.UUID) {
  for _, answer := range applicationAdministration.GetApplicationFileUploadAnswersWithFileRecord(c, courseParticipationUUIDs) {
    fileID := answer.FileID
    questionTitle := answer.QuestionTitle
    ex.AddFile(
      fmt.Sprintf("Application File: %s-%s", answer.FileID, questionTitle),
      fmt.Sprintf("student/application_files/%s", MakeUniqueFileNameWithEnding(answer)),
      func() (io.Reader, error) {
        reader, _, err := files.StorageServiceSingleton.DownloadFile(c, fileID)
        return reader, err
      },
    )
  }
}

