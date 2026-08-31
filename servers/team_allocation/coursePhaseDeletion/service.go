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
	tx, err := s.conn.Begin(c)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for deleting course phase %s: %w", coursePhaseID, err)
	}
	defer promptSDK.DeferDBRollback(tx, c)
	qtx := s.queries.WithTx(tx)

	if err := qtx.DeleteTeamsByCoursePhase(c, coursePhaseID); err != nil {
		return fmt.Errorf("failed to delete teams for course phase %s: %w", coursePhaseID, err)
	}
	if err := qtx.DeleteSkillsByCoursePhase(c, coursePhaseID); err != nil {
		return fmt.Errorf("failed to delete skills for course phase %s: %w", coursePhaseID, err)
	}
	if err := qtx.DeleteSurveyTimeframeByCoursePhase(c, coursePhaseID); err != nil {
		return fmt.Errorf("failed to delete survey timeframe for course phase %s: %w", coursePhaseID, err)
	}
	if err := qtx.DeleteTeaseWorkspaceByCoursePhase(c, coursePhaseID); err != nil {
		return fmt.Errorf("failed to delete tease workspace for course phase %s: %w", coursePhaseID, err)
	}

	if err := tx.Commit(c); err != nil {
		return fmt.Errorf("failed to commit deletion of course phase %s: %w", coursePhaseID, err)
	}
	return nil
}
