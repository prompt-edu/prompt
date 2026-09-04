package config

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
)

// RegisterRoutes mounts the SDK config endpoint. The SDK registers it as a bare
// /config, so the group has to be phase-scoped for :coursePhaseID to resolve; it
// brings its own auth middleware, so the group must not already carry one.
func RegisterRoutes(rg *gin.RouterGroup, svc *Service, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	promptTypes.RegisterConfigEndpoint(rg, authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), svc)
}
