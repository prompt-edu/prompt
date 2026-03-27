package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/auth/service"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
)

// setupCoursePhaseAuthRouter sets up the course phase auth endpoints
// @Summary Course Phase Auth Endpoints
// @Description Endpoints for course phase authentication and participation
// @Tags auth
// @Security BearerAuth
func setupCoursePhaseAuthRouter(router *gin.RouterGroup, authMiddleware func() gin.HandlerFunc, permissionIDMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	auth := router.Group("/auth", authMiddleware())

	coursePhaseauth := auth.Group("/course_phase/:coursePhaseID", authMiddleware())
	// this endpoint could also be exposed without any authentication
	coursePhaseauth.GET("/roles", getCoursePhaseAuthRoles)
	// returns a 401 if the user is not a student of the course
	coursePhaseauth.GET("is_student", permissionIDMiddleware(permissionValidation.CourseStudent), getCoursePhaseParticipation)

  auth.GET("/subject_identifiers", getSubjectIdentifiers)
}

// getCoursePhaseAuthRoles godoc
// @Summary Get course phase roles
// @Description Get the role mapping for a course phase
// @Tags auth
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} authDTO.GetCourseRoles
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/course_phase/{coursePhaseID}/roles [get]
func getCoursePhaseAuthRoles(c *gin.Context) {
	// Get the course phase ID from the URL
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid course phase ID"})
		return
	}

	roleMapping, err := service.GetCourseRoles(c, coursePhaseID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to get course roles"})
		return
	}

	c.JSON(http.StatusOK, roleMapping)
}

// getCoursePhaseParticipation godoc
// @Summary Get course phase participation
// @Description Check if the user is a student of the course phase
// @Tags auth
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} authDTO.GetCoursePhaseParticipation
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/course_phase/{coursePhaseID}/is_student [get]
func getCoursePhaseParticipation(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid course phase ID"})
		return
	}

	matriculationNumber := c.GetString("matriculationNumber")
	universityLogin := c.GetString("universityLogin")

	if matriculationNumber == "" || universityLogin == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "missing matriculation number or university login"})
		return
	}

	participation, err := service.GetCoursePhaseParticipation(c, coursePhaseID, matriculationNumber, universityLogin)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to get course phase participation"})
		return
	}

	c.JSON(http.StatusOK, participation)
}

// getSubjectIdentifiers godoc
// @Summary Get subject identifiers
// @Description For the user with the passed auth token, get the subject identifiers: userID, studentID, courseParticipationIDs
// @Tags auth
// @Produce json
// @Success 200 {object} keycloakCoreRequests.SubjectIdentifiers
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/subject_identifiers [get]
func getSubjectIdentifiers(c *gin.Context) {

  subjectIdentifiers, err := service.GetSubjectIdentifiers(c)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "failed to get subject identifiers"})
		return
	}

  c.JSON(http.StatusOK, subjectIdentifiers)
}
