package service

import (
	"context"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdk "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/applicationAdministration"
	"github.com/prompt-edu/prompt/servers/core/coursePhaseType/coursePhaseTypeDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/instructorNote/instructorNoteDTO"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
)

// ApplicationDataProvider is the slice of the application administration service
// the privacy export needs.
type ApplicationDataProvider interface {
	GetAllApplicationAnswers(ctx context.Context, courseParticipationIDs []uuid.UUID) ([]applicationAdministration.ApplicationDataExportPerCourseParticipation, error)
	GetApplicationFileUploadAnswersWithFileRecord(ctx context.Context, courseParticipationIDs []uuid.UUID) []db.GetAllApplicationAnswersFileUploadWithFileRecordByCourseParticipationIDsRow
}

// SubjectIdentifierResolver resolves the identifiers of the data subject an
// export or deletion applies to.
type SubjectIdentifierResolver interface {
	GetSubjectIdentifiers(c *gin.Context) (sdk.SubjectIdentifiers, error)
	AssembleSubjectIdentifiers(ctx context.Context, userID uuid.UUID, studentID *uuid.UUID) (sdk.SubjectIdentifiers, error)
}

// CoursePhaseTypeProvider lists the phase types that have to be asked for the
// subject's data.
type CoursePhaseTypeProvider interface {
	GetCoursePhaseTypesForStudentCourses(ctx context.Context, studentID uuid.UUID) ([]coursePhaseTypeDTO.CoursePhaseType, error)
}

// StudentProvider reads the student record and enrollments an export contains.
type StudentProvider interface {
	GetStudentByID(ctx context.Context, id uuid.UUID) (studentDTO.Student, error)
	GetStudentEnrollmentsByID(ctx context.Context, id uuid.UUID) (studentDTO.StudentEnrollmentsDTO, error)
}

// InstructorNoteProvider reads the anonymised instructor notes an export contains.
type InstructorNoteProvider interface {
	GetStudentNotesForAuthorWithoutStudent(ctx context.Context, authorID uuid.UUID) ([]instructorNoteDTO.InstructorNote, error)
	GetStudentNotesByIDWithoutAuthor(ctx context.Context, id uuid.UUID) ([]instructorNoteDTO.InstructorNote, error)
}

// ApplicationFileProvider reads and deletes the files uploaded with an application.
type ApplicationFileProvider interface {
	DownloadFile(ctx context.Context, fileID uuid.UUID) (io.ReadCloser, string, error)
	DeleteFile(ctx context.Context, fileID uuid.UUID, hardDelete bool) error
}

// ExportStorageProvider stores the export ZIPs the privacy export produces.
type ExportStorageProvider interface {
	GetUploadURL(ctx context.Context, exportRequestID, serviceName string) (string, error)
	GetDownloadURL(ctx context.Context, objectKey string) (string, error)
	DeleteFile(ctx context.Context, objectKey string) error
	GetFileSize(ctx context.Context, objectKey string) (int64, error)
}

// DeletionMailer sends the confirmation mail of a completed deletion request.
type DeletionMailer interface {
	SendMail(recipientAddress, subject, htmlBody string) error
}

type PrivacyService struct {
	queries          db.Queries
	conn             *pgxpool.Pool
	applications     ApplicationDataProvider
	subjects         SubjectIdentifierResolver
	coursePhaseTypes CoursePhaseTypeProvider
	students         StudentProvider
	instructorNotes  InstructorNoteProvider
	applicationFiles ApplicationFileProvider
	exportStorage    ExportStorageProvider
	mailer           DeletionMailer
}

func NewPrivacyService(
	queries db.Queries,
	conn *pgxpool.Pool,
	applications ApplicationDataProvider,
	subjects SubjectIdentifierResolver,
	coursePhaseTypes CoursePhaseTypeProvider,
	students StudentProvider,
	instructorNotes InstructorNoteProvider,
	applicationFiles ApplicationFileProvider,
	exportStorage ExportStorageProvider,
	mailer DeletionMailer,
) *PrivacyService {
	return &PrivacyService{
		queries:          queries,
		conn:             conn,
		applications:     applications,
		subjects:         subjects,
		coursePhaseTypes: coursePhaseTypes,
		students:         students,
		instructorNotes:  instructorNotes,
		applicationFiles: applicationFiles,
		exportStorage:    exportStorage,
		mailer:           mailer,
	}
}
