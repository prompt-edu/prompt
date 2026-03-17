package copy

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

// setupCopyRouter sets up phase copy endpoints.
// @Summary Phase Copy Endpoints
// @Description Copy course phase configuration between phases.
// @Tags copy
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body promptTypes.PhaseCopyRequest true "Phase copy payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /copy [post]
func setupCopyRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	promptTypes.RegisterCopyEndpoint(routerGroup, authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), &AssessmentCopyHandler{})
}
