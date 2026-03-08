package participants

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/ls1intum/prompt-sdk"
	log "github.com/sirupsen/logrus"
)

func setupParticipantsRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	participantsRouter := routerGroup.Group("/participants")

	participantsRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getParticipants)
}

func getParticipants(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.WithError(err).Error("Failed to parse course phase ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	participants, err := GetParticipationsForCoursePhase(c, authHeader, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to get participants")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve participants"})
		return
	}

	c.JSON(http.StatusOK, participants)
}
