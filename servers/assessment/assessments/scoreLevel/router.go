package scoreLevel

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	log "github.com/sirupsen/logrus"
)

var _ = scoreLevelDTO.ScoreLevelWithParticipation{}

// setupScoreLevelRouter sets up score level endpoints.
// @Summary Score Level Endpoints
// @Description Access score levels for assessments.
// @Tags score_levels
// @Security BearerAuth
func setupScoreLevelRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	scoreLevelRouter := routerGroup.Group("/student-assessment/scoreLevel")

	scoreLevelRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getAllScoreLevels)
	scoreLevelRouter.GET("/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getScoreLevelByCourseParticipationID)
}

// getAllScoreLevels godoc
// @Summary List score levels
// @Description List score levels for the course phase.
// @Tags score_levels
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} scoreLevelDTO.ScoreLevelWithParticipation
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/scoreLevel [get]
func getAllScoreLevels(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	scoreLevels, err := GetAllScoreLevels(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, scoreLevels)
}

// getScoreLevelByCourseParticipationID godoc
// @Summary Get score level for student
// @Description Get score level for a course participation.
// @Tags score_levels
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param courseParticipationID path string true "Course participation ID"
// @Success 200 {string} string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/scoreLevel/{courseParticipationID} [get]
func getScoreLevelByCourseParticipationID(c *gin.Context) {
	courseParticipationID, err := uuid.Parse(c.Param("courseParticipationID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	scoreLevel, err := GetScoreLevelByCourseParticipationID(c, courseParticipationID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, scoreLevel)
}

func handleError(c *gin.Context, statusCode int, err error) {
	log.Error(err)
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
