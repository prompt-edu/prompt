package service

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	sdk "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/core/applicationAdministration"
	"github.com/prompt-edu/prompt/servers/core/instructorNote"
	"github.com/prompt-edu/prompt/servers/core/storage/files"
	"github.com/prompt-edu/prompt/servers/core/student"
)

func AggregateSubjectDataFromCore(ctx context.Context, doc ServiceExportRequest, subjectIdentifiers sdk.SubjectIdentifiers) (err error) {
	defer func() { UpdateExportDocStatus(err, context.WithoutCancel(ctx), doc.ExportDoc.ID) }()

	ex, err := utils.NewExport()
	if err != nil {
		return
	}

	defer ex.Close()

	if subjectIdentifiers.UserID != uuid.Nil {
		getSubjectDataForUser(ctx, ex, subjectIdentifiers.UserID)
	}

	if subjectIdentifiers.StudentID != uuid.Nil {
		getSubjectDataForStudent(ctx, ex, subjectIdentifiers.StudentID, subjectIdentifiers.CourseParticipationIDs)
	}

	err = ex.UploadTo(ctx, doc.PresignedUploadURL)
	return
}

func getSubjectDataForUser(ctx context.Context, ex *utils.Export, userUUID uuid.UUID) {

	ex.AddJSON("Instructor Notes as Author", "user/instructor_notes.json", func() (any, error) {
		return instructorNote.GetStudentNotesForAuthorWithoutStudent(ctx, userUUID)
	})

}

func getSubjectDataForStudent(ctx context.Context, ex *utils.Export, studentUUID uuid.UUID, courseParticipationUUIDs []uuid.UUID) {

	ex.AddJSON("Student record", "student/student_record.json", func() (any, error) {
		return student.GetStudentByID(ctx, studentUUID)
	})

	ex.AddJSON("Enrollments", "student/enrollments.json", func() (any, error) {
		return student.GetStudentEnrollmentsByID(ctx, studentUUID)
	})

	ex.AddJSON("Instructor Notes as Receiver", "student/instructor_notes.json", func() (any, error) {
		return instructorNote.GetStudentNotesByIDWithoutAuthor(ctx, studentUUID)
	})

	ex.AddJSON("Application Data", "student/application.json", func() (any, error) {
		return applicationAdministration.GetAllApplicationAnswers(ctx, courseParticipationUUIDs)
	})

	addApplicationFiles(ctx, ex, courseParticipationUUIDs)

}

func addApplicationFiles(ctx context.Context, ex *utils.Export, courseParticipationUUIDs []uuid.UUID) {
	for _, answer := range applicationAdministration.GetApplicationFileUploadAnswersWithFileRecord(ctx, courseParticipationUUIDs) {
		fileID := answer.FileID
		questionTitle := answer.QuestionTitle
		ex.AddFile(
			fmt.Sprintf("Application File: %s-%s", answer.FileID, questionTitle),
			fmt.Sprintf("student/application_files/%s", MakeUniqueFileNameWithEnding(answer)),
			func() (io.Reader, error) {
				reader, _, err := files.StorageServiceSingleton.DownloadFile(ctx, fileID)
				return reader, err
			},
		)
	}
}
