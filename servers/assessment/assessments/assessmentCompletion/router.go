package assessmentCompletion

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion/assessmentCompletionDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	"github.com/prompt-edu/prompt/servers/assessment/utils"
	log "github.com/sirupsen/logrus"
)

// setupAssessmentCompletionRouter sets up assessment completion endpoints.
// @Summary Assessment Completion Endpoints
// @Description Manage assessment completion and grades.
// @Tags assessment_completions
// @Security BearerAuth
func setupAssessmentCompletionRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	assessmentCompletionRouter := routerGroup.Group("/student-assessment/completed")

	// course phase communication
	assessmentCompletionRouter.GET("grade", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getAllGrades)
	assessmentCompletionRouter.GET("grade/course-participation/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getStudentGrade)

	assessmentCompletionRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), listAssessmentCompletionsByCoursePhase)
	assessmentCompletionRouter.POST("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), createOrUpdateAssessmentCompletion)
	assessmentCompletionRouter.PUT("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), createOrUpdateAssessmentCompletion)
	assessmentCompletionRouter.POST("/mark-complete", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), markAssessmentAsCompleted)
	assessmentCompletionRouter.GET("/course-participation/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getAssessmentCompletion)
	assessmentCompletionRouter.PUT("/course-participation/:courseParticipationID/unmark", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), unmarkAssessmentAsCompleted)
	assessmentCompletionRouter.DELETE("/course-participation/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), deleteAssessmentCompletion)

	assessmentCompletionRouter.GET("/my-grade-suggestion", authMiddleware(promptSDK.CourseStudent), getMyGradeSuggestion)
}

// getAllGrades godoc
// @Summary List grades
// @Description List grade suggestions for all course participations.
// @Tags assessment_completions
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} assessmentCompletionDTO.GradeWithParticipation
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/completed/grade [get]
func getAllGrades(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	grades, err := GetAllGrades(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, grades)
}

// getStudentGrade godoc
// @Summary Get grade for student
// @Description Get grade suggestion for a course participation.
// @Tags assessment_completions
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param courseParticipationID path string true "Course participation ID"
// @Success 200 {number} float64
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/completed/grade/course-participation/{courseParticipationID} [get]
func getStudentGrade(c *gin.Context) {
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
	grade, err := GetStudentGrade(c, courseParticipationID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, grade)
}

// listAssessmentCompletionsByCoursePhase godoc
// @Summary List assessment completions
// @Description List assessment completions for a course phase.
// @Tags assessment_completions
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} assessmentCompletionDTO.AssessmentCompletion
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/completed [get]
func listAssessmentCompletionsByCoursePhase(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	completions, err := ListAssessmentCompletionsByCoursePhase(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, assessmentCompletionDTO.GetAssessmentCompletionDTOsFromDBModels(completions))
}

// createOrUpdateAssessmentCompletion godoc
// @Summary Create or update assessment completion
// @Description Create or update an assessment completion.
// @Tags assessment_completions
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param completion body assessmentCompletionDTO.AssessmentCompletion true "Assessment completion payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/completed [post]
// @Router /course_phase/{coursePhaseID}/student-assessment/completed [put]
func createOrUpdateAssessmentCompletion(c *gin.Context) {
	var req assessmentCompletionDTO.AssessmentCompletion
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	err := CreateOrUpdateAssessmentCompletion(c, req)
	if err != nil {
		if errors.Is(err, coursePhaseConfig.ErrNotStarted) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Assessment completion created/updated successfully"})
}

// markAssessmentAsCompleted godoc
// @Summary Mark assessment as completed
// @Description Mark an assessment as completed.
// @Tags assessment_completions
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param completion body assessmentCompletionDTO.AssessmentCompletion true "Assessment completion payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/completed/mark-complete [post]
func markAssessmentAsCompleted(c *gin.Context) {
	var req assessmentCompletionDTO.AssessmentCompletion
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	err := MarkAssessmentAsCompleted(c, req)
	if err != nil {
		if errors.Is(err, coursePhaseConfig.ErrNotStarted) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Assessment marked as completed successfully"})
}

// deleteAssessmentCompletion godoc
// @Summary Delete assessment completion
// @Description Delete an assessment completion by course participation ID.
// @Tags assessment_completions
// @Param coursePhaseID path string true "Course phase ID"
// @Param courseParticipationID path string true "Course participation ID"
// @Success 200 {string} string "OK"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/completed/course-participation/{courseParticipationID} [delete]
func deleteAssessmentCompletion(c *gin.Context) {
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
	if err := DeleteAssessmentCompletion(c, courseParticipationID, coursePhaseID); err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

// unmarkAssessmentAsCompleted godoc
// @Summary Unmark assessment as completed
// @Description Unmark an assessment completion.
// @Tags assessment_completions
// @Param coursePhaseID path string true "Course phase ID"
// @Param courseParticipationID path string true "Course participation ID"
// @Success 200 {string} string "OK"
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/completed/course-participation/{courseParticipationID}/unmark [put]
func unmarkAssessmentAsCompleted(c *gin.Context) {
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
	if err := UnmarkAssessmentAsCompleted(c, courseParticipationID, coursePhaseID); err != nil {
		// Check if the error is due to deadline being passed
		if errors.Is(err, coursePhaseConfig.ErrDeadlinePassed) {
			handleError(c, http.StatusForbidden, err)
		} else {
			handleError(c, http.StatusInternalServerError, err)
		}
		return
	}
	c.Status(http.StatusOK)
}

// getAssessmentCompletion godoc
// @Summary Get assessment completion
// @Description Get an assessment completion by course participation ID.
// @Tags assessment_completions
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param courseParticipationID path string true "Course participation ID"
// @Success 200 {object} assessmentCompletionDTO.AssessmentCompletion
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/completed/course-participation/{courseParticipationID} [get]
func getAssessmentCompletion(c *gin.Context) {
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
	assessmentCompletion, err := GetAssessmentCompletion(c, courseParticipationID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, assessmentCompletion)
}

// getMyGradeSuggestion godoc
// @Summary Get my grade suggestion
// @Description Get grade suggestion for the current student.
// @Tags assessment_completions
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {number} float64
// @Success 204 {string} string "No Content"
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/completed/my-grade-suggestion [get]
func getMyGradeSuggestion(c *gin.Context) {
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
	if !config.GradeSuggestionVisible {
		handleError(c, http.StatusForbidden, fmt.Errorf("grade suggestions are not visible to students"))
		return
	}

	courseParticipationID, err := utils.GetUserCourseParticipationID(c)
	if err != nil {
		handleError(c, utils.GetUserCourseParticipationIDErrorStatus(err), err)
		return
	}

	if !config.ResultsReleased {
		c.Status(http.StatusNoContent)
		return
	}

	exists, err := CheckAssessmentCompletionExists(c, courseParticipationID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	if exists {
		completion, err := GetAssessmentCompletion(c, courseParticipationID, coursePhaseID)
		if err != nil {
			handleError(c, http.StatusInternalServerError, err)
			return
		}
		if !completion.Completed {
			c.Status(http.StatusNoContent)
			return
		}
		c.JSON(http.StatusOK, completion.GradeSuggestion)
	}
	c.Status(http.StatusNoContent)
}

func handleError(c *gin.Context, statusCode int, err error) {
	log.Error(err)
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
