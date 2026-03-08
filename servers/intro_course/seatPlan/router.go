package seatPlan

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/intro_course/seatPlan/seatPlanDTO"
	log "github.com/sirupsen/logrus"
)

func setupSeatPlanRouter(router *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	seatPlanRouter := router.Group("/seat_plan")

	// Post initial seat plan with seat names
	seatPlanRouter.POST("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), createSeatPlan)
	seatPlanRouter.DELETE("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), deleteSeatPlan)

	// Update seat plan (assigned tutor, assigned student, hasMac, deviceID)
	seatPlanRouter.PUT("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), updateSeatPlan)

	seatPlanRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getSeatPlan)

	seatPlanRouter.GET("/own-assignment", authMiddleware(promptSDK.CourseStudent), getOwnSeatAssignment)

}

// createSeatPlan godoc
// @Summary Create seat plan
// @Description Creates the initial seat plan with seat names.
// @Tags seat-plan
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body []string true "Seat names"
// @Success 201
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/seat_plan [post]
func createSeatPlan(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var seatNames []string
	if err := c.BindJSON(&seatNames); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	// validate uniqueness of seat names
	unique := validateSeatNames(seatNames)
	if !unique {
		handleError(c, http.StatusBadRequest, errors.New("seat names must be unique"))
		return
	}

	err = CreateSeatPlan(c, coursePhaseID, seatNames)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusCreated)
}

// getSeatPlan godoc
// @Summary Get seat plan
// @Description Returns the seat plan for the course phase.
// @Tags seat-plan
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} seatPlanDTO.Seat
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/seat_plan [get]
func getSeatPlan(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	seats, err := GetSeatPlan(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, seats)
}

// updateSeatPlan godoc
// @Summary Update seat plan
// @Description Updates the seat plan assignments and device information.
// @Tags seat-plan
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body []seatPlanDTO.Seat true "Seat plan updates"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/seat_plan [put]
func updateSeatPlan(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var seats []seatPlanDTO.Seat
	if err := c.BindJSON(&seats); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = UpdateSeatPlan(c, coursePhaseID, seats)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

// deleteSeatPlan godoc
// @Summary Delete seat plan
// @Description Deletes the seat plan for the course phase.
// @Tags seat-plan
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/seat_plan [delete]
func deleteSeatPlan(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = DeleteSeatPlan(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

// getOwnSeatAssignment godoc
// @Summary Get own seat assignment
// @Description Returns the seat assignment for the authenticated student.
// @Tags seat-plan
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} seatPlanDTO.SeatAssignment
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/seat_plan/own-assignment [get]
func getOwnSeatAssignment(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipationID, ok := c.Get("courseParticipationID")
	if !ok {
		log.Error("Error getting courseParticipationID from context")
		handleError(c, http.StatusInternalServerError, errors.New("error getting courseParticipationID from context"))
		return
	}

	seatAssignment, err := GetOwnSeatAssignment(c, coursePhaseID, courseParticipationID.(uuid.UUID))

	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, seatAssignment)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
