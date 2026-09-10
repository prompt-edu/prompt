package config

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

func RegisterRoutes(routerGroup *gin.RouterGroup) {
	promptTypes.RegisterConfigModule(routerGroup, &configHandler{}, promptSDK.PromptAdmin, promptSDK.CourseLecturer)
}
