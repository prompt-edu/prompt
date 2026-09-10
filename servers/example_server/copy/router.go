package copy

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

func RegisterRoutes(routerGroup *gin.RouterGroup, service *CopyService) {
	promptTypes.RegisterCopyModule(routerGroup.Group("", audit.Describe(auditCopyAction)),
		service, promptSDK.PromptAdmin, promptSDK.CourseLecturer)
}
