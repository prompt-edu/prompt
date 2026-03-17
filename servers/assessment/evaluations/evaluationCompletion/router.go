package evaluationCompletion

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationCompletion/evaluationCompletionDTO"
	"github.com/prompt-edu/prompt/servers/assessment/utils"
	log "github.com/sirupsen/logrus"
)

// setupEvaluationCompletionRouter sets up evaluation completion endpoints.
// @Summary Evaluation Completion Endpoints
// @Description Manage evaluation completion for students.
// @Tags evaluation_completions
// @Security BearerAuth
func setupEvaluationCompletionRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	evaluationRouter := routerGroup.Group("/evaluation/completed")

	evaluationRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), listEvaluationCompletionsByCoursePhase)

	evaluationRouter.POST("/my-completion", authMiddleware(promptSDK.CourseStudent), createOrUpdateMyEvaluationCompletion)
	evaluationRouter.PUT("/my-completion", authMiddleware(promptSDK.CourseStudent), createOrUpdateMyEvaluationCompletion)
	evaluationRouter.POST("/my-completion/mark-complete", authMiddleware(promptSDK.CourseStudent), markMyEvaluationAsCompleted)
	evaluationRouter.PUT("/my-completion/unmark", authMiddleware(promptSDK.CourseStudent), unmarkMyEvaluationAsCompleted)
	evaluationRouter.GET("/my-completions", authMiddleware(promptSDK.CourseStudent), getMyEvaluationCompletions)

}

// listEvaluationCompletionsByCoursePhase godoc
// @Summary List evaluation completions by course phase
// @Description List evaluation completions for a course phase.
// @Tags evaluation_completions
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} evaluationCompletionDTO.EvaluationCompletion
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/evaluation/completed [get]
func listEvaluationCompletionsByCoursePhase(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	completions, err := ListEvaluationCompletionsByCoursePhase(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, evaluationCompletionDTO.GetEvaluationCompletionDTOsFromDBModels(completions))
}

// createOrUpdateMyEvaluationCompletion godoc
// @Summary Create or update my evaluation completion
// @Description Create or update evaluation completion for the current student.
// @Tags evaluation_completions
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param completion body evaluationCompletionDTO.EvaluationCompletion true "Evaluation completion payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/evaluation/completed/my-completion [post]
// @Router /course_phase/{coursePhaseID}/evaluation/completed/my-completion [put]
func createOrUpdateMyEvaluationCompletion(c *gin.Context) {
	var req evaluationCompletionDTO.EvaluationCompletion
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	statusCode, err := utils.ValidateStudentOwnership(c, req.AuthorCourseParticipationID)
	if err != nil {
		handleError(c, statusCode, err)
		return
	}

	err = CreateOrUpdateEvaluationCompletion(c, req)
	if err != nil {
		if errors.Is(err, coursePhaseConfig.ErrNotStarted) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Evaluation completion created/updated successfully"})
}

// markMyEvaluationAsCompleted godoc
// @Summary Mark my evaluation as completed
// @Description Mark evaluation as completed for the current student.
// @Tags evaluation_completions
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param completion body evaluationCompletionDTO.EvaluationCompletion true "Evaluation completion payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/evaluation/completed/my-completion/mark-complete [post]
func markMyEvaluationAsCompleted(c *gin.Context) {
	var req evaluationCompletionDTO.EvaluationCompletion
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	statusCode, err := utils.ValidateStudentOwnership(c, req.AuthorCourseParticipationID)
	if err != nil {
		handleError(c, statusCode, err)
		return
	}

	err = MarkEvaluationAsCompleted(c, req)
	if err != nil {
		if errors.Is(err, coursePhaseConfig.ErrNotStarted) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Evaluation marked as completed successfully"})
}

// unmarkMyEvaluationAsCompleted godoc
// @Summary Unmark my evaluation as completed
// @Description Unmark evaluation as completed for the current student.
// @Tags evaluation_completions
// @Accept json
// @Param coursePhaseID path string true "Course phase ID"
// @Param completion body evaluationCompletionDTO.EvaluationCompletion true "Evaluation completion payload"
// @Success 200 {string} string "OK"
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/evaluation/completed/my-completion/unmark [put]
func unmarkMyEvaluationAsCompleted(c *gin.Context) {
	var req struct {
		CourseParticipationID       uuid.UUID `json:"courseParticipationID"`
		CoursePhaseID               uuid.UUID `json:"coursePhaseID"`
		AuthorCourseParticipationID uuid.UUID `json:"authorCourseParticipationID"`
	}
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	statusCode, err := utils.ValidateStudentOwnership(c, req.AuthorCourseParticipationID)
	if err != nil {
		handleError(c, statusCode, err)
		return
	}

	if err := UnmarkEvaluationAsCompleted(c, req.CourseParticipationID, req.CoursePhaseID, req.AuthorCourseParticipationID); err != nil {
		if errors.Is(err, coursePhaseConfig.ErrDeadlinePassed) {
			handleError(c, http.StatusForbidden, err)
		} else {
			handleError(c, http.StatusInternalServerError, err)
		}
		return
	}
	c.Status(http.StatusOK)
}

// getMyEvaluationCompletions godoc
// @Summary List my evaluation completions
// @Description List evaluation completions for the current student.
// @Tags evaluation_completions
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} evaluationCompletionDTO.EvaluationCompletion
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/evaluation/completed/my-completions [get]
func getMyEvaluationCompletions(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	userCourseParticipationUUID, err := utils.GetUserCourseParticipationID(c)
	if err != nil {
		handleError(c, utils.GetUserCourseParticipationIDErrorStatus(err), err)
		return
	}

	evaluationCompletions, err := GetEvaluationCompletionsForAuthorInPhase(c, userCourseParticipationUUID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, evaluationCompletionDTO.GetEvaluationCompletionDTOsFromDBModels(evaluationCompletions))
}

func handleError(c *gin.Context, statusCode int, err error) {
	log.Error(err)
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
