package coursePhaseDeletion

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
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

// HandleCoursePhaseDeletion removes every row this service stores for the given course phase.
// Deleting the teams and skills cascades to the student responses, allocations and tutors, see
// db/query/coursePhaseDeletion.sql. The handler is idempotent: deleting a course phase without any
// stored data succeeds.
func (s *CoursePhaseDeletionService) HandleCoursePhaseDeletion(c *gin.Context, coursePhaseID uuid.UUID) error {
	// Deleting the teams takes row locks that can block behind a concurrent transaction, so bound
	// the work rather than holding a pool connection for as long as that transaction runs.
	ctx, cancel := db.GetTimeoutContext(c)
	defer cancel()

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for deleting course phase %s: %w", coursePhaseID, err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	if err := qtx.DeleteTeamsByCoursePhase(ctx, coursePhaseID); err != nil {
		return fmt.Errorf("failed to delete teams for course phase %s: %w", coursePhaseID, err)
	}
	if err := qtx.DeleteSkillsByCoursePhase(ctx, coursePhaseID); err != nil {
		return fmt.Errorf("failed to delete skills for course phase %s: %w", coursePhaseID, err)
	}
	if err := qtx.DeleteSurveyTimeframeByCoursePhase(ctx, coursePhaseID); err != nil {
		return fmt.Errorf("failed to delete survey timeframe for course phase %s: %w", coursePhaseID, err)
	}
	if err := qtx.DeleteTeaseWorkspaceByCoursePhase(ctx, coursePhaseID); err != nil {
		return fmt.Errorf("failed to delete tease workspace for course phase %s: %w", coursePhaseID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit deletion of course phase %s: %w", coursePhaseID, err)
	}
	return nil
}
