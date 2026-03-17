package timeframe

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/timeframe/timeframeDTO"
	log "github.com/sirupsen/logrus"
)

func setupTimeframeRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	teamRouter := routerGroup.Group("/timeframe")

	teamRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseStudent), getTimeframe)
	teamRouter.PUT("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), setTimeframe)
}

// getTimeframe godoc
// @Summary Get timeframe
// @Description Get the timeframe for team self-assignment
// @Tags timeframe
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} timeframeDTO.Timeframe
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/timeframe [get]
func getTimeframe(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}
	timeframe, err := GetTimeframe(c, coursePhaseID)
	if err != nil {
		if err.Error() == "timeframe not set" {
			c.JSON(http.StatusOK, timeframeDTO.Timeframe{TimeframeSet: false})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, timeframe)
}

// setTimeframe godoc
// @Summary Set timeframe
// @Description Set the timeframe during which students can self-assign to teams
// @Tags timeframe
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body timeframeDTO.Timeframe true "Timeframe start and end times"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/timeframe [put]
func setTimeframe(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var request timeframeDTO.Timeframe
	if err := c.BindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = SetTimeframe(c, coursePhaseID, request.StartTime, request.EndTime)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
