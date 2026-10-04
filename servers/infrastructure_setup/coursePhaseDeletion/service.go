// Package coursePhaseDeletion implements the SDK CoursePhaseDeletionHandler for the
// infrastructure setup phase.
package coursePhaseDeletion

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

// deletionTimeout bounds the deletion. The deletes take row locks that can queue behind
// a running execution, and holding a pool connection for that long would starve the
// endpoints a lecturer is using.
const deletionTimeout = 30 * time.Second

// CoursePhaseDeletionService removes everything the service stores for a course phase.
type CoursePhaseDeletionService struct {
	queries *db.Queries
	conn    *pgxpool.Pool
}

// NewCoursePhaseDeletionService creates a CoursePhaseDeletionService.
func NewCoursePhaseDeletionService(pool *pgxpool.Pool) *CoursePhaseDeletionService {
	return &CoursePhaseDeletionService{queries: db.New(pool), conn: pool}
}

// HandleCoursePhaseDeletion implements promptTypes.CoursePhaseDeletionHandler.
//
// It matters more here than for most phases: the provider configs hold the encrypted
// credentials of external systems, which would otherwise stay in this database for a
// phase that no longer exists. The external resources are not touched - this phase never
// deletes them, so a group it provisioned outlives the phase and has to be removed in
// the provider by hand.
//
// The handler is idempotent: deleting a phase this service stores nothing for succeeds.
func (s *CoursePhaseDeletionService) HandleCoursePhaseDeletion(c *gin.Context, coursePhaseID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(c.Request.Context(), deletionTimeout)
	defer cancel()

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction for deleting course phase %s: %w", coursePhaseID, err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	if err := qtx.DeleteResourceInstancesByCoursePhase(ctx, coursePhaseID); err != nil {
		return fmt.Errorf("delete resource instances of course phase %s: %w", coursePhaseID, err)
	}
	if err := qtx.DeleteResourceConfigsByCoursePhase(ctx, coursePhaseID); err != nil {
		return fmt.Errorf("delete resource configs of course phase %s: %w", coursePhaseID, err)
	}
	if err := qtx.DeleteProviderConfigsByCoursePhase(ctx, coursePhaseID); err != nil {
		return fmt.Errorf("delete provider configs of course phase %s: %w", coursePhaseID, err)
	}
	if err := qtx.DeleteCoursePhaseConfigByCoursePhase(ctx, coursePhaseID); err != nil {
		return fmt.Errorf("delete phase config of course phase %s: %w", coursePhaseID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit deletion of course phase %s: %w", coursePhaseID, err)
	}
	return nil
}
