package template

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	log "github.com/sirupsen/logrus"
)

// setupTemplateRouter creates a router group for template server endpoints.
func setupTemplateRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	templateRouter := routerGroup.Group("info")

	templateRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getTemplateInfo)
}

// getTemplateInfo godoc
// @Summary Get template info
// @Description Get template-specific info for a course phase.
// @Tags template
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {string} string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/info [get]
func getTemplateInfo(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	templateInfo, err := GetTemplateInfo(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, templateInfo)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
