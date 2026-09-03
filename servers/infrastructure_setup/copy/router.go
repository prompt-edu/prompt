package copy

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
)

// RegisterRoutes mounts the SDK copy endpoint. The group is not phase-scoped, so the
// auth middleware is applied per route here.
func RegisterRoutes(rg *gin.RouterGroup, svc *Service, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	promptTypes.RegisterCopyEndpoint(rg, authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), svc)
}
