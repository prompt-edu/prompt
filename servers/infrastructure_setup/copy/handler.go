// Package copy implements the PhaseCopyHandler interface for the infrastructure setup phase.
package copy

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
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

// HandlePhaseCopy copies provider config stubs (without credentials) and resource
// configs from a source phase to a target phase.
func (h *InfrastructureSetupCopyHandler) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
	if CopyServiceSingleton == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "copy service not initialized"})
		return nil
	}
	ctx := c.Request.Context()

	if err := CopyServiceSingleton.queries.CopyProviderConfigsWithEmptyCredentials(ctx, req.SourceCoursePhaseID, req.TargetCoursePhaseID); err != nil {
		log.WithError(err).Error("copy provider configs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}
	if err := CopyServiceSingleton.queries.CopyResourceConfigs(ctx, req.SourceCoursePhaseID, req.TargetCoursePhaseID); err != nil {
		log.WithError(err).Error("copy resource configs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "phase data copied successfully"})
	return nil
}

// ConfigHandler implements the PhaseConfigHandler interface (empty config for this phase).
type ConfigHandler struct{}

func (h *ConfigHandler) GetPhaseConfig(_ context.Context, _ uuid.UUID) (json.RawMessage, error) {
	return json.RawMessage("{}"), nil
}

func (h *ConfigHandler) UpdatePhaseConfig(_ context.Context, _ uuid.UUID, _ json.RawMessage) error {
	return nil
}

// InitCopyModule registers copy routes and initialises the singleton.
func InitCopyModule(routerGroup *gin.RouterGroup, conn *pgxpool.Pool) {
	promptTypes.RegisterCopyEndpoint(routerGroup, promptSDK.AuthenticationMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), &InfrastructureSetupCopyHandler{})
	CopyServiceSingleton = &CopyService{
		queries: db.New(conn),
		conn:    conn,
	}
}

