package coursePhaseType

import (
	"net/http"

	"github.com/gin-gonic/gin"
	sdk "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/coursePhaseType/coursePhaseTypeDTO"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

// SubjectIdentifierProvider resolves the authenticated user's subject identifiers.
type SubjectIdentifierProvider interface {
	GetSubjectIdentifiers(c *gin.Context) (sdk.SubjectIdentifiers, error)
}

type coursePhaseTypeHandler struct {
	service  *CoursePhaseTypeService
	subjects SubjectIdentifierProvider
}

// RegisterRoutes mounts the course phase type endpoints on the given router group.
// @Summary Course Phase Type Endpoints
// @Description Endpoints for retrieving course phase types
// @Tags course_phase_types
// @Security BearerAuth
func RegisterRoutes(router *gin.RouterGroup, service *CoursePhaseTypeService, subjects SubjectIdentifierProvider, authMiddleware func() gin.HandlerFunc) {
	handler := &coursePhaseTypeHandler{service: service, subjects: subjects}

	course := router.Group("/course_phase_types", authMiddleware())
	course.GET("", handler.getCoursePhaseTypes)
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
func (h *coursePhaseTypeHandler) getCoursePhaseTypes(c *gin.Context) {
	var (
		coursePhaseTypes []coursePhaseTypeDTO.CoursePhaseType
		err              error
	)

	if c.Query("for_self") == "true" {
		subject, subjErr := h.subjects.GetSubjectIdentifiers(c)
		if subjErr != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: subjErr.Error()})
			return
		}
		coursePhaseTypes, err = h.service.GetCoursePhaseTypesForStudent(c, subject.StudentID)
	} else {
		coursePhaseTypes, err = h.service.GetAllCoursePhaseTypes(c)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, coursePhaseTypes)
}
