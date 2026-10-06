package assessments

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	log "github.com/sirupsen/logrus"
)

// RegisterRoutes sets up assessment endpoints.
// @Summary Assessment Endpoints
// @Description Manage assessments for course participation.
// @Tags assessments
// @Security BearerAuth
type assessmentGuard interface {
	RequireAssessmentEnabled() gin.HandlerFunc
	RequireIndependentAssessmentEnabled() gin.HandlerFunc
}

func RegisterRoutes(routerGroup *gin.RouterGroup, service *AssessmentService, guard assessmentGuard, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	assessmentRouter := routerGroup.Group("/student-assessment")

	assessmentRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), service.listAssessmentsByCoursePhase)
	// The grading form posts on every score selection, so auditing this route would
	// bury the log and start dropping events. Completion transitions are audited instead.
	assessmentRouter.POST("", audit.Skip(), authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), guard.RequireAssessmentEnabled(), service.createOrUpdateAssessment)
	assessmentRouter.POST("/independent", audit.Skip(), authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), guard.RequireAssessmentEnabled(), guard.RequireIndependentAssessmentEnabled(), service.createOrUpdateIndependentAssessment)
	assessmentRouter.GET("/:courseParticipationID/export", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), service.exportStudentAssessment)
	assessmentRouter.GET("/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), service.getStudentAssessment)
	assessmentRouter.GET("/course-participation/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), service.listAssessmentsByStudentInPhase)
	assessmentRouter.DELETE("/:assessmentID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), guard.RequireAssessmentEnabled(), service.deleteAssessment)
	assessmentRouter.DELETE("/independent/:independentAssessmentID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), guard.RequireAssessmentEnabled(), service.deleteOwnIndependentAssessment)

	assessmentRouter.GET("/my-results", authMiddleware(promptSDK.CourseStudent), service.getMyAssessmentResults)
}

// listAssessmentsByCoursePhase godoc
// @Summary List assessments by course phase
// @Description List all assessments for a course phase.
// @Tags assessments
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} assessmentDTO.Assessment
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment [get]
func (s *AssessmentService) listAssessmentsByCoursePhase(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	assessments, err := s.ListAssessmentsByCoursePhase(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, assessmentDTO.GetAssessmentDTOsFromDBModels(assessments))
}

// createOrUpdateAssessment godoc
// @Summary Create or update assessment
// @Description Create or update an assessment for a student. The author identity is taken from the authenticated JWT and any client-sent author fields are ignored.
// @Tags assessments
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param assessment body assessmentDTO.CreateOrUpdateAssessmentRequest true "Assessment payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment [post]
func (s *AssessmentService) createOrUpdateAssessment(c *gin.Context) {
	saveScoreFromRequest(c, s.CreateOrUpdateAssessment)
}

// createOrUpdateIndependentAssessment godoc
// @Summary Create or update independent assessment
// @Description Create or update the caller's own score for a student, kept apart from other assessors' scores and from the final assessment. The author identity is taken from the authenticated JWT and any client-sent author fields are ignored.
// @Tags assessments
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param assessment body assessmentDTO.CreateOrUpdateAssessmentRequest true "Assessment payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/independent [post]
func (s *AssessmentService) createOrUpdateIndependentAssessment(c *gin.Context) {
	saveScoreFromRequest(c, s.CreateOrUpdateIndependentAssessment)
}

func saveScoreFromRequest(c *gin.Context, save func(context.Context, assessmentDTO.CreateOrUpdateAssessmentRequest) error) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var req assessmentDTO.CreateOrUpdateAssessmentRequest
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	tokenUser, ok := keycloakTokenVerifier.GetTokenUser(c)
	if !ok {
		handleError(c, http.StatusUnauthorized, errors.New("authenticated user not found in context"))
		return
	}
	req.Author = tokenUser.FirstName + " " + tokenUser.LastName
	req.AuthorID = tokenUser.ID
	// The authorized phase is the one in the URL; ignore any client-sent phase.
	req.CoursePhaseID = coursePhaseID

	err = save(c, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidScoreLevel), errors.Is(err, ErrCompetencyNotInPhase):
			handleError(c, http.StatusBadRequest, err)
		case errors.Is(err, assessmentCompletion.ErrAssessmentCompleted), errors.Is(err, coursePhaseConfig.ErrNotStarted):
			handleError(c, http.StatusConflict, err)
		default:
			handleError(c, http.StatusInternalServerError, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Assessment created/updated successfully"})
}

// getStudentAssessment godoc
// @Summary Get student assessment
// @Description Get an assessment for a course participation.
// @Tags assessments
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param courseParticipationID path string true "Course participation ID"
// @Success 200 {object} assessmentDTO.StudentAssessment
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/{courseParticipationID} [get]
func (s *AssessmentService) getStudentAssessment(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	courseParticipationID, err := uuid.Parse(c.Param("courseParticipationID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	tokenUser, ok := keycloakTokenVerifier.GetTokenUser(c)
	if !ok {
		handleError(c, http.StatusUnauthorized, errors.New("authenticated user not found in context"))
		return
	}

	studentAssessment, err := s.GetStudentAssessment(c, coursePhaseID, courseParticipationID, tokenUser.ID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, studentAssessment)
}

// exportStudentAssessment godoc
// @Summary Export student assessment
// @Description Export one student assessment in a text format.
// @Tags assessments
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param courseParticipationID path string true "Course participation ID"
// @Param format query string false "Export format" default(json)
// @Success 200 {object} assessmentDTO.AssessmentExport
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/{courseParticipationID}/export [get]
func (s *AssessmentService) exportStudentAssessment(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	courseParticipationID, err := uuid.Parse(c.Param("courseParticipationID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	format := c.DefaultQuery("format", AssessmentExportFormatJSON)
	export, err := s.ExportStudentAssessment(c, coursePhaseID, courseParticipationID, format)
	if err != nil {
		if errors.Is(err, ErrUnsupportedAssessmentExportFormat) {
			handleError(c, http.StatusBadRequest, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Header("Content-Disposition", "attachment; filename=assessment_"+courseParticipationID.String()+".json")
	c.JSON(http.StatusOK, export)
}

// getMyAssessmentResults godoc
// @Summary Get my assessment results
// @Description Get assessment results for the current student.
// @Tags assessments
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {object} assessmentDTO.StudentAssessmentResults
// @Success 204 {string} string "No Content"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/my-results [get]
func (s *AssessmentService) getMyAssessmentResults(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	config, err := s.coursePhaseConfig.GetStoredCoursePhaseConfig(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	// Do not expose results before they are released, or at all on evaluation-only phases
	if !config.AssessmentEnabled || !config.ResultsReleased {
		c.Status(http.StatusNoContent)
		return
	}

	courseParticipationID, err := keycloakTokenVerifier.GetUserCourseParticipationID(c)
	if err != nil {
		handleError(c, keycloakTokenVerifier.GetUserCourseParticipationIDErrorStatus(err), err)
		return
	}

	// Students can only see results after they have a completed assessment
	exists, err := s.assessmentCompletion.CheckAssessmentCompletionExists(c, courseParticipationID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	if !exists {
		c.Status(http.StatusNoContent)
		return
	}
	completion, err := s.assessmentCompletion.GetAssessmentCompletion(c, courseParticipationID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	if !completion.Completed {
		c.Status(http.StatusNoContent)
		return
	}

	results, err := s.GetStudentAssessmentResults(c, coursePhaseID, courseParticipationID, config)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, results)
}

// deleteAssessment godoc
// @Summary Delete assessment
// @Description Delete an assessment by ID.
// @Tags assessments
// @Param coursePhaseID path string true "Course phase ID"
// @Param assessmentID path string true "Assessment ID"
// @Success 200 {string} string "OK"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/{assessmentID} [delete]
func (s *AssessmentService) deleteAssessment(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	assessmentID, err := uuid.Parse(c.Param("assessmentID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	if err := s.DeleteAssessment(c, assessmentID, coursePhaseID); err != nil {
		if errors.Is(err, ErrAssessmentNotInPhase) || errors.Is(err, ErrAssessmentNotFound) {
			handleError(c, http.StatusNotFound, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.String(http.StatusOK, "OK")
}

// deleteOwnIndependentAssessment godoc
// @Summary Delete own independent assessment
// @Description Delete one of the caller's own independent scores. Scores of other assessors are reported as not found.
// @Tags assessments
// @Param coursePhaseID path string true "Course phase ID"
// @Param independentAssessmentID path string true "Independent assessment ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/independent/{independentAssessmentID} [delete]
func (s *AssessmentService) deleteOwnIndependentAssessment(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	independentAssessmentID, err := uuid.Parse(c.Param("independentAssessmentID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	tokenUser, ok := keycloakTokenVerifier.GetTokenUser(c)
	if !ok {
		handleError(c, http.StatusUnauthorized, errors.New("authenticated user not found in context"))
		return
	}

	err = s.DeleteOwnIndependentAssessment(c, independentAssessmentID, coursePhaseID, tokenUser.ID)
	switch {
	case err == nil:
		c.Status(http.StatusNoContent)
	case errors.Is(err, ErrAssessmentNotFound):
		handleError(c, http.StatusNotFound, err)
	case errors.Is(err, assessmentCompletion.ErrAssessmentCompleted), errors.Is(err, coursePhaseConfig.ErrNotStarted):
		handleError(c, http.StatusConflict, err)
	default:
		handleError(c, http.StatusInternalServerError, err)
	}
}

// listAssessmentsByStudentInPhase godoc
// @Summary List assessments for student in phase
// @Description List assessments for a course participation in the course phase.
// @Tags assessments
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param courseParticipationID path string true "Course participation ID"
// @Success 200 {array} assessmentDTO.Assessment
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/course-participation/{courseParticipationID} [get]
func (s *AssessmentService) listAssessmentsByStudentInPhase(c *gin.Context) {
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
	assessments, err := s.ListAssessmentsByStudentInPhase(c, courseParticipationID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, assessmentDTO.GetAssessmentDTOsFromDBModels(assessments))
}

func handleError(c *gin.Context, statusCode int, err error) {
	log.Error(err)
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
