package evaluations

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationDTO"
	"github.com/prompt-edu/prompt/servers/assessment/utils"
	log "github.com/sirupsen/logrus"
)

// setupEvaluationRouter sets up evaluation endpoints.
// @Summary Evaluation Endpoints
// @Description Manage evaluations for course participations.
// @Tags evaluations
// @Security BearerAuth
func setupEvaluationRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	evaluationRouter := routerGroup.Group("/evaluation")

	// Admin/Lecturer/Editor endpoints - overview of all evaluations
	evaluationRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getAllEvaluationsByPhase)
	evaluationRouter.GET("/tutor/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getEvaluationsForTutorInPhase)

	// Student endpoints - access to own evaluations only
	evaluationRouter.GET("/my-evaluations", authMiddleware(promptSDK.CourseStudent), getMyEvaluations)
	evaluationRouter.POST("", authMiddleware(promptSDK.CourseStudent), createOrUpdateEvaluation)
	evaluationRouter.DELETE("/:evaluationID", authMiddleware(promptSDK.CourseStudent), deleteEvaluation)
}

// getAllEvaluationsByPhase godoc
// @Summary List evaluations by course phase
// @Description List evaluations for a course phase.
// @Tags evaluations
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} evaluationDTO.Evaluation
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/evaluation [get]
func getAllEvaluationsByPhase(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	evaluations, err := GetEvaluationsByPhase(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, evaluations)
}

// getEvaluationsForTutorInPhase godoc
// @Summary List evaluations for tutor in phase
// @Description List evaluations for a tutor in a course phase.
// @Tags evaluations
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param courseParticipationID path string true "Tutor course participation ID"
// @Success 200 {array} evaluationDTO.Evaluation
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/evaluation/tutor/{courseParticipationID} [get]
func getEvaluationsForTutorInPhase(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	tutorID, err := uuid.Parse(c.Param("courseParticipationID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	evaluations, err := GetEvaluationsForTutorInPhase(c, tutorID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, evaluations)
}

// getMyEvaluations godoc
// @Summary List my evaluations
// @Description List evaluations authored by the current student.
// @Tags evaluations
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} evaluationDTO.Evaluation
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/evaluation/my-evaluations [get]
func getMyEvaluations(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipationID, err := utils.GetUserCourseParticipationID(c)
	if err != nil {
		handleError(c, utils.GetUserCourseParticipationIDErrorStatus(err), err)
		return
	}

	evaluations, err := GetEvaluationsForAuthorInPhase(c, courseParticipationID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, evaluations)
}

// createOrUpdateEvaluation godoc
// @Summary Create or update evaluation
// @Description Create or update an evaluation for the current student.
// @Tags evaluations
// @Accept json
// @Param coursePhaseID path string true "Course phase ID"
// @Param evaluation body evaluationDTO.CreateOrUpdateEvaluationRequest true "Evaluation payload"
// @Success 201 {string} string "Created"
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/evaluation [post]
func createOrUpdateEvaluation(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var request evaluationDTO.CreateOrUpdateEvaluationRequest
	if err := c.BindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	statusCode, err := utils.ValidateStudentOwnership(c, request.AuthorCourseParticipationID)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": "Students can only create evaluations as the author"})
		return
	}

	err = CreateOrUpdateEvaluation(c, coursePhaseID, request)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusCreated)
}

// deleteEvaluation godoc
// @Summary Delete evaluation
// @Description Delete an evaluation by ID.
// @Tags evaluations
// @Param coursePhaseID path string true "Course phase ID"
// @Param evaluationID path string true "Evaluation ID"
// @Success 200 {string} string "OK"
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/evaluation/{evaluationID} [delete]
func deleteEvaluation(c *gin.Context) {
	evaluationID, err := uuid.Parse(c.Param("evaluationID"))
	if err != nil {
		log.Error("Error parsing evaluationID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipationID, er := utils.GetUserCourseParticipationID(c)
	if er != nil {
		log.Error("Error getting student courseParticipationID: ", er)
		handleError(c, utils.GetUserCourseParticipationIDErrorStatus(er), er)
		return
	}

	// Ensure the user is the author of the evaluation or has the right permissions
	if !isEvaluationAuthor(c, evaluationID, courseParticipationID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this evaluation"})
		return
	}

	err = DeleteEvaluation(c, evaluationID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

func isEvaluationAuthor(c *gin.Context, evaluationID, authorID uuid.UUID) bool {
	evaluation, err := GetEvaluationByID(c, evaluationID)
	if err != nil {
		log.Error("Error fetching evaluation: ", err)
		return false
	}

	return evaluation.AuthorCourseParticipationID == authorID
}

func handleError(c *gin.Context, statusCode int, err error) {
	log.Error(err)
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
