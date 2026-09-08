package copy

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

func RegisterRoutes(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	// RegisterCopyEndpoint takes a single middleware, so the label rides on a zero-length group.
	promptTypes.RegisterCopyEndpoint(routerGroup.Group("", audit.Describe(auditCopyAction)),
		authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), &selfTeamCopyHandler{})
}
