package course

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/course/courseDTO"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/utils"
	log "github.com/sirupsen/logrus"
)

// Id Middleware for all routes with a course id
// Role middleware for all without id -> possible additional filtering in subroutes required
// RegisterRoutes mounts the course endpoints on the given router group.
// @Summary Course Endpoints
// @Description Endpoints for managing courses and course graphs
// @Tags courses
// @Security BearerAuth
func RegisterRoutes(router *gin.RouterGroup, service *CourseService) {
	setupCourseRouter(router, service, keycloakTokenVerifier.KeycloakMiddleware, permissionValidation.CheckAccessControlByRole, checkAccessControlByIDWrapper)
}

// initializes the handler func with CheckCoursePermissions
func checkAccessControlByIDWrapper(allowedRoles ...string) gin.HandlerFunc {
	return permissionValidation.CheckAccessControlByID(permissionValidation.CheckCoursePermission, "uuid", allowedRoles...)
}

func setupCourseRouter(router *gin.RouterGroup, s *CourseService, authMiddleware func() gin.HandlerFunc, permissionRoleMiddleware, permissionIDMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	course := router.Group("/courses", authMiddleware())
	course.GET("/", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer, permissionValidation.CourseEditor, permissionValidation.CourseStudent), s.getAllCourses)
	course.GET("/:uuid", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer, permissionValidation.CourseEditor), s.getCourseByID)
	course.POST("/", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer), s.createCourse)
	course.PUT("/:uuid/phase_graph", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), s.updateCoursePhaseOrder)
	course.GET("/:uuid/phase_graph", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer, permissionValidation.CourseEditor), s.getCoursePhaseGraph)
	course.GET("/:uuid/participation_data_graph", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer, permissionValidation.CourseEditor), s.getParticipationDataGraph)
	course.PUT("/:uuid/participation_data_graph", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), s.updateParticipationDataGraph)
	course.GET("/:uuid/phase_data_graph", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer, permissionValidation.CourseEditor), s.getPhaseDataGraph)
	course.PUT("/:uuid/phase_data_graph", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), s.updatePhaseDataGraph)

	course.PUT("/:uuid/archive", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer, permissionValidation.CourseEditor), s.archiveCourse)

	course.PUT("/:uuid", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), s.updateCourseData)
	course.GET("/self", s.getOwnCourses)

	course.PUT("/:uuid/template", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), s.updateCourseTemplateStatus)
	course.GET("/:uuid/template", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), s.checkCourseTemplateStatus)

	course.GET("/template", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), s.getTemplateCourses)

	course.GET("/check-name", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer), s.checkCourseNameAvailability)

	course.DELETE("/:uuid", permissionIDMiddleware(permissionValidation.PromptAdmin, permissionValidation.CourseLecturer), s.deleteCourse)
}

// getOwnCourses godoc
// @Summary Get own courses
// @Description Get the course IDs for the current user
// @Tags courses
// @Produce json
// @Success 200 {array} uuid.UUID
// @Router /courses/self [get]
func (s *CourseService) getOwnCourses(c *gin.Context) {
	matriculationNumber := c.GetString("matriculationNumber")
	universityLogin := c.GetString("universityLogin")

	if matriculationNumber == "" || universityLogin == "" {
		// we need to ensure that it is still usable if you do not have a matriculation number or university login
		// i.e. prompt admins might not have a student role
		log.Debug("no matriculation number or university login found")
		c.IndentedJSON(http.StatusOK, []uuid.UUID{})
		return
	}

	courseIDs, err := s.GetOwnCourseIDs(c, matriculationNumber, universityLogin)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, courseIDs)
}

// getAllCourses godoc
// @Summary Get all courses
// @Description Get all courses accessible to the user
// @Tags courses
// @Produce json
// @Success 200 {array} courseDTO.CourseWithPhases
// @Failure 403 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/ [get]
func (s *CourseService) getAllCourses(c *gin.Context) {
	rolesVal, exists := c.Get("userRoles")
	if !exists {
		handleError(c, http.StatusForbidden, errors.New("missing user roles"))
		return
	}

	userRoles := rolesVal.(map[string]bool)

	courses, err := s.GetAllCourses(c, userRoles)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, courses)
}

// getCourseByID godoc
// @Summary Get course by ID
// @Description Get a course by UUID
// @Tags courses
// @Produce json
// @Param uuid path string true "Course UUID"
// @Success 200 {object} courseDTO.CourseWithPhases
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid} [get]
func (s *CourseService) getCourseByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	course, err := s.GetCourseByID(c, id)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("failed to get course"))
		return
	}

	c.IndentedJSON(http.StatusOK, course)
}

