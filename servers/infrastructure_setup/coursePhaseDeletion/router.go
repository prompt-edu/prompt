package coursePhaseDeletion

import (
	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

// RegisterRoutes mounts the phase deletion endpoint. The router group must carry the
// :coursePhaseID path parameter, and the SDK protects the endpoint itself.
func RegisterRoutes(rg *gin.RouterGroup, svc *CoursePhaseDeletionService) {
	promptTypes.RegisterPhaseDeletionModule(rg, svc)
}
