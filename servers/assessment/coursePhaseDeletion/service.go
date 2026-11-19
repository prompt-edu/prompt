package coursePhaseDeletion

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

// deletionTimeout bounds the transaction: the deletes take row locks that can block behind a
// concurrent transaction, and the request should not hold a pool connection for that long.
const deletionTimeout = 10 * time.Second

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

// HandleCoursePhaseDeletion removes every row this service stores for the given course phase in
// one transaction. Schemas the phase owns are handed to another phase that still uses them and
// deleted otherwise, see db/query/coursePhaseDeletion.sql. Global templates are left alone. The
// handler is idempotent: deleting a course phase without any stored data succeeds.
func (s *CoursePhaseDeletionService) HandleCoursePhaseDeletion(c *gin.Context, coursePhaseID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(c.Request.Context(), deletionTimeout)
	defer cancel()

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for deleting course phase %s: %w", coursePhaseID, err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	steps := []struct {
		action string
		run    func(context.Context, uuid.UUID) error
	}{
		{"delete assessments", qtx.DeleteAssessmentsByCoursePhase},
		{"delete independent assessments", qtx.DeleteIndependentAssessmentsByCoursePhase},
		{"delete category assessments", qtx.DeleteCategoryAssessmentsByCoursePhase},
		{"delete assessment completions", qtx.DeleteAssessmentCompletionsByCoursePhase},
		{"delete evaluations", qtx.DeleteEvaluationsByCoursePhase},
		{"delete evaluation completions", qtx.DeleteEvaluationCompletionsByCoursePhase},
		{"delete feedback items", qtx.DeleteFeedbackItemsByCoursePhase},
		{"delete action items", qtx.DeleteActionItemsByCoursePhase},
		{"delete evaluation reminders", qtx.DeleteEvaluationRemindersByCoursePhase},
		{"delete results released mails", qtx.DeleteResultsReleasedMailsByCoursePhase},
		{"delete config", qtx.DeleteCoursePhaseConfigByCoursePhase},
		// The schema steps rely on the phase's own rows being gone, so they must stay last.
		{"transfer owned assessment schemas", qtx.TransferOwnedAssessmentSchemas},
		{"delete owned assessment schemas", qtx.DeleteAssessmentSchemasOwnedByCoursePhase},
	}
	for _, step := range steps {
		if err := step.run(ctx, coursePhaseID); err != nil {
			return fmt.Errorf("failed to %s of course phase %s: %w", step.action, coursePhaseID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit deletion of course phase %s: %w", coursePhaseID, err)
	}
	return nil
}
