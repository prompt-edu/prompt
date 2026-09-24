package copy

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

// RegisterRoutes mounts the standardized POST /copy endpoint. Core calls it on the service's api
// group, outside :coursePhaseID, with the source and target phase in the request body.
func RegisterRoutes(routerGroup *gin.RouterGroup, service *CopyService) {
	promptTypes.RegisterCopyModule(routerGroup.Group("", audit.Describe(auditCopyAction)),
		service, promptSDK.PromptAdmin, promptSDK.CourseLecturer)
}
