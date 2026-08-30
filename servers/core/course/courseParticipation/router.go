package courseParticipation

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/course/courseParticipation/courseParticipationDTO"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

// RegisterRoutes mounts the course participation endpoints on the given router group.
// @Summary Course Participation Endpoints
// @Description Endpoints for managing course participations
// @Tags course_participation
// @Security BearerAuth
func RegisterRoutes(routerGroup *gin.RouterGroup, service *CourseParticipationService) {
	setupCourseParticipationRouter(routerGroup, service, keycloakTokenVerifier.KeycloakMiddleware, checkAccessControlByIDWrapper)
}

// initializes the handler func with CheckCoursePermissions
func checkAccessControlByIDWrapper(allowedRoles ...string) gin.HandlerFunc {
	return permissionValidation.CheckAccessControlByID(permissionValidation.CheckCoursePermission, "uuid", allowedRoles...)
}

func setupCourseParticipationRouter(router *gin.RouterGroup, s *CourseParticipationService, authMiddleware func() gin.HandlerFunc, permissionIDMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	// incoming path should be /course/:uuid/
	courseParticipation := router.Group("/courses/:uuid/participations", authMiddleware())
	courseParticipation.GET("", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer, permissionValidation.CourseEditor), s.getCourseParticipationsForCourse)
	courseParticipation.POST("/enroll", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), s.createCourseParticipation)
	courseParticipation.GET("/self", s.getOwnCourseParticipation)
}

// getOwnCourseParticipation godoc
// @Summary Get own course participation
// @Description Get the participation of the current user in a course
// @Tags course_participation
// @Produce json
// @Param uuid path string true "Course UUID"
// @Success 200 {object} courseParticipationDTO.GetOwnCourseParticipation
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/participations/self [get]
func (s *CourseParticipationService) getOwnCourseParticipation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	matriculationNumber := c.GetString("matriculationNumber")
	universityLogin := c.GetString("universityLogin")

	if matriculationNumber == "" || universityLogin == "" {
		// potentially users without studentIDs are using the system -> no error shall be thrown
		c.JSON(http.StatusOK, courseParticipationDTO.GetOwnCourseParticipation{
			ID:                 uuid.Nil,
			ActiveCoursePhases: []uuid.UUID{},
		})
		return
	}

	courseParticipation, err := s.GetOwnCourseParticipation(c, id, matriculationNumber, universityLogin)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, courseParticipation)
}

// getCourseParticipationsForCourse godoc
// @Summary Get all participations for a course
// @Description Get all participations for a given course
// @Tags course_participation
// @Produce json
// @Param uuid path string true "Course UUID"
// @Success 200 {array} courseParticipationDTO.GetCourseParticipation
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/participations [get]
func (s *CourseParticipationService) getCourseParticipationsForCourse(c *gin.Context) {
	id, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipations, err := s.GetAllCourseParticipationsForCourse(c, id)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, courseParticipations)
}

// createCourseParticipation godoc
// @Summary Enroll in a course
// @Description Enroll a user in a course
// @Tags course_participation
// @Accept json
// @Produce json
// @Param uuid path string true "Course UUID"
// @Param newCourseParticipation body courseParticipationDTO.CreateCourseParticipation true "Participation to create"
// @Success 200 {object} courseParticipationDTO.GetCourseParticipation
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/participations/enroll [post]
func (s *CourseParticipationService) createCourseParticipation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var newCourseParticipation courseParticipationDTO.CreateCourseParticipation
	if err := c.BindJSON(&newCourseParticipation); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	// get course id from path
	newCourseParticipation.CourseID = id

	if err := Validate(newCourseParticipation); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	courseParticipation, err := s.CreateCourseParticipation(c, nil, newCourseParticipation)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, courseParticipation)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, utils.ErrorResponse{
		Error: err.Error(),
	})
}
