package example

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	log "github.com/sirupsen/logrus"
)

// RegisterRoutes mounts the example endpoints on the given router group. Route
// registration is separate from construction, so a caller can obtain the service
// without a router and the handlers are methods on the service they use.
func RegisterRoutes(routerGroup *gin.RouterGroup, service *ExampleService, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	exampleRouter := routerGroup.Group("info")

	exampleRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service.getExampleInfo)
}

// getExampleInfo godoc
// @Summary Get example info
// @Description Get example-specific info for a course phase.
// @Tags example
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {string} string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/info [get]
func (s *ExampleService) getExampleInfo(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	exampleInfo, err := s.GetExampleInfo(c.Request.Context(), coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, exampleInfo)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
