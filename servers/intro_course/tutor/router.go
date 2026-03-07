package tutor

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/intro_course/coreRequests"
	"github.com/prompt-edu/prompt/servers/intro_course/tutor/tutorDTO"
	log "github.com/sirupsen/logrus"
)

func setupTutorRouter(router *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	tutorRouter := router.Group("/tutor")
	tutorRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), getTutors)
	// we need the courseID to add students to keycloak groups
	tutorRouter.POST("/course/:courseID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), importTutors)
	tutorRouter.PUT("/:tutorID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), updateGitLabUsername)
}

// getTutors godoc
// @Summary Get tutors
// @Description Returns all tutors for the course phase.
// @Tags tutor
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {array} tutorDTO.Tutor
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/tutor [get]
func getTutors(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	tutors, err := GetTutors(c, coursePhaseID)
	if err != nil {
		log.Error("Error getting tutors: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, tutors)
}

// importTutors godoc
// @Summary Import tutors
// @Description Imports tutors and assigns them to the Keycloak groups.
// @Tags tutor
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param courseID path string true "Course UUID"
// @Param request body []tutorDTO.Tutor true "List of tutors to import"
// @Success 201
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/tutor/course/{courseID} [post]
func importTutors(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseID, err := uuid.Parse(c.Param("courseID"))
	if err != nil {
		log.Error("Error parsing courseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var tutors []tutorDTO.Tutor
	if err := c.BindJSON(&tutors); err != nil {
		log.Error("Error binding tutors: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	// Add Tutors to keycloak group
	tutorIDs := make([]uuid.UUID, len(tutors))
	for i, tutor := range tutors {
		tutorIDs[i] = tutor.ID
	}
	err = coreRequests.SendAddStudentsToKeycloakGroup(c.GetHeader("Authorization"), courseID, tutorIDs, KEYCLOAK_GROUP_NAME)
	if err != nil {
		log.Error("Error adding tutors to custom keycloak group: ", err)
		handleError(c, http.StatusInternalServerError, errors.New("error adding tutors to custom keycloak group"))
		return
	}
	err = coreRequests.SendAddStudentsToKeycloakGroup(c.GetHeader("Authorization"), courseID, tutorIDs, "editor")
	if err != nil {
		log.Error("Error adding tutors to editor keycloak group: ", err)
		handleError(c, http.StatusInternalServerError, errors.New("error adding tutors to editor keycloak group"))
		return
	}

	if err := ImportTutors(c, coursePhaseID, tutors); err != nil {
		log.Error("Error importing tutors: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusCreated)
}

// updateGitLabUsername godoc
// @Summary Update tutor GitLab username
// @Description Updates the GitLab username for a tutor.
// @Tags tutor
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param tutorID path string true "Tutor UUID"
// @Param request body tutorDTO.UpdateTutor true "GitLab username update"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/tutor/{tutorID} [put]
func updateGitLabUsername(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.Error("Error parsing coursePhaseID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	tutorID, err := uuid.Parse(c.Param("tutorID"))
	if err != nil {
		log.Error("Error parsing tutorID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var tutor tutorDTO.UpdateTutor
	if err := c.BindJSON(&tutor); err != nil {
		log.Error("Error binding tutor: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if err := UpdateGitLabUsername(c, coursePhaseID, tutorID, tutor); err != nil {
		log.Error("Error updating gitlab username: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
