package interviewSlot

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
	interviewSlotDTO "github.com/prompt-edu/prompt/servers/interview/interviewSlot/interviewSlotDTO"
	log "github.com/sirupsen/logrus"
)

var _ db.InterviewSlot

func setupInterviewSlotRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	interviewRouter := routerGroup.Group("/interview-slots")

	// Admin routes - manage interview slots
	interviewRouter.POST("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), createInterviewSlot)
	interviewRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.PromptLecturer, promptSDK.CourseStudent), getAllInterviewSlots)
	interviewRouter.GET("/:slotId", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.PromptLecturer, promptSDK.CourseStudent), getInterviewSlot)
	interviewRouter.PUT("/:slotId", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), updateInterviewSlot)
	interviewRouter.DELETE("/:slotId", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), deleteInterviewSlot)
}

func writeServiceError(c *gin.Context, err error) {
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		c.JSON(serviceErr.StatusCode, gin.H{"error": serviceErr.Message})
		return
	}

	log.Errorf("Unexpected interview slot service error: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}

// createInterviewSlot godoc
// @Summary Create a new interview slot
// @Description Creates a new interview time slot for the course phase
// @Tags interview-slots
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body interviewSlotDTO.CreateInterviewSlotRequest true "Interview slot details"
// @Success 201 {object} db.InterviewSlot
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/interview-slots [post]
func createInterviewSlot(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	var req interviewSlotDTO.CreateInterviewSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slot, err := CreateInterviewSlot(c.Request.Context(), coursePhaseID, req)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, slot)
}

// getAllInterviewSlots godoc
// @Summary Get all interview slots for a course phase
// @Description Retrieves all interview slots with assignment details
// @Tags interview-slots
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} interviewSlotDTO.InterviewSlotResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/interview-slots [get]
func getAllInterviewSlots(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	slots, err := GetAllInterviewSlots(c.Request.Context(), coursePhaseID, c.GetHeader("Authorization"))
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, slots)
}

// getInterviewSlot godoc
// @Summary Get a specific interview slot
// @Description Retrieves details of a specific interview slot
// @Tags interview-slots
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param slotId path string true "Interview Slot UUID"
// @Success 200 {object} interviewSlotDTO.InterviewSlotResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/interview-slots/{slotId} [get]
func getInterviewSlot(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	slotID, err := uuid.Parse(c.Param("slotId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot ID"})
		return
	}

	slot, err := GetInterviewSlot(c.Request.Context(), coursePhaseID, slotID, c.GetHeader("Authorization"))
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, slot)
}

// updateInterviewSlot godoc
// @Summary Update an interview slot
// @Description Updates an existing interview slot
// @Tags interview-slots
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param slotId path string true "Interview Slot UUID"
// @Param request body interviewSlotDTO.UpdateInterviewSlotRequest true "Updated slot details"
// @Success 200 {object} db.InterviewSlot
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/interview-slots/{slotId} [put]
func updateInterviewSlot(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	slotID, err := uuid.Parse(c.Param("slotId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot ID"})
		return
	}

	var req interviewSlotDTO.UpdateInterviewSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slot, err := UpdateInterviewSlot(c.Request.Context(), coursePhaseID, slotID, req)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, slot)
}

// deleteInterviewSlot godoc
// @Summary Delete an interview slot
// @Description Deletes an interview slot and all its assignments
// @Tags interview-slots
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param slotId path string true "Interview Slot UUID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/interview-slots/{slotId} [delete]
func deleteInterviewSlot(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	slotID, err := uuid.Parse(c.Param("slotId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot ID"})
		return
	}

	if err := DeleteInterviewSlot(c.Request.Context(), coursePhaseID, slotID); err != nil {
		writeServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
