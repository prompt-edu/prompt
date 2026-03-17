package developerProfile

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/intro_course/developerProfile/developerProfileDTO"
	log "github.com/sirupsen/logrus"
)

func setupDeveloperProfileRouter(router *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	developerProfile := router.Group("/developer_profile")
	developerProfile.POST("", authMiddleware(promptSDK.CourseStudent), createDeveloperProfile)
	developerProfile.GET("/self", authMiddleware(promptSDK.CourseStudent), getOwnDeveloperProfile)
	// Getting all developer profiles is only allowed for lecturers
	developerProfile.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getAllDeveloperProfiles)
	developerProfile.PUT("/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), updateDeveloperProfile)

	// Export for the next phase
	devices := router.Group("/devices")
	devices.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getDevicesForAllParticipations)
	devices.GET("/:courseParticipationID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getDevicesForCourseParticipation)

}

// createDeveloperProfile godoc
// @Summary Create developer profile
// @Description Creates a developer profile for the authenticated student.
// @Tags developer-profile
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body developerProfileDTO.PostDeveloperProfile true "Developer profile payload"
// @Success 201
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/developer_profile [post]
func createDeveloperProfile(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	// get course participation id from context
	courseParticipationID, ok := c.Get("courseParticipationID")
	if !ok {
		log.Error("Error getting courseParticipationID from context")
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	var request developerProfileDTO.PostDeveloperProfile
	if err := c.BindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = validateDeveloperProfileUDIDs(request)
	if err != nil {
		log.Error("Error validating UDIDs: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = CreateDeveloperProfile(c, coursePhaseID, courseParticipationID.(uuid.UUID), request)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusCreated)
}

// getOwnDeveloperProfile godoc
// @Summary Get own developer profile
// @Description Returns the developer profile for the authenticated student.
// @Tags developer-profile
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} developerProfileDTO.DeveloperProfile
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/developer_profile/self [get]
func getOwnDeveloperProfile(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipationID, ok := c.Get("courseParticipationID")
	if !ok {
		log.Error("Error getting courseParticipationID from context")
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	developerProfile, err := GetOwnDeveloperProfile(c, coursePhaseID, courseParticipationID.(uuid.UUID))
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, developerProfile)
}

// getAllDeveloperProfiles godoc
// @Summary Get all developer profiles
// @Description Returns all developer profiles for the course phase.
// @Tags developer-profile
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} developerProfileDTO.DeveloperProfile
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/developer_profile [get]
func getAllDeveloperProfiles(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	developerProfiles, err := GetAllDeveloperProfiles(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, developerProfiles)
}

// updateDeveloperProfile godoc
// @Summary Update developer profile
// @Description Creates or updates a developer profile for a course participation.
// @Tags developer-profile
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param courseParticipationID path string true "Course Participation UUID"
// @Param request body developerProfileDTO.DeveloperProfile true "Developer profile payload"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/developer_profile/{courseParticipationID} [put]
func updateDeveloperProfile(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipationID, err := uuid.Parse(c.Param("courseParticipationID"))
	if err != nil {
		log.Error("Error parsing courseParticipationID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var devProfile developerProfileDTO.DeveloperProfile
	if err := c.BindJSON(&devProfile); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = CreateOrUpdateDeveloperProfile(c, coursePhaseID, courseParticipationID, devProfile)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

// getDevicesForAllParticipations godoc
// @Summary Get devices for all participations
// @Description Returns the device list for each course participation in the course phase.
// @Tags devices
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} developerProfileDTO.DeviceWithParticipationID
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/devices [get]
func getDevicesForAllParticipations(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	devices, err := GetDevicesForCoursePhase(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, devices)
}

// getDevicesForCourseParticipation godoc
// @Summary Get devices for course participation
// @Description Returns device identifiers for the specified course participation.
// @Tags devices
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param courseParticipationID path string true "Course Participation UUID"
// @Success 200 {array} string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/devices/{courseParticipationID} [get]
func getDevicesForCourseParticipation(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipationID, err := uuid.Parse(c.Param("courseParticipationID"))
	if err != nil {
		log.Error("Error parsing courseParticipationID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	devices, err := GetDevicesForCourseParticipation(c, coursePhaseID, courseParticipationID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "No devices found"})
		return
	} else if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, devices)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
