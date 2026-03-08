package copy

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

func setupCopyRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	promptTypes.RegisterCopyEndpoint(routerGroup, authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), &IntroCourseCopyHandler{})
}
