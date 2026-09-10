package copy

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
)

// RegisterRoutes mounts the SDK copy module. The group is not phase-scoped, so the
// module applies the auth middleware per route rather than on the group.
func RegisterRoutes(rg *gin.RouterGroup, svc *Service) {
	promptTypes.RegisterCopyModule(rg, svc, promptSDK.PromptAdmin, promptSDK.CourseLecturer)
}
