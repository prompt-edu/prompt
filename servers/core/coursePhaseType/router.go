package coursePhaseType

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

// setupCoursePhaseTypeRouter sets up the course phase type endpoints
// @Summary Course Phase Type Endpoints
// @Description Endpoints for retrieving course phase types
// @Tags course_phase_types
// @Security BearerAuth
func setupCoursePhaseTypeRouter(router *gin.RouterGroup) {
	course := router.Group("/course_phase_types")
	course.GET("", getAllCoursePhaseTypes)
}

// getAllCoursePhaseTypes godoc
// @Summary Get all course phase types
// @Description Get a list of all available course phase types
// @Tags course_phase_types
// @Produce json
// @Success 200 {array} coursePhaseTypeDTO.CoursePhaseType
// @Failure 500 {object} utils.ErrorResponse
// @Router /course_phase_types [get]
func getAllCoursePhaseTypes(c *gin.Context) {
	coursePhaseTypes, err := GetAllCoursePhaseTypes(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.IndentedJSON(http.StatusOK, coursePhaseTypes)
}
