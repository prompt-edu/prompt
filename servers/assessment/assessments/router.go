package assessments

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	"github.com/prompt-edu/prompt/servers/assessment/utils"
	log "github.com/sirupsen/logrus"
)

// setupAssessmentRouter sets up assessment endpoints.
// @Summary Assessment Endpoints
// @Description Manage assessments for course participation.
// @Tags assessments
// @Security BearerAuth
func setupAssessmentRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	assessmentRouter := routerGroup.Group("/student-assessment")

	assessmentRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), listAssessmentsByCoursePhase)
	assessmentRouter.POST("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), createOrUpdateAssessment)
	assessmentRouter.GET("/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getStudentAssessment)
	assessmentRouter.GET("/course-participation/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), listAssessmentsByStudentInPhase)
	assessmentRouter.DELETE("/:assessmentID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), deleteAssessment)

	assessmentRouter.GET("/my-results", authMiddleware(promptSDK.CourseStudent), getMyAssessmentResults)
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
func listAssessmentsByCoursePhase(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	assessments, err := ListAssessmentsByCoursePhase(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, assessmentDTO.GetAssessmentDTOsFromDBModels(assessments))
}

// createOrUpdateAssessment godoc
// @Summary Create or update assessment
// @Description Create or update an assessment for a student.
// @Tags assessments
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param assessment body assessmentDTO.CreateOrUpdateAssessmentRequest true "Assessment payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment [post]
func createOrUpdateAssessment(c *gin.Context) {
	var req assessmentDTO.CreateOrUpdateAssessmentRequest
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	err := CreateOrUpdateAssessment(c, req)
	if err != nil {
		// Check if it's a validation error
		if errors.Is(err, ErrValidationFailed) || errors.Is(err, ErrInvalidScoreLevel) {
			handleError(c, http.StatusBadRequest, err)
		} else {
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
func getStudentAssessment(c *gin.Context) {
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

	studentAssessment, err := GetStudentAssessment(c, coursePhaseID, courseParticipationID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, studentAssessment)
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
func getMyAssessmentResults(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	config, err := coursePhaseConfig.GetCoursePhaseConfig(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	// Do not expose results before they are released
	if !config.ResultsReleased {
		c.Status(http.StatusNoContent)
		return
	}

	courseParticipationID, err := utils.GetUserCourseParticipationID(c)
	if err != nil {
		handleError(c, utils.GetUserCourseParticipationIDErrorStatus(err), err)
		return
	}

	// Students can only see results after they have a completed assessment
	exists, err := assessmentCompletion.CheckAssessmentCompletionExists(c, courseParticipationID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	if !exists {
		c.Status(http.StatusNoContent)
		return
	}
	completion, err := assessmentCompletion.GetAssessmentCompletion(c, courseParticipationID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	if !completion.Completed {
		c.Status(http.StatusNoContent)
		return
	}

	results, err := GetStudentAssessmentResults(c, coursePhaseID, courseParticipationID, config)
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
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/{assessmentID} [delete]
func deleteAssessment(c *gin.Context) {
	assessmentID, err := uuid.Parse(c.Param("assessmentID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	if err := DeleteAssessment(c, assessmentID); err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.String(http.StatusOK, "OK")
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
func listAssessmentsByStudentInPhase(c *gin.Context) {
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
	assessments, err := ListAssessmentsByStudentInPhase(c, courseParticipationID, coursePhaseID)
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
