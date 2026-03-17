package feedbackItem

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/feedbackItem/feedbackItemDTO"
	"github.com/prompt-edu/prompt/servers/assessment/utils"
	log "github.com/sirupsen/logrus"
)

// setupFeedbackItemRouter sets up feedback item endpoints.
// @Summary Feedback Item Endpoints
// @Description Manage feedback items for evaluations.
// @Tags feedback_items
// @Security BearerAuth
func setupFeedbackItemRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	feedbackItemRouter := routerGroup.Group("/evaluation/feedback-items")

	feedbackItemRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), listFeedbackItemsForCoursePhase)
	feedbackItemRouter.GET("/course-participation/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getFeedbackItemsForParticipantInPhase)
	feedbackItemRouter.GET("/tutor/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getFeedbackItemsForTutorInPhase)

	// Student endpoints - access to own feedback items only
	feedbackItemRouter.GET("/my-feedback", authMiddleware(promptSDK.CourseStudent), getMyFeedbackItems)
	feedbackItemRouter.POST("", authMiddleware(promptSDK.CourseStudent), createFeedbackItem)
	feedbackItemRouter.PUT("/:feedbackItemID", authMiddleware(promptSDK.CourseStudent), updateFeedbackItem)
	feedbackItemRouter.DELETE("/:feedbackItemID", authMiddleware(promptSDK.CourseStudent), deleteFeedbackItem)
}

// listFeedbackItemsForCoursePhase godoc
// @Summary List feedback items for course phase
// @Description List feedback items for a course phase.
// @Tags feedback_items
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} feedbackItemDTO.FeedbackItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/evaluation/feedback-items [get]
func listFeedbackItemsForCoursePhase(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	feedbackItems, err := ListFeedbackItemsForCoursePhase(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, feedbackItems)
}

// getFeedbackItemsForParticipantInPhase godoc
// @Summary List feedback items for participant
// @Description List feedback items for a course participation in a course phase.
// @Tags feedback_items
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param courseParticipationID path string true "Course participation ID"
// @Success 200 {array} feedbackItemDTO.FeedbackItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/evaluation/feedback-items/course-participation/{courseParticipationID} [get]
func getFeedbackItemsForParticipantInPhase(c *gin.Context) {
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

	feedbackItems, err := ListFeedbackItemsForParticipantInPhase(c, courseParticipationID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, feedbackItems)
}

// getFeedbackItemsForTutorInPhase godoc
// @Summary List feedback items for tutor
// @Description List feedback items for a tutor in a course phase.
// @Tags feedback_items
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param courseParticipationID path string true "Tutor course participation ID"
// @Success 200 {array} feedbackItemDTO.FeedbackItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/evaluation/feedback-items/tutor/{courseParticipationID} [get]
func getFeedbackItemsForTutorInPhase(c *gin.Context) {
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

	feedbackItems, err := ListFeedbackItemsForTutorInPhase(c, courseParticipationID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, feedbackItems)
}

// getMyFeedbackItems godoc
// @Summary List my feedback items
// @Description List feedback items authored by the current student.
// @Tags feedback_items
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} feedbackItemDTO.FeedbackItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/evaluation/feedback-items/my-feedback [get]
func getMyFeedbackItems(c *gin.Context) {
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

	feedbackItems, err := ListFeedbackItemsByAuthorInPhase(c, courseParticipationID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, feedbackItems)
}

// createFeedbackItem godoc
// @Summary Create feedback item
// @Description Create a feedback item for the current student.
// @Tags feedback_items
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param feedbackItem body feedbackItemDTO.CreateFeedbackItemRequest true "Feedback item payload"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/evaluation/feedback-items [post]
func createFeedbackItem(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var req feedbackItemDTO.CreateFeedbackItemRequest
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipationID, err := utils.GetUserCourseParticipationID(c)
	if err != nil {
		handleError(c, utils.GetUserCourseParticipationIDErrorStatus(err), err)
		return
	}

	// Students can only create feedback items where they are the author
	if req.AuthorCourseParticipationID != courseParticipationID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Students can only create feedback items as the author"})
		return
	}

	req.CoursePhaseID = coursePhaseID

	err = CreateFeedbackItem(c, req)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Feedback item created successfully"})
}

// updateFeedbackItem godoc
// @Summary Update feedback item
// @Description Update a feedback item.
// @Tags feedback_items
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param feedbackItemID path string true "Feedback item ID"
// @Param feedbackItem body feedbackItemDTO.UpdateFeedbackItemRequest true "Feedback item payload"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/evaluation/feedback-items/{feedbackItemID} [put]
func updateFeedbackItem(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var req feedbackItemDTO.UpdateFeedbackItemRequest
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	statusCode, err := utils.ValidateStudentOwnership(c, req.AuthorCourseParticipationID)
	if err != nil {
		handleError(c, statusCode, err)
		return
	}

	req.CoursePhaseID = coursePhaseID

	err = UpdateFeedbackItem(c, req)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Feedback item updated successfully"})
}

// deleteFeedbackItem godoc
// @Summary Delete feedback item
// @Description Delete a feedback item by ID.
// @Tags feedback_items
// @Param coursePhaseID path string true "Course phase ID"
// @Param feedbackItemID path string true "Feedback item ID"
// @Success 200 {string} string "OK"
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/evaluation/feedback-items/{feedbackItemID} [delete]
func deleteFeedbackItem(c *gin.Context) {
	feedbackItemID, err := uuid.Parse(c.Param("feedbackItemID"))
	if err != nil {
		log.Error("Error parsing feedbackItemID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipationID, err := utils.GetUserCourseParticipationID(c)
	if err != nil {
		handleError(c, utils.GetUserCourseParticipationIDErrorStatus(err), err)
		return
	}
	if !IsFeedbackItemAuthor(c, feedbackItemID, courseParticipationID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this feedback item"})
		return
	}

	err = DeleteFeedbackItem(c, feedbackItemID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

func handleError(c *gin.Context, statusCode int, err error) {
	log.Error(err)
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
