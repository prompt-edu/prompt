package coursePhase

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseDTO"
	"github.com/prompt-edu/prompt/servers/core/meta"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/utils"
	log "github.com/sirupsen/logrus"
)

// RegisterRoutes mounts the course phase endpoints on the given router group.
// @Summary Course Phase Endpoints
// @Description Endpoints for managing course phases
// @Tags course_phases
// @Security BearerAuth
func RegisterRoutes(router *gin.RouterGroup, service *CoursePhaseService, authMiddleware func() gin.HandlerFunc, checkCoursePhasePermission, checkCoursePermission permissionValidation.PermissionCheck) {
	setupCoursePhaseRouter(router, service, authMiddleware, checkAccessControlByIDWrapper(checkCoursePhasePermission), checkAccessControlByCourseIDWrapper(checkCoursePermission))
}

func checkAccessControlByIDWrapper(check permissionValidation.PermissionCheck) func(allowedRoles ...string) gin.HandlerFunc {
	return func(allowedRoles ...string) gin.HandlerFunc {
		return permissionValidation.CheckAccessControlByID(check, "uuid", allowedRoles...)
	}
}

func checkAccessControlByCourseIDWrapper(check permissionValidation.PermissionCheck) func(allowedRoles ...string) gin.HandlerFunc {
	return func(allowedRoles ...string) gin.HandlerFunc {
		return permissionValidation.CheckAccessControlByID(check, "courseID", allowedRoles...)
	}
}

func setupCoursePhaseRouter(router *gin.RouterGroup, s *CoursePhaseService, authMiddleware func() gin.HandlerFunc, permissionIDMiddleware, permissionCourseIDMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	coursePhase := router.Group("/course_phases", authMiddleware())
	coursePhase.GET("/:uuid", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer, permissionValidation.CourseEditor, permissionValidation.CourseStudent), s.getCoursePhaseByID)
	coursePhase.GET("/:uuid/course_phase_data", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer, permissionValidation.CourseEditor, permissionValidation.CourseStudent), s.getPrevPhaseDataByCoursePhaseID)
	coursePhase.GET("/:uuid/participation_status_counts", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer, permissionValidation.CourseEditor), s.getCoursePhaseParticipationStatusCounts)

	// getting the course ID here to do correct rights management
	coursePhase.POST("/course/:courseID", permissionCourseIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), s.createCoursePhase)
	coursePhase.PUT("/:uuid", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), s.updateCoursePhase)
	coursePhase.DELETE("/:uuid", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), s.deleteCoursePhase)
}

// createCoursePhase godoc
// @Summary Create a course phase
// @Description Create a new course phase for a course
// @Tags course_phases
// @Accept json
// @Produce json
// @Param courseID path string true "Course UUID"
// @Param newCoursePhase body coursePhaseDTO.CreateCoursePhase true "Course phase to create"
// @Success 201 {object} coursePhaseDTO.CoursePhase
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /course_phases/course/{courseID} [post]
func (s *CoursePhaseService) createCoursePhase(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("courseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var newCoursePhase coursePhaseDTO.CreateCoursePhase
	if err := c.BindJSON(&newCoursePhase); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	// validate that the courseIDs are identical
	if newCoursePhase.CourseID != courseID {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if err := validateCreateCoursePhase(newCoursePhase); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	coursePhase, err := s.CreateCoursePhase(c, newCoursePhase)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusCreated, coursePhase)

}

// getCoursePhaseByID godoc
// @Summary Get course phase by ID
// @Description Get a course phase by UUID
// @Tags course_phases
// @Produce json
// @Param uuid path string true "Course Phase UUID"
// @Success 200 {object} coursePhaseDTO.CoursePhase
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /course_phases/{uuid} [get]
func (s *CoursePhaseService) getCoursePhaseByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	coursePhase, err := s.GetCoursePhaseByID(c, id)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	// shadow the restricted data for students
	courseTokenIdentifier := c.GetString("courseTokenIdentifier")

	userRoles, exists := c.Get("userRoles")
	if !exists {
		log.Error("userRoles not found in context")
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	userRolesMap, ok := userRoles.(map[string]bool)
	if !ok {
		log.Error("invalid roles format in context")
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	if !hasRestrictedDataAccess(userRolesMap, courseTokenIdentifier) {
		// Hide restricted data for unauthorized users.
		coursePhase.RestrictedData = meta.MetaData{}
	}

	c.IndentedJSON(http.StatusOK, coursePhase)
}

// updateCoursePhase godoc
// @Summary Update a course phase
// @Description Update an existing course phase
// @Tags course_phases
// @Accept json
// @Produce json
// @Param updatedCoursePhase body coursePhaseDTO.UpdateCoursePhase true "Course phase to update"
// @Param uuid path string true "Course Phase UUID"
// @Success 200 {string} string "OK"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /course_phases/{uuid} [put]
func (s *CoursePhaseService) updateCoursePhase(c *gin.Context) {
	var updatedCoursePhase coursePhaseDTO.UpdateCoursePhase
	if err := c.BindJSON(&updatedCoursePhase); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if err := validateUpdateCoursePhase(updatedCoursePhase); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err := s.UpdateCoursePhase(c, updatedCoursePhase)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

// deleteCoursePhase godoc
// @Summary Delete a course phase
// @Description Delete a course phase by UUID
// @Tags course_phases
// @Produce json
// @Param uuid path string true "Course Phase UUID"
// @Success 200 {string} string "OK"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /course_phases/{uuid} [delete]
func (s *CoursePhaseService) deleteCoursePhase(c *gin.Context) {
	id, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = s.DeleteCoursePhase(c, id)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

// getPrevPhaseDataByCoursePhaseID godoc
// @Summary Get previous phase data by course phase ID
// @Description Get data from previous phases for a given course phase
// @Tags course_phases
// @Produce json
// @Param uuid path string true "Course Phase UUID"
// @Success 200 {object} coursePhaseDTO.PrevCoursePhaseData
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /course_phases/{uuid}/course_phase_data [get]
func (s *CoursePhaseService) getPrevPhaseDataByCoursePhaseID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	coursePhaseData, err := s.GetPrevPhaseDataByCoursePhaseID(c, id)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, coursePhaseData)
}

func hasRestrictedDataAccess(userRolesMap map[string]bool, courseTokenIdentifier string) bool {
	return userRolesMap[permissionValidation.PromptAdmin] ||
		userRolesMap[fmt.Sprintf("%s-%s", courseTokenIdentifier, permissionValidation.CourseLecturer)] ||
		userRolesMap[fmt.Sprintf("%s-%s", courseTokenIdentifier, permissionValidation.CourseEditor)]
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, utils.ErrorResponse{
		Error: err.Error(),
	})
}

// getCoursePhaseParticipationStatusCounts godoc
// @Summary Get course phase participation status counts
// @Description Get counts of participation statuses for a course phase
// @Tags course_phases
// @Produce json
// @Param uuid path string true "Course Phase UUID"
// @Success 200 {object} map[string]int "Status counts"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /course_phases/{uuid}/participation_status_counts [get]
func (s *CoursePhaseService) getCoursePhaseParticipationStatusCounts(c *gin.Context) {
	id, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	countsMap, err := s.GetCoursePhaseParticipationStatusCounts(c, id)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, countsMap)
}
