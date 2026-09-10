// Package copy implements the SDK PhaseCopyHandler for the infrastructure setup phase.
package copy

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
	log "github.com/sirupsen/logrus"
)

// Service handles phase-level data duplication.
//
// It implements promptTypes.PhaseCopyHandler itself, so RegisterRoutes passes it
// straight to the SDK and there is no separate handler type to keep in sync.
type Service struct {
	queries *db.Queries
	conn    *pgxpool.Pool
}

// NewService creates a Service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{queries: db.New(pool), conn: pool}
}

// HandlePhaseCopy implements promptTypes.PhaseCopyHandler.
//
// The SDK's RegisterCopyEndpoint writes the HTTP response for both outcomes, so this
// only returns the error - writing one here too would put two JSON bodies on the wire.
func (s *Service) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
	if err := s.CopyPhase(c.Request.Context(), req.SourceCoursePhaseID, req.TargetCoursePhaseID); err != nil {
		log.WithError(err).Error("copy infrastructure setup phase")
		return err
	}
	return nil
}

// CopyPhase copies the phase config, provider config stubs (without credentials) and
// resource configs from a source phase to a target phase. Everything happens in one
// transaction so a failure part-way cannot leave a half-copied phase.
func (s *Service) CopyPhase(ctx context.Context, sourceCoursePhaseID, targetCoursePhaseID uuid.UUID) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	// The config row must exist on the target: the config endpoint reports nothing
	// configured at all when it is missing.
	if err := qtx.CopyCoursePhaseConfig(ctx, db.CopyCoursePhaseConfigParams{
		SourceCoursePhaseID: sourceCoursePhaseID,
		TargetCoursePhaseID: targetCoursePhaseID,
	}); err != nil {
		return fmt.Errorf("copy course phase config: %w", err)
	}
	if err := qtx.CopyProviderConfigsWithEmptyCredentials(ctx, db.CopyProviderConfigsWithEmptyCredentialsParams{
		SourceCoursePhaseID: sourceCoursePhaseID,
		TargetCoursePhaseID: targetCoursePhaseID,
	}); err != nil {
		return fmt.Errorf("copy provider configs: %w", err)
	}
	if err := qtx.CopyResourceConfigs(ctx, db.CopyResourceConfigsParams{
		SourceCoursePhaseID: sourceCoursePhaseID,
		TargetCoursePhaseID: targetCoursePhaseID,
	}); err != nil {
		return fmt.Errorf("copy resource configs: %w", err)
	}

	return tx.Commit(ctx)
}
