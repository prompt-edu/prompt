// Package copy implements the PhaseCopyHandler interface for the infrastructure setup phase.
package copy

import (
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
}

var CopyServiceSingleton *CopyService

// InfrastructureSetupCopyHandler implements the PhaseCopyHandler interface.
type InfrastructureSetupCopyHandler struct{}

// HandlePhaseCopy copies provider config stubs (without credentials) and resource
// configs from a source phase to a target phase.
func (h *InfrastructureSetupCopyHandler) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
	if CopyServiceSingleton == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "copy service not initialized"})
		return nil
	}
	ctx := c.Request.Context()

	if err := CopyServiceSingleton.queries.CopyProviderConfigsWithEmptyCredentials(ctx, db.CopyProviderConfigsWithEmptyCredentialsParams{
		SourceCoursePhaseID: req.SourceCoursePhaseID,
		TargetCoursePhaseID: req.TargetCoursePhaseID,
	}); err != nil {
		log.WithError(err).Error("copy provider configs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}
	if err := CopyServiceSingleton.queries.CopyResourceConfigs(ctx, db.CopyResourceConfigsParams{
		SourceCoursePhaseID: req.SourceCoursePhaseID,
		TargetCoursePhaseID: req.TargetCoursePhaseID,
	}); err != nil {
		log.WithError(err).Error("copy resource configs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "phase data copied successfully"})
	return nil
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

	providers, err := CopyServiceSingleton.queries.ListProviderConfigs(ctx, coursePhaseID)
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

// InitCopyModule registers copy routes and initialises the singleton.
func InitCopyModule(routerGroup *gin.RouterGroup, conn *pgxpool.Pool) {
	promptTypes.RegisterCopyEndpoint(routerGroup, promptSDK.AuthenticationMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), &InfrastructureSetupCopyHandler{})
	promptTypes.RegisterConfigEndpoint(routerGroup, promptSDK.AuthenticationMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), &ConfigHandler{})
	CopyServiceSingleton = &CopyService{
		queries: db.New(conn),
	}
}
