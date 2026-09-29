package coursePhaseDeletion

import (
	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

// RegisterRoutes mounts the phase deletion endpoint. The router group must carry the
// :coursePhaseID path parameter, and the SDK protects the endpoint itself.
func RegisterRoutes(routerGroup *gin.RouterGroup, service *CoursePhaseDeletionService) {
	// The module deletes only its own data, so the derived "Deleted course phase" would mislead.
	promptTypes.RegisterPhaseDeletionModule(routerGroup.Group("", audit.Describe("Deleted the team allocation phase data")), service)
}
