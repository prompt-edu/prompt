package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
)

type PrivacyService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

func NewPrivacyService(queries db.Queries, conn *pgxpool.Pool) *PrivacyService {
	return &PrivacyService{
		queries: queries,
		conn:    conn,
	}
}

func (s *PrivacyService) DataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	exp.AddJSON("Team Assignment", "self_team_allocation.json", func() (any, error) {
		return s.queries.GetAssignmentsByParticipationIDs(c, subject.CourseParticipationIDs)
	})
	exp.AddJSON("Tutor Assignment", "tutor.json", func() (any, error) {
		return s.queries.GetTutorsByCourseParticipationIDs(c, tutorIDs(subject))
	})
	return nil
}

func (s *PrivacyService) DataDeletionHandler(c *gin.Context, subject sdkAuth.SubjectIdentifiers) error {
	tx, err := s.conn.Begin(c)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, c)
	qtx := s.queries.WithTx(tx)

	if err := qtx.DeleteAssignmentsByCourseParticipationIDs(c, subject.CourseParticipationIDs); err != nil {
		return err
	}
	if err := qtx.DeleteTutorsByCourseParticipationIDs(c, tutorIDs(subject)); err != nil {
		return err
	}

	return tx.Commit(c)
}

// tutorIDs lists the ids a tutor row of the subject can carry: a tutor imported From Course is
// stored under their student id, a course lecturer or editor under their Keycloak user id, and
// older rows under a course participation id. A re-import keeps the id a tutor already has, so a
// row can carry any of them whatever its source.
func tutorIDs(subject sdkAuth.SubjectIdentifiers) []uuid.UUID {
	ids := append([]uuid.UUID{}, subject.CourseParticipationIDs...)
	if subject.StudentID != uuid.Nil {
		ids = append(ids, subject.StudentID)
	}
	if subject.UserID != uuid.Nil {
		ids = append(ids, subject.UserID)
	}
	return ids
}
