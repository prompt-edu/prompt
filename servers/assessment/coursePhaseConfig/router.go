package coursePhaseConfig

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig/coursePhaseConfigDTO"
	log "github.com/sirupsen/logrus"
)

// RegisterRoutes sets up course phase config endpoints.
// @Summary Course Phase Config Endpoints
// @Description Manage course phase configuration and communication data.
// @Tags course_phase_config
// @Security BearerAuth
func RegisterRoutes(routerGroup *gin.RouterGroup, service *CoursePhaseConfigService, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	coursePhaseRouter := routerGroup.Group("/config")

	coursePhaseRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), service.getCoursePhaseConfig)
	coursePhaseRouter.PUT("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service.createOrUpdateCoursePhaseConfig)
	coursePhaseRouter.POST("/release", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), audit.Describe("Released assessment results"), service.releaseResults)
	coursePhaseRouter.POST("/unrelease", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), audit.Describe("Withdrew released assessment results"), service.unreleaseResults)
	coursePhaseRouter.GET("/reminders/incomplete", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), service.getIncompleteReminderRecipients)
	coursePhaseRouter.POST("/reminders/send", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), audit.Describe("Sent evaluation reminders"), service.sendEvaluationReminder)

	coursePhaseRouter.GET("participations", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), service.getParticipationsForCoursePhase)
	coursePhaseRouter.GET("teams", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), service.getTeamsForCoursePhase)

}

// getCoursePhaseConfig godoc
// @Summary Get course phase config
// @Description Get the course phase configuration.
// @Tags course_phase_config
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {object} coursePhaseConfigDTO.CoursePhaseConfig
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/config [get]
func (s *CoursePhaseConfigService) getCoursePhaseConfig(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.WithError(err).Error("Failed to parse course phase ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	config, err := s.GetCoursePhaseConfig(c, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to get course phase config")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve course phase config"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// createOrUpdateCoursePhaseConfig godoc
// @Summary Create or update course phase config
// @Description Create or update the course phase configuration.
// @Tags course_phase_config
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param config body coursePhaseConfigDTO.CreateOrUpdateCoursePhaseConfigRequest true "Course phase config payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/config [put]
func (s *CoursePhaseConfigService) createOrUpdateCoursePhaseConfig(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.WithError(err).Error("Failed to parse course phase ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	var request coursePhaseConfigDTO.CreateOrUpdateCoursePhaseConfigRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.WithError(err).Error("Failed to bind request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	err = s.CreateOrUpdateCoursePhaseConfig(c, coursePhaseID, request)
	if err != nil {
		if errors.Is(err, ErrCannotChangeSchemaWithData) || errors.Is(err, ErrCannotDisableAssessmentWithData) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, assessmentSchemas.ErrSchemaNotAccessible) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		log.WithError(err).Error("Failed to create or update course phase config")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create or update course phase config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course phase config created/updated successfully"})
}

// releaseResults godoc
// @Summary Release assessment results
// @Description Release assessment results for the course phase.
// @Tags course_phase_config
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/config/release [post]
func (s *CoursePhaseConfigService) releaseResults(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.WithError(err).Error("Failed to parse course phase ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	err = s.ReleaseResults(c, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to release results")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to release results"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Results released successfully"})
}

// unreleaseResults godoc
// @Summary Unrelease assessment results
// @Description Unrelease assessment results for the course phase.
// @Tags course_phase_config
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/config/unrelease [post]
func (s *CoursePhaseConfigService) unreleaseResults(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.WithError(err).Error("Failed to parse course phase ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	err = s.UnreleaseResults(c, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to unrelease results")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unrelease results"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Results unreleased successfully"})
}

// getParticipationsForCoursePhase godoc
// @Summary List participations for course phase
// @Description Get course participations for a course phase from core service.
// @Tags course_phase_config
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} coursePhaseConfigDTO.AssessmentParticipationWithStudent
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/config/participations [get]
func (s *CoursePhaseConfigService) getParticipationsForCoursePhase(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.WithError(err).Error("Failed to parse course phase ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	participations, err := GetParticipationsForCoursePhase(c, authHeader, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to get participations for course phase")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve participations"})
		return
	}

	c.JSON(http.StatusOK, participations)
}

