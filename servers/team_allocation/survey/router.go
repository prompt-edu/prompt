package survey

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/team_allocation/survey/surveyDTO"
	log "github.com/sirupsen/logrus"
)

// SetupSurveyRouter sets up survey endpoints under /survey.
func setupSurveyRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	surveyRouter := routerGroup.Group("/survey")
	// Endpoints accessible to CourseStudents.
	surveyRouter.GET("/form", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseStudent), getSurveyForm)
	surveyRouter.GET("/answers", authMiddleware(promptSDK.CourseStudent), getStudentSurveyResponses)
	surveyRouter.POST("/answers", authMiddleware(promptSDK.CourseStudent), submitSurveyResponses)

	surveyRouter.PUT("/timeframe", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), setSurveyTimeframe)
	surveyRouter.GET("/timeframe", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getSurveyTimeframe)
	surveyRouter.GET("/statistics", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getSurveyStatistics)

}

// getSurveyForm godoc
// @Summary Get survey form
// @Description Get available survey data (teams, skills, and deadline) for a course phase
// @Tags survey
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} surveyDTO.SurveyForm
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/survey/form [get]
// getAvailableSurveyData returns teams and skills if the survey has started.
// Expects coursePhaseID to be provided as a query parameter.
func getSurveyForm(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	surveyData, err := GetSurveyForm(c, coursePhaseID)
	if err != nil {
		if err.Error() == "survey has not started yet" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, surveyData)
}

// getStudentSurveyResponses godoc
// @Summary Get student survey responses
// @Description Get the authenticated student's survey answers
// @Tags survey
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} surveyDTO.StudentSurveyResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/survey/answers [get]
// getStudentSurveyResponses returns the student's submitted survey answers.
// Expects courseParticipationID to be provided as a query parameter.
func getStudentSurveyResponses(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	courseParticipationID, ok := c.Get("courseParticipationID")
	if !ok {
		log.Error("Error getting courseParticipationID from context")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	responses, err := GetStudentSurveyResponses(c, courseParticipationID.(uuid.UUID), coursePhaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, responses)
}

// submitSurveyResponses godoc
// @Summary Submit survey responses
// @Description Create or update the authenticated student's survey answers
// @Tags survey
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body surveyDTO.StudentSurveyResponse true "Survey responses"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/survey/answers [post]
// submitSurveyResponses accepts and stores (or overwrites) the student's survey answers.
// Expects courseParticipationID and coursePhaseID as query parameters.
func submitSurveyResponses(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	courseParticipationID, ok := c.Get("courseParticipationID")
	if !ok {
		log.Error("Error getting courseParticipationID from context")
		c.JSON(http.StatusBadRequest, gin.H{"error": "courseParticipationID not found in context"})
		return
	}

	var submission surveyDTO.StudentSurveyResponse
	if err := c.BindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// validate the request
	err = ValidateStudentResponse(c, coursePhaseID, submission)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = SubmitSurveyResponses(c, courseParticipationID.(uuid.UUID), coursePhaseID, submission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// setSurveyTimeframe godoc
// @Summary Set survey timeframe
// @Description Set or update the survey timeframe for a course phase
// @Tags survey
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body surveyDTO.SurveyTimeframe true "Survey timeframe"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/survey/timeframe [put]
// setSurveyTimeframe allows lecturers to set or update the survey timeframe.
// Expects coursePhaseID as a query parameter and the new timeframe in the JSON body.
func setSurveyTimeframe(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req surveyDTO.SurveyTimeframe
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = SetSurveyTimeframe(c, coursePhaseID, req.SurveyStart, req.SurveyDeadline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// getSurveyTimeframe godoc
// @Summary Get survey timeframe
// @Description Get the survey timeframe for a course phase
// @Tags survey
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} surveyDTO.SurveyTimeframe
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/survey/timeframe [get]
func getSurveyTimeframe(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	timeframe, err := GetSurveyTimeframe(c, coursePhaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, timeframe)
}

// getSurveyStatistics godoc
// @Summary Get survey statistics
// @Description Get aggregated team preference and skill distribution statistics for a course phase
// @Tags survey
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} surveyDTO.SurveyStatistics
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/survey/statistics [get]
func getSurveyStatistics(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	statistics, err := GetSurveyStatistics(c, coursePhaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, statistics)
}
