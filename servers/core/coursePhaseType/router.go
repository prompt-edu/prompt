package coursePhaseType

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authService "github.com/prompt-edu/prompt/servers/core/auth/service"
	"github.com/prompt-edu/prompt/servers/core/coursePhaseType/coursePhaseTypeDTO"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

// setupCoursePhaseTypeRouter sets up the course phase type endpoints
// @Summary Course Phase Type Endpoints
// @Description Endpoints for retrieving course phase types
// @Tags course_phase_types
// @Security BearerAuth
func setupCoursePhaseTypeRouter(router *gin.RouterGroup, authMiddleware func() gin.HandlerFunc) {
	course := router.Group("/course_phase_types", authMiddleware())
	course.GET("", getCoursePhaseTypes)
}

// getCoursePhaseTypes godoc
// @Summary Get course phase types
// @Description Get all course phase types, or only those the authenticated user has been involved in when for_self=true.
// @Tags course_phase_types
// @Produce json
// @Param for_self query bool false "Restrict to phase types the authenticated user has been involved in"
// @Success 200 {array} coursePhaseTypeDTO.CoursePhaseType
// @Failure 500 {object} utils.ErrorResponse
// @Router /course_phase_types [get]
func getCoursePhaseTypes(c *gin.Context) {
	var (
		coursePhaseTypes []coursePhaseTypeDTO.CoursePhaseType
		err              error
	)

	if c.Query("for_self") == "true" {
		subject, subjErr := authService.GetSubjectIdentifiers(c)
		if subjErr != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: subjErr.Error()})
			return
		}
		coursePhaseTypes, err = GetCoursePhaseTypesForStudent(c, subject.StudentID)
	} else {
		coursePhaseTypes, err = GetAllCoursePhaseTypes(c)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, coursePhaseTypes)
}
