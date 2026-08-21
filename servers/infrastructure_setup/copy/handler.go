// Package copy implements the PhaseCopyHandler interface for the infrastructure setup phase.
package copy

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
	log "github.com/sirupsen/logrus"
)

// CopyService handles phase-level data duplication.
type CopyService struct {
	queries *db.Queries
	conn    *pgxpool.Pool
}

var CopyServiceSingleton *CopyService

// InfrastructureSetupCopyHandler implements the PhaseCopyHandler interface.
type InfrastructureSetupCopyHandler struct{}

// HandlePhaseCopy copies the phase config, provider config stubs (without credentials)
// and resource configs from a source phase to a target phase. Everything happens in one
// transaction so a failure part-way cannot leave a half-copied phase.
func (h *InfrastructureSetupCopyHandler) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
	if CopyServiceSingleton == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "copy service not initialized"})
		return nil
	}

	if err := CopyServiceSingleton.copyPhase(c.Request.Context(), req); err != nil {
		log.WithError(err).Error("copy infrastructure setup phase")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "phase data copied successfully"})
	return nil
}

func (s *CopyService) copyPhase(ctx context.Context, req promptTypes.PhaseCopyRequest) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	// The config row must exist on the target: HandlePhaseConfig reports nothing
	// configured at all when it is missing.
	if err := qtx.CopyCoursePhaseConfig(ctx, db.CopyCoursePhaseConfigParams{
		SourceCoursePhaseID: req.SourceCoursePhaseID,
		TargetCoursePhaseID: req.TargetCoursePhaseID,
	}); err != nil {
		return fmt.Errorf("copy course phase config: %w", err)
	}
	if err := qtx.CopyProviderConfigsWithEmptyCredentials(ctx, db.CopyProviderConfigsWithEmptyCredentialsParams{
		SourceCoursePhaseID: req.SourceCoursePhaseID,
		TargetCoursePhaseID: req.TargetCoursePhaseID,
	}); err != nil {
		return fmt.Errorf("copy provider configs: %w", err)
	}
	if err := qtx.CopyResourceConfigs(ctx, db.CopyResourceConfigsParams{
		SourceCoursePhaseID: req.SourceCoursePhaseID,
		TargetCoursePhaseID: req.TargetCoursePhaseID,
	}); err != nil {
		return fmt.Errorf("copy resource configs: %w", err)
	}

	return tx.Commit(ctx)
}

// ConfigHandler implements the PhaseConfigHandler interface for core config status checks.
// Upstream wiring (teams, team allocation) is owned by the phase configurator and
// surfaced by core, so this handler only reports phase-local readiness.
type ConfigHandler struct{}

func (h *ConfigHandler) HandlePhaseConfig(c *gin.Context) (map[string]bool, error) {
	empty := map[string]bool{
		"semesterTag":    false,
		"providerConfig": false,
		"resourceConfig": false,
	}
	if CopyServiceSingleton == nil {
		return empty, nil
	}

	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		return nil, err
	}

	ctx := c.Request.Context()
	cfg, err := CopyServiceSingleton.queries.GetCoursePhaseConfig(ctx, coursePhaseID)
	if err != nil {
		// No config row yet means nothing is configured but is not an error.
		return empty, nil
	}

	// Only providers holding credentials count: a copied phase carries provider rows
	// with empty credentials, which cannot provision anything.
	providers, err := CopyServiceSingleton.queries.ListConfiguredProviderConfigs(ctx, coursePhaseID)
	if err != nil {
		return empty, nil
	}
	resources, err := CopyServiceSingleton.queries.ListResourceConfigs(ctx, coursePhaseID)
	if err != nil {
		return empty, nil
	}

	return map[string]bool{
		"semesterTag":    cfg.SemesterTag != "",
		"providerConfig": len(providers) > 0,
		"resourceConfig": len(resources) > 0,
	}, nil
}

// InitCopyModule registers the copy and config routes and initialises the singleton.
// The config endpoint is registered as a bare /config by the SDK, so it must be mounted
// on the phase-scoped group for :coursePhaseID to resolve.
func InitCopyModule(copyGroup, coursePhaseGroup *gin.RouterGroup, conn *pgxpool.Pool) {
	promptTypes.RegisterCopyEndpoint(copyGroup, promptSDK.AuthenticationMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), &InfrastructureSetupCopyHandler{})
	promptTypes.RegisterConfigEndpoint(coursePhaseGroup, promptSDK.AuthenticationMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), &ConfigHandler{})
	CopyServiceSingleton = &CopyService{
		queries: db.New(conn),
		conn:    conn,
	}
}
