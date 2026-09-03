package service

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	sdk "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/utils"
)

func (s *PrivacyService) AggregateSubjectDataFromCore(ctx context.Context, doc ServiceExportRequest, subjectIdentifiers sdk.SubjectIdentifiers) (err error) {
	defer func() { s.UpdateExportDocStatus(err, context.WithoutCancel(ctx), doc.ExportDoc.ID) }()

	ex, err := utils.NewExport()
	if err != nil {
		return
	}

	defer ex.Close()

	if subjectIdentifiers.UserID != uuid.Nil {
		s.getSubjectDataForUser(ctx, ex, subjectIdentifiers.UserID)
	}

	if subjectIdentifiers.StudentID != uuid.Nil {
		s.getSubjectDataForStudent(ctx, ex, subjectIdentifiers.StudentID, subjectIdentifiers.CourseParticipationIDs)
	}

	err = ex.UploadTo(ctx, doc.PresignedUploadURL)
	return
}

func (s *PrivacyService) getSubjectDataForUser(ctx context.Context, ex *utils.Export, userUUID uuid.UUID) {

	ex.AddJSON("Instructor Notes as Author", "user/instructor_notes.json", func() (any, error) {
		return s.instructorNotes.GetStudentNotesForAuthorWithoutStudent(ctx, userUUID)
	})

}

func (s *PrivacyService) getSubjectDataForStudent(ctx context.Context, ex *utils.Export, studentUUID uuid.UUID, courseParticipationUUIDs []uuid.UUID) {

	ex.AddJSON("Student record", "student/student_record.json", func() (any, error) {
		return s.students.GetStudentByID(ctx, studentUUID)
	})

	ex.AddJSON("Enrollments", "student/enrollments.json", func() (any, error) {
		return s.students.GetStudentEnrollmentsByID(ctx, studentUUID)
	})

	ex.AddJSON("Instructor Notes as Receiver", "student/instructor_notes.json", func() (any, error) {
		return s.instructorNotes.GetStudentNotesByIDWithoutAuthor(ctx, studentUUID)
	})

	ex.AddJSON("Application Data", "student/application.json", func() (any, error) {
		return s.applications.GetAllApplicationAnswers(ctx, courseParticipationUUIDs)
	})

	s.addApplicationFiles(ctx, ex, courseParticipationUUIDs)

}

func (s *PrivacyService) addApplicationFiles(ctx context.Context, ex *utils.Export, courseParticipationUUIDs []uuid.UUID) {
	for _, answer := range s.applications.GetApplicationFileUploadAnswersWithFileRecord(ctx, courseParticipationUUIDs) {
		fileID := answer.FileID
		questionTitle := answer.QuestionTitle
		ex.AddFile(
			fmt.Sprintf("Application File: %s-%s", answer.FileID, questionTitle),
			fmt.Sprintf("student/application_files/%s", MakeUniqueFileNameWithEnding(answer)),
			func() (io.Reader, error) {
				reader, _, err := s.applicationFiles.DownloadFile(ctx, fileID)
				return reader, err
			},
		)
	}
}