// getTeamsForCoursePhase godoc
// @Summary List teams for course phase
// @Description Get teams for a course phase from core service.
// @Tags course_phase_config
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/config/teams [get]
func (s *CoursePhaseConfigService) getTeamsForCoursePhase(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.WithError(err).Error("Failed to parse course phase ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	teams, err := GetTeamsForCoursePhase(c, authHeader, coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to get teams for course phase")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve teams"})
		return
	}

	c.JSON(http.StatusOK, teams)
}

// getIncompleteReminderRecipients godoc
// @Summary List incomplete reminder recipients by evaluation type
// @Description Returns authors who have not fully completed evaluations for the selected type.
// @Tags course_phase_config
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param type query string true "Evaluation type (self|peer|tutor)"
// @Success 200 {object} coursePhaseConfigDTO.EvaluationReminderRecipients
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/config/reminders/incomplete [get]
func (s *CoursePhaseConfigService) getIncompleteReminderRecipients(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.WithError(err).Error("Failed to parse course phase ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	evaluationType, err := parseEvaluationType(c.Query("type"))
	if err != nil {
		log.WithError(err).Error("Invalid reminder evaluation type")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authHeader := c.GetHeader("Authorization")
	response, err := s.getEvaluationReminderRecipients(c, authHeader, coursePhaseID, evaluationType)
	if err != nil {
		log.WithError(err).Error("Failed to compute incomplete reminder recipients")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compute reminder recipients"})
		return
	}
	if !response.EvaluationEnabled {
		c.JSON(http.StatusConflict, gin.H{
			"error": "evaluation type is disabled for this course phase",
		})
		return
	}
	if !response.DeadlinePassed {
		deadlineMessage := "evaluation deadline has not passed yet"
		if response.Deadline != nil {
			deadlineMessage = "evaluation deadline has not passed yet (deadline: " + response.Deadline.Format(time.RFC3339) + ")"
		}
		c.JSON(http.StatusConflict, gin.H{
			"error": deadlineMessage,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func parseEvaluationType(raw string) (assessmentType.AssessmentType, error) {
	switch raw {
	case string(assessmentType.Self):
		return assessmentType.Self, nil
	case string(assessmentType.Peer):
		return assessmentType.Peer, nil
	case string(assessmentType.Tutor):
		return assessmentType.Tutor, nil
	default:
		return "", errors.New("invalid evaluation type, expected one of: self, peer, tutor")
	}
}

// sendEvaluationReminder godoc
// @Summary Send evaluation reminders
// @Description Computes incomplete recipients in assessment and triggers core manual mailing for the selected type.
// @Tags course_phase_config
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param request body coursePhaseConfigDTO.SendEvaluationReminderRequest true "Evaluation reminder send request"
// @Success 200 {object} coursePhaseConfigDTO.EvaluationReminderSendReport
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/config/reminders/send [post]
func (s *CoursePhaseConfigService) sendEvaluationReminder(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.WithError(err).Error("Failed to parse course phase ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	var request coursePhaseConfigDTO.SendEvaluationReminderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.WithError(err).Error("Failed to bind request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	if request.EvaluationType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid evaluation type, expected one of: self, peer, tutor"})
		return
	}
	evaluationType, err := parseEvaluationType(string(request.EvaluationType))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := s.sendEvaluationReminderManualTrigger(
		c,
		c.GetHeader("Authorization"),
		coursePhaseID,
		evaluationType,
	)
	if err != nil {
		if errors.Is(err, ErrReminderDeadlineNotPassed) ||
			errors.Is(err, ErrReminderEvaluationDisabled) ||
			errors.Is(err, ErrReminderTemplateIncomplete) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		log.WithError(err).Error("Failed to send evaluation reminder")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send evaluation reminder"})
		return
	}

	c.JSON(http.StatusOK, report)
}