// createCourse godoc
// @Summary Create a course
// @Description Create a new course
// @Tags courses
// @Accept json
// @Produce json
// @Param newCourse body courseDTO.CreateCourse true "Course to create"
// @Success 201 {object} courseDTO.Course
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/ [post]
func (s *CourseService) createCourse(c *gin.Context) {
	userID := c.GetString("userID")

	var newCourse courseDTO.CreateCourse
	if err := c.BindJSON(&newCourse); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if err := validateCreateCourse(newCourse); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	course, err := s.CreateCourse(c, newCourse, userID)
	if err != nil {
		if errors.Is(err, ErrDuplicateCourseIdentifier) {
			handleError(c, http.StatusConflict, err)
			return
		}
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("failed to create course"))
		return
	}
	c.IndentedJSON(http.StatusCreated, course)
}

// checkCourseNameAvailability godoc
// @Summary Check course name availability
// @Description Check if a course name is already taken for a given semester tag
// @Tags courses
// @Produce json
// @Param name query string true "Course name"
// @Param semesterTag query string true "Semester tag"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/check-name [get]
func (s *CourseService) checkCourseNameAvailability(c *gin.Context) {
	name := c.Query("name")
	semesterTag := c.Query("semesterTag")
	if name == "" || semesterTag == "" {
		handleError(c, http.StatusBadRequest, errors.New("name and semesterTag are required"))
		return
	}

	exists, err := s.CheckCourseNameExists(c, name, semesterTag)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("failed to check course name availability"))
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"exists": exists})
}

// getCoursePhaseGraph godoc
// @Summary Get course phase graph
// @Description Get the phase graph for a course
// @Tags courses
// @Produce json
// @Param uuid path string true "Course UUID"
// @Success 200 {array} courseDTO.CoursePhaseGraph
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/phase_graph [get]
func (s *CourseService) getCoursePhaseGraph(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	graph, err := s.GetCoursePhaseGraph(c, courseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, graph)
}

// getParticipationDataGraph godoc
// @Summary Get participation data graph
// @Description Get the participation data graph for a course
// @Tags courses
// @Produce json
// @Param uuid path string true "Course UUID"
// @Success 200 {array} courseDTO.MetaDataGraphItem
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/participation_data_graph [get]
func (s *CourseService) getParticipationDataGraph(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	graph, err := s.GetParticipationDataGraph(c, courseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, graph)
}

// getPhaseDataGraph godoc
// @Summary Get phase data graph
// @Description Get the phase data graph for a course
// @Tags courses
// @Produce json
// @Param uuid path string true "Course UUID"
// @Success 200 {array} courseDTO.MetaDataGraphItem
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/phase_data_graph [get]
func (s *CourseService) getPhaseDataGraph(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	graph, err := s.GetPhaseDataGraph(c, courseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, graph)
}

// updateCoursePhaseOrder godoc
// @Summary Update course phase order
// @Description Update the phase order for a course
// @Tags courses
// @Accept json
// @Produce json
// @Param uuid path string true "Course UUID"
// @Param graphUpdate body courseDTO.UpdateCoursePhaseGraph true "Phase graph update"
// @Success 200 {string} string "OK"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/phase_graph [put]
func (s *CourseService) updateCoursePhaseOrder(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var graphUpdate courseDTO.UpdateCoursePhaseGraph
	if err := c.BindJSON(&graphUpdate); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if err := validateUpdateCourseOrder(c, courseID, graphUpdate.PhaseGraph); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = s.UpdateCoursePhaseOrder(c, courseID, graphUpdate)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("failed to update course phase order"))
		return
	}

	c.Status(http.StatusOK)
}

// updateParticipationDataGraph godoc
// @Summary Update participation data graph
// @Description Update the participation data graph for a course
// @Tags courses
// @Accept json
// @Produce json
// @Param uuid path string true "Course UUID"
// @Param newGraph body []courseDTO.MetaDataGraphItem true "Participation data graph update"
// @Success 200 {string} string "OK"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/participation_data_graph [put]
func (s *CourseService) updateParticipationDataGraph(c *gin.Context) {
	newGraph, courseID, err := parseAndValidateMetaDataGraph(c)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = s.UpdateParticipationDataGraph(c, courseID, newGraph)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("failed to update meta data order"))
		return
	}

	c.Status(http.StatusOK)
}

// updatePhaseDataGraph godoc
// @Summary Update phase data graph
// @Description Update the phase data graph for a course
// @Tags courses
// @Accept json
// @Produce json
// @Param uuid path string true "Course UUID"
// @Param newGraph body []courseDTO.MetaDataGraphItem true "Phase data graph update"
// @Success 200 {string} string "OK"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/phase_data_graph [put]
func (s *CourseService) updatePhaseDataGraph(c *gin.Context) {
	newGraph, courseID, err := parseAndValidateMetaDataGraph(c)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = s.UpdatePhaseDataGraph(c, courseID, newGraph)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("failed to update meta data order"))
		return
	}

	c.Status(http.StatusOK)
}

