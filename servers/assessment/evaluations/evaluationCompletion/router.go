package evaluationCompletion

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationCompletion/evaluationCompletionDTO"
	log "github.com/sirupsen/logrus"
)

// RegisterRoutes sets up evaluation completion endpoints.
// @Summary Evaluation Completion Endpoints
// @Description Manage evaluation completion for students.
// @Tags evaluation_completions
// @Security BearerAuth
func RegisterRoutes(routerGroup *gin.RouterGroup, service *EvaluationCompletionService, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	evaluationRouter := routerGroup.Group("/evaluation/completed")

	evaluationRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), service.listEvaluationCompletionsByCoursePhase)

	evaluationRouter.POST("/my-completion", audit.Describe("Saved own evaluation completion"), authMiddleware(promptSDK.CourseStudent), service.createOrUpdateMyEvaluationCompletion)
	evaluationRouter.PUT("/my-completion", audit.Describe("Saved own evaluation completion"), authMiddleware(promptSDK.CourseStudent), service.createOrUpdateMyEvaluationCompletion)
	evaluationRouter.POST("/my-completion/mark-complete", audit.Describe("Marked own evaluation as completed"), authMiddleware(promptSDK.CourseStudent), service.markMyEvaluationAsCompleted)
	evaluationRouter.PUT("/my-completion/unmark", audit.Describe("Unmarked own evaluation completion"), authMiddleware(promptSDK.CourseStudent), service.unmarkMyEvaluationAsCompleted)
	evaluationRouter.GET("/my-completions", authMiddleware(promptSDK.CourseStudent), service.getMyEvaluationCompletions)

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
func (s *EvaluationCompletionService) listEvaluationCompletionsByCoursePhase(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	completions, err := s.ListEvaluationCompletionsByCoursePhase(c, coursePhaseID)
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
func (s *EvaluationCompletionService) createOrUpdateMyEvaluationCompletion(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var req evaluationCompletionDTO.EvaluationCompletion
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	// The authorized phase is the one in the URL; ignore any client-sent phase.
	req.CoursePhaseID = coursePhaseID

	statusCode, err := keycloakTokenVerifier.ValidateStudentOwnership(c, req.AuthorCourseParticipationID, "evaluation completions")
	if err != nil {
		handleError(c, statusCode, err)
		return
	}

	err = s.CreateOrUpdateEvaluationCompletion(c, c.GetHeader("Authorization"), req)
	if err != nil {
		if errors.Is(err, coursePhaseConfig.ErrNotStarted) || IsTargetAuthorizationError(err) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		if errors.Is(err, ErrEvaluationAlreadyCompleted) {
			handleError(c, http.StatusConflict, err)
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
func (s *EvaluationCompletionService) markMyEvaluationAsCompleted(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var req evaluationCompletionDTO.EvaluationCompletion
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	// The authorized phase is the one in the URL; ignore any client-sent phase.
	req.CoursePhaseID = coursePhaseID

	statusCode, err := keycloakTokenVerifier.ValidateStudentOwnership(c, req.AuthorCourseParticipationID, "evaluation completions")
	if err != nil {
		handleError(c, statusCode, err)
		return
	}

	err = s.MarkEvaluationAsCompleted(c, c.GetHeader("Authorization"), req)
	if err != nil {
		if errors.Is(err, coursePhaseConfig.ErrNotStarted) || IsTargetAuthorizationError(err) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		if errors.Is(err, ErrEvaluationAlreadyCompleted) {
			handleError(c, http.StatusConflict, err)
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
func (s *EvaluationCompletionService) unmarkMyEvaluationAsCompleted(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var req struct {
		CourseParticipationID       uuid.UUID `json:"courseParticipationID"`
		CoursePhaseID               uuid.UUID `json:"coursePhaseID"`
		AuthorCourseParticipationID uuid.UUID `json:"authorCourseParticipationID"`
	}
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	// The authorized phase is the one in the URL; ignore any client-sent phase.
	req.CoursePhaseID = coursePhaseID

	statusCode, err := keycloakTokenVerifier.ValidateStudentOwnership(c, req.AuthorCourseParticipationID, "evaluation completions")
	if err != nil {
		handleError(c, statusCode, err)
		return
	}

	if err := s.UnmarkEvaluationAsCompleted(c, req.CourseParticipationID, req.CoursePhaseID, req.AuthorCourseParticipationID); err != nil {
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
func (s *EvaluationCompletionService) getMyEvaluationCompletions(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	userCourseParticipationUUID, err := keycloakTokenVerifier.GetUserCourseParticipationID(c)
	if err != nil {
		handleError(c, keycloakTokenVerifier.GetUserCourseParticipationIDErrorStatus(err), err)
		return
	}

	evaluationCompletions, err := s.GetEvaluationCompletionsForAuthorInPhase(c, userCourseParticipationUUID, coursePhaseID)
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
