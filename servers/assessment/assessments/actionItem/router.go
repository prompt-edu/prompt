package actionItem

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/actionItem/actionItemDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	"github.com/prompt-edu/prompt/servers/assessment/utils"
	log "github.com/sirupsen/logrus"
)

// setupActionItemRouter sets up action item endpoints.
// @Summary Action Item Endpoints
// @Description Manage action items for assessments.
// @Tags action_items
// @Security BearerAuth
func setupActionItemRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	actionItemRouter := routerGroup.Group("student-assessment/action-item")

	// course phase communication
	actionItemRouter.GET("/action", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getAllActionItemsForCoursePhaseCommunication)
	actionItemRouter.GET("/action/course-participation/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getStudentActionItemsForCoursePhaseCommunication)

	// action item management
	actionItemRouter.GET("/course-participation/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getActionItemsForStudent)
	actionItemRouter.POST("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), createActionItem)
	actionItemRouter.PUT("/:id", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), updateActionItem)
	actionItemRouter.DELETE("/:id", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), deleteActionItem)

	// student access to own action items
	actionItemRouter.GET("/my-action-items", authMiddleware(promptSDK.CourseStudent), getMyActionItems)
}

// getAllActionItemsForCoursePhaseCommunication godoc
// @Summary List action items for course phase communication
// @Description List all action items for a course phase.
// @Tags action_items
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} actionItemDTO.ActionItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/action-item/action [get]
func getAllActionItemsForCoursePhaseCommunication(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	actionItems, err := GetAllActionItemsForCoursePhaseCommunication(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, actionItems)
}

// getStudentActionItemsForCoursePhaseCommunication godoc
// @Summary List action items for a student (communication)
// @Description List action items for a course participation in a course phase.
// @Tags action_items
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param courseParticipationID path string true "Course participation ID"
// @Success 200 {array} actionItemDTO.ActionItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/action-item/action/course-participation/{courseParticipationID} [get]
func getStudentActionItemsForCoursePhaseCommunication(c *gin.Context) {
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

	actionItems, err := GetStudentActionItemsForCoursePhaseCommunication(c, courseParticipationID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, actionItems)
}

// createActionItem godoc
// @Summary Create action item
// @Description Create a new action item.
// @Tags action_items
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param actionItem body actionItemDTO.CreateActionItemRequest true "Action item payload"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/action-item [post]
func createActionItem(c *gin.Context) {
	var req actionItemDTO.CreateActionItemRequest
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	err := CreateActionItem(c, req)
	if err != nil {
		if errors.Is(err, assessmentCompletion.ErrAssessmentCompleted) || errors.Is(err, coursePhaseConfig.ErrNotStarted) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Action item created successfully"})
}

// updateActionItem godoc
// @Summary Update action item
// @Description Update an action item.
// @Tags action_items
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param id path string true "Action item ID"
// @Param actionItem body actionItemDTO.UpdateActionItemRequest true "Action item payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/action-item/{id} [put]
func updateActionItem(c *gin.Context) {
	actionItemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var req actionItemDTO.UpdateActionItemRequest
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	// Ensure the ID from URL matches the one in the request
	req.ID = actionItemID

	err = UpdateActionItem(c, req)
	if err != nil {
		if errors.Is(err, assessmentCompletion.ErrAssessmentCompleted) || errors.Is(err, coursePhaseConfig.ErrNotStarted) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Action item updated successfully"})
}

// deleteActionItem godoc
// @Summary Delete action item
// @Description Delete an action item.
// @Tags action_items
// @Param coursePhaseID path string true "Course phase ID"
// @Param id path string true "Action item ID"
// @Success 200 {string} string "OK"
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/action-item/{id} [delete]
func deleteActionItem(c *gin.Context) {
	actionItemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = DeleteActionItem(c, actionItemID)
	if err != nil {
		if errors.Is(err, assessmentCompletion.ErrAssessmentCompleted) || errors.Is(err, coursePhaseConfig.ErrNotStarted) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

// getActionItemsForStudent godoc
// @Summary List action items for student
// @Description List action items for a course participation in the course phase.
// @Tags action_items
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param courseParticipationID path string true "Course participation ID"
// @Success 200 {array} actionItemDTO.ActionItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/action-item/course-participation/{courseParticipationID} [get]
func getActionItemsForStudent(c *gin.Context) {
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

	actionItems, err := ListActionItemsForStudentInPhase(c, courseParticipationID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, actionItems)
}

// getMyActionItems godoc
// @Summary List my action items
// @Description List action items for the current student.
// @Tags action_items
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} actionItemDTO.ActionItem
// @Success 204 {string} string "No Content"
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/student-assessment/action-item/my-action-items [get]
func getMyActionItems(c *gin.Context) {
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
	if !config.ActionItemsVisible {
		handleError(c, http.StatusForbidden, fmt.Errorf("action items are not visible to students"))
		return
	}

	if !config.ResultsReleased {
		c.JSON(http.StatusOK, make([]actionItemDTO.ActionItem, 0))
		return
	}

	courseParticipationID, err := utils.GetUserCourseParticipationID(c)
	if err != nil {
		handleError(c, utils.GetUserCourseParticipationIDErrorStatus(err), err)
		return
	}

	exists, err := assessmentCompletion.CheckAssessmentCompletionExists(c, courseParticipationID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	if exists {
		completion, err := assessmentCompletion.GetAssessmentCompletion(c, courseParticipationID, coursePhaseID)
		if err != nil {
			handleError(c, http.StatusInternalServerError, err)
			return
		}
		if !completion.Completed {
			c.Status(http.StatusNoContent)
			return
		}
	}

	actionItems, err := ListActionItemsForStudentInPhase(c, courseParticipationID, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, actionItems)
}

func handleError(c *gin.Context, statusCode int, err error) {
	log.Error(err)
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