func parseAndValidateMetaDataGraph(c *gin.Context) ([]courseDTO.MetaDataGraphItem, uuid.UUID, error) {
	courseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		return nil, uuid.UUID{}, err
	}

	var newGraph []courseDTO.MetaDataGraphItem
	if err := c.BindJSON(&newGraph); err != nil {
		return nil, uuid.UUID{}, err
	}

	if err := validateMetaDataGraph(c, courseID, newGraph); err != nil {
		return nil, uuid.UUID{}, err
	}

	return newGraph, courseID, nil
}

// archiveCourse godoc
// @Summary Archive or unarchive a course
// @Description Set archived=true (with archived_on=NOW()) or archived=false (with archived_on=NULL)
// @Tags courses
// @Accept json
// @Produce json
// @Param uuid path string true "Course UUID"
// @Param update body courseDTO.CourseArchiveStatus true "Archive status update"
// @Success 200 {object} courseDTO.Course "Updated course"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/archive [put]
func (s *CourseService) archiveCourse(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var update courseDTO.CourseArchiveStatus
	if err := c.BindJSON(&update); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	updatedCourse, err := s.UpdateCourseArchiveStatus(c, courseID, update.Archived)
	if err != nil {
		utils.RespondWithDBError(c, err)
		return
	}

	c.JSON(http.StatusOK, updatedCourse)
}

// updateCourseData godoc
// @Summary Update course data
// @Description Update the data for a course
// @Tags courses
// @Accept json
// @Produce json
// @Param uuid path string true "Course UUID"
// @Param update body courseDTO.UpdateCourseData true "Course data update"
// @Success 200 {string} string "OK"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid} [put]
func (s *CourseService) updateCourseData(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var update courseDTO.UpdateCourseData
	if err := c.BindJSON(&update); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = validateUpdateCourseData(update)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = s.UpdateCourseData(c, courseID, update)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("failed to update course data"))
		return
	}

	c.Status(http.StatusOK)
}

// deleteCourse godoc
// @Summary Delete a course
// @Description Delete a course by UUID
// @Tags courses
// @Produce json
// @Param uuid path string true "Course UUID"
// @Success 200 {string} string "OK"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid} [delete]
func (s *CourseService) deleteCourse(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = s.DeleteCourse(c, courseID)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("failed to delete course"))
		return
	}

	c.Status(http.StatusOK)
}

// updateCourseTemplateStatus godoc
// @Summary Update course template status
// @Description Update whether a course is marked as a template
// @Tags courses
// @Accept json
// @Produce json
// @Param uuid path string true "Course UUID"
// @Param update body courseDTO.CourseTemplateStatus true "Template status update"
// @Success 200 {string} string "OK"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/template [put]
func (s *CourseService) updateCourseTemplateStatus(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var update courseDTO.CourseTemplateStatus
	if err := c.BindJSON(&update); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = s.UpdateCourseTemplateStatus(c, courseID, update.IsTemplate)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("failed to update course template status"))
		return
	}

	c.Status(http.StatusOK)
}

// getTemplateCourses godoc
// @Summary Get template courses
// @Description Get all courses marked as templates accessible to the user
// @Tags courses
// @Produce json
// @Success 200 {array} courseDTO.Course
// @Failure 403 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/template [get]
func (s *CourseService) getTemplateCourses(c *gin.Context) {
	rolesVal, exists := c.Get("userRoles")
	if !exists {
		handleError(c, http.StatusForbidden, errors.New("missing user roles"))
		return
	}

	userRoles := rolesVal.(map[string]bool)

	courses, err := s.GetTemplateCourses(c, userRoles)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.IndentedJSON(http.StatusOK, courses)
}

// checkCourseTemplateStatus godoc
// @Summary Check course template status
// @Description Get the template status of a course
// @Tags courses
// @Produce json
// @Param uuid path string true "Course UUID"
// @Success 200 {object} courseDTO.CourseTemplateStatus
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/template [get]
func (s *CourseService) checkCourseTemplateStatus(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	isTemplate, err := s.CheckCourseTemplateStatus(c, courseID)
	if err != nil {
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("failed to check if course is template"))
		return
	}

	c.IndentedJSON(http.StatusOK, courseDTO.CourseTemplateStatus{
		IsTemplate: isTemplate,
	})
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, utils.ErrorResponse{
		Error: err.Error(),
	})
}
