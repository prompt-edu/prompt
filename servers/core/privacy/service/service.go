package service

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdk "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/applicationAdministration"
	"github.com/prompt-edu/prompt/servers/core/coursePhaseType/coursePhaseTypeDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
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

type PrivacyService struct {
	queries          db.Queries
	conn             *pgxpool.Pool
	applications     ApplicationDataProvider
	subjects         SubjectIdentifierResolver
	coursePhaseTypes CoursePhaseTypeProvider
}

func NewPrivacyService(queries db.Queries, conn *pgxpool.Pool, applications ApplicationDataProvider, subjects SubjectIdentifierResolver, coursePhaseTypes CoursePhaseTypeProvider) *PrivacyService {
	return &PrivacyService{
		queries:          queries,
		conn:             conn,
		applications:     applications,
		subjects:         subjects,
		coursePhaseTypes: coursePhaseTypes,
	}
}
