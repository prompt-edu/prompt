package copy

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

// RegisterRoutes sets up phase copy endpoints.
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
func RegisterRoutes(routerGroup *gin.RouterGroup, service *CopyService, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	// RegisterCopyEndpoint takes a single middleware, so the label rides on a zero-length group.
	promptTypes.RegisterCopyEndpoint(routerGroup.Group("", audit.Describe(auditCopyAction)), authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service)
}
