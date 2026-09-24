package coursePhaseDeletion

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/presentation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/presentation/storage"
)

type CoursePhaseDeletionService struct {
	queries *db.Queries
	conn    *pgxpool.Pool
	storage storage.Adapter
}

func NewCoursePhaseDeletionService(queries *db.Queries, conn *pgxpool.Pool, storageAdapter storage.Adapter) *CoursePhaseDeletionService {
	return &CoursePhaseDeletionService{
		queries: queries,
		conn:    conn,
		storage: storageAdapter,
	}
}

// HandleCoursePhaseDeletion removes every row and every uploaded material this service stores
// for the given course phase. Deleting the presentations cascades to their materials, feedback
// forms, answers and contributors, and deleting the config cascades to the feedback categories,
// see db/query/coursePhaseDeletion.sql. The handler is idempotent: deleting a course phase without
// any stored data succeeds.
func (s *CoursePhaseDeletionService) HandleCoursePhaseDeletion(c *gin.Context, coursePhaseID uuid.UUID) error {
	ctx := c.Request.Context()

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for deleting course phase %s: %w", coursePhaseID, err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	// Presentations first: each holds its slot with ON DELETE RESTRICT, and their feedback
	// answers restrict deleting the categories the config cascades to.
	if err := qtx.DeletePresentationsByPhase(ctx, coursePhaseID); err != nil {
		return fmt.Errorf("failed to delete presentations for course phase %s: %w", coursePhaseID, err)
	}
	if err := qtx.DeletePresentationSlotsByPhase(ctx, coursePhaseID); err != nil {
		return fmt.Errorf("failed to delete presentation slots for course phase %s: %w", coursePhaseID, err)
	}
	if err := qtx.DeleteCoursePhaseConfig(ctx, coursePhaseID); err != nil {
		return fmt.Errorf("failed to delete config for course phase %s: %w", coursePhaseID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit deletion of course phase %s: %w", coursePhaseID, err)
	}

	// The objects go by prefix rather than by the keys of the rows just deleted: a retry after a
	// failure here finds no rows left, but still finds the objects. The prefix also covers uploads
	// that were presigned but never completed.
	if err := s.storage.DeletePrefix(ctx, storage.CoursePhasePrefix(coursePhaseID)); err != nil {
		return fmt.Errorf("failed to delete stored materials for course phase %s: %w", coursePhaseID, err)
	}
	return nil
}
