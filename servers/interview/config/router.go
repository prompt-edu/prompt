package config

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

func setupConfigRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	promptTypes.RegisterConfigEndpoint(routerGroup, authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), &ConfigHandler{})
}
