package config

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
)

// RegisterRoutes mounts the SDK config module. The SDK registers it as a bare
// /config, so the group has to be phase-scoped for :coursePhaseID to resolve; the
// module applies route-level auth, so the group must not already carry one.
func RegisterRoutes(rg *gin.RouterGroup, svc *Service) {
	promptTypes.RegisterConfigModule(rg, svc, promptSDK.PromptAdmin, promptSDK.CourseLecturer)
}
