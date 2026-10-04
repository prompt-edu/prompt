package copy

import (
	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

func RegisterRoutes(routerGroup *gin.RouterGroup) {
	promptTypes.RegisterCopyModule(routerGroup, &InterviewCopyHandler{}, promptSDK.PromptAdmin, promptSDK.CourseLecturer)
}
