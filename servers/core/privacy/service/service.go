package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt/servers/core/applicationAdministration"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

// ApplicationDataProvider is the slice of the application administration service
// the privacy export needs.
type ApplicationDataProvider interface {
	GetAllApplicationAnswers(ctx context.Context, courseParticipationIDs []uuid.UUID) ([]applicationAdministration.ApplicationDataExportPerCourseParticipation, error)
	GetApplicationFileUploadAnswersWithFileRecord(ctx context.Context, courseParticipationIDs []uuid.UUID) []db.GetAllApplicationAnswersFileUploadWithFileRecordByCourseParticipationIDsRow
}

type PrivacyService struct {
	queries      db.Queries
	conn         *pgxpool.Pool
	applications ApplicationDataProvider
}

func NewPrivacyService(queries db.Queries, conn *pgxpool.Pool, applications ApplicationDataProvider) *PrivacyService {
	return &PrivacyService{
		queries:      queries,
		conn:         conn,
		applications: applications,
	}
}
