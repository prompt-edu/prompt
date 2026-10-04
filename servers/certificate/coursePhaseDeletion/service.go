package coursePhaseDeletion

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
)

type CoursePhaseDeletionService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

func NewCoursePhaseDeletionService(queries db.Queries, conn *pgxpool.Pool) *CoursePhaseDeletionService {
	return &CoursePhaseDeletionService{
		queries: queries,
		conn:    conn,
	}
}

// HandleCoursePhaseDeletion removes every row this service stores for the given course phase: the
// certificate downloads and the phase config with its template. The handler is idempotent:
// deleting a course phase without any stored data succeeds.
func (s *CoursePhaseDeletionService) HandleCoursePhaseDeletion(c *gin.Context, coursePhaseID uuid.UUID) error {
	ctx := c.Request.Context()

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for deleting course phase %s: %w", coursePhaseID, err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	if err := qtx.DeleteCertificateDownloadsByCoursePhase(ctx, coursePhaseID); err != nil {
		return fmt.Errorf("failed to delete certificate downloads for course phase %s: %w", coursePhaseID, err)
	}
	if err := qtx.DeleteCoursePhaseConfigByCoursePhase(ctx, coursePhaseID); err != nil {
		return fmt.Errorf("failed to delete certificate config for course phase %s: %w", coursePhaseID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit deletion of course phase %s: %w", coursePhaseID, err)
	}
	return nil
}
