package config

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

func RegisterRoutes(routerGroup *gin.RouterGroup, service *ConfigService) {
	promptTypes.RegisterConfigModule(routerGroup, service, promptSDK.PromptAdmin, promptSDK.CourseLecturer)
}
