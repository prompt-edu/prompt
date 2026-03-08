package copy

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/course/copy/courseCopyDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/utils"
	log "github.com/sirupsen/logrus"
)

// Id Middleware for all routes with a course id
// Role middleware for all without id -> possible additional filtering in subroutes required
// setupCourseRouter sets up the course endpoints
// @Summary Course Endpoints
// @Description Endpoints for managing courses and course graphs
// @Tags courses
// @Security BearerAuth
func setupCourseCopyRouter(router *gin.RouterGroup, authMiddleware func() gin.HandlerFunc, permissionRoleMiddleware, permissionIDMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	course := router.Group("/courses", authMiddleware())

	course.GET("/:uuid/copyable", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), checkCourseCopyable)
	course.POST("/:uuid/copy", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), copyCourse)
}

// copyCourse godoc
// @Summary Copy a course
// @Description Copy a course by UUID
// @Tags courses
// @Produce json
// @Param uuid path string true "Course UUID"
// @Success 201 {object} courseDTO.Course
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/copy [post]
func copyCourse(c *gin.Context) {
	userID := c.GetString("userID")

	courseVariables := courseCopyDTO.CopyCourseRequest{}
	if err := c.BindJSON(&courseVariables); err != nil {
		handleError(c, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	originalCourseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, fmt.Errorf("invalid course UUID: %w", err))
		return
	}

	newCourse, err := CopyCourse(c, originalCourseID, courseVariables, userID)
	if err != nil {
		log.Error("Copy course failed: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, newCourse)
}

// checkCourseCopyable godoc
// @Summary Check if a course is copyable
// @Description Returns whether the course can be copied based on the availability of the /copy endpoint in all course phases
// @Tags courses
// @Produce json
// @Param uuid path string true "Course UUID"
// @Success 200 {object} courseCopyDTO.CheckCourseCopyableResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/copyable [get]
func checkCourseCopyable(c *gin.Context) {
	originalCourseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, fmt.Errorf("invalid course UUID: %w", err))
		return
	}

	missing, err := CheckAllCoursePhasesCopyable(c, originalCourseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, fmt.Errorf("check failed: %w", err))
		return
	}

	if len(missing) > 0 {
		c.JSON(http.StatusOK, courseCopyDTO.CheckCourseCopyableResponse{
			Copyable:          false,
			MissingPhaseTypes: missing,
		})
		return
	}
	c.JSON(http.StatusOK, courseCopyDTO.CheckCourseCopyableResponse{
		Copyable:          true,
		MissingPhaseTypes: []string{},
	})
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, utils.ErrorResponse{
		Error: err.Error(),
	})
}
