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

// ConfigHandler implements the PhaseConfigHandler interface for core config status checks.
type ConfigHandler struct{}

func (h *ConfigHandler) HandlePhaseConfig(c *gin.Context) (map[string]bool, error) {
	if CopyServiceSingleton == nil {
		return map[string]bool{
			"sourcePhase": false,
			"semesterTag": false,
		}, nil
	}

	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		return nil, err
	}

	cfg, err := CopyServiceSingleton.queries.GetCoursePhaseConfig(c.Request.Context(), coursePhaseID)
	if err != nil {
		return map[string]bool{
			"sourcePhase": false,
			"semesterTag": false,
		}, nil
	}

	return map[string]bool{
		"sourcePhase": cfg.TeamSourceCoursePhaseID != nil || cfg.StudentSourceCoursePhaseID != nil,
		"semesterTag": cfg.SemesterTag != "",
	}, nil
}

// InitCopyModule registers copy routes and initialises the singleton.
func InitCopyModule(routerGroup *gin.RouterGroup, conn *pgxpool.Pool) {
	promptTypes.RegisterCopyEndpoint(routerGroup, promptSDK.AuthenticationMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), &InfrastructureSetupCopyHandler{})
	promptTypes.RegisterConfigEndpoint(routerGroup, promptSDK.AuthenticationMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), &ConfigHandler{})
	CopyServiceSingleton = &CopyService{
		queries: db.New(conn),
		conn:    conn,
	}
}
