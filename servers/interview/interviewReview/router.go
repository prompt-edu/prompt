package interviewReview

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/interview/interviewReview/interviewReviewDTO"
	log "github.com/sirupsen/logrus"
)

func setupInterviewReviewRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	reviewRouter := routerGroup.Group("/interview-review")

	reviewRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getAllInterviewReviews)
	reviewRouter.PUT("/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), upsertInterviewReview)

	// Data-graph provided outputs consumed by downstream course phases.
	reviewRouter.GET("/score", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getScores)
	reviewRouter.GET("/scoreLevel", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getScoreLevels)
}

func handleError(c *gin.Context, statusCode int, err error) {
	log.Error(err)
	c.JSON(statusCode, gin.H{"error": err.Error()})
}

// getAllInterviewReviews godoc
// @Summary List interview reviews
// @Description List all interview reviews for the course phase
// @Tags interview-review
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} interviewReviewDTO.InterviewReview
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/interview-review [get]
func getAllInterviewReviews(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	reviews, err := GetInterviewReviews(c.Request.Context(), coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, reviews)
}

// upsertInterviewReview godoc
// @Summary Create or update an interview review
// @Description Store the score, interviewer and answers for a course participation
// @Tags interview-review
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param courseParticipationID path string true "Course Participation UUID"
// @Param request body interviewReviewDTO.UpdateInterviewReviewRequest true "Interview review"
// @Success 200 {object} interviewReviewDTO.InterviewReview
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/interview-review/{courseParticipationID} [put]
func upsertInterviewReview(c *gin.Context) {
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

	var req interviewReviewDTO.UpdateInterviewReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if req.Score != nil && (*req.Score < 1 || *req.Score > 5) {
		handleError(c, http.StatusBadRequest, errScoreOutOfRange)
		return
	}

	review, err := UpsertInterviewReview(c.Request.Context(), coursePhaseID, courseParticipationID, req)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, review)
}

// getScores godoc
// @Summary Interview score data-graph output
// @Description List interview scores keyed by course participation for downstream phases
// @Tags interview-review
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} interviewReviewDTO.ScoreWithParticipation
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/interview-review/score [get]
func getScores(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	scores, err := GetScores(c.Request.Context(), coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, scores)
}

// getScoreLevels godoc
// @Summary Interview score level data-graph output
// @Description List interview score levels keyed by course participation for downstream phases
// @Tags interview-review
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} interviewReviewDTO.ScoreLevelWithParticipation
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/interview-review/scoreLevel [get]
func getScoreLevels(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	scoreLevels, err := GetScoreLevels(c.Request.Context(), coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, scoreLevels)
}
