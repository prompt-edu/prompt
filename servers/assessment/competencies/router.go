package competencies

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/competencies/competencyDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

// setupCompetencyRouter sets up competency endpoints.
// @Summary Competency Endpoints
// @Description Manage competencies for assessment categories.
// @Tags competencies
// @Security BearerAuth
func setupCompetencyRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	competencyRouter := routerGroup.Group("/competency")

	competencyRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), listCompetencies)
	competencyRouter.GET("/:competencyID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getCompetency)
	competencyRouter.GET("/category/:categoryID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), listCompetenciesByCategory)
	competencyRouter.POST("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), createCompetency)
	competencyRouter.PUT("/:competencyID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), updateCompetency)
	competencyRouter.DELETE("/:competencyID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), deleteCompetency)
}

// listCompetencies godoc
// @Summary List competencies
// @Description List all competencies for the course phase.
// @Tags competencies
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} competencyDTO.Competency
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/competency [get]
func listCompetencies(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	competencies, err := ListCompetenciesForCoursePhase(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, competencyDTO.GetCompetencyDTOsFromDBModels(competencies))
}

// getCompetency godoc
// @Summary Get competency
// @Description Get a competency by ID.
// @Tags competencies
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param competencyID path string true "Competency ID"
// @Success 200 {object} competencyDTO.Competency
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/competency/{competencyID} [get]
func getCompetency(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	competencyID, err := uuid.Parse(c.Param("competencyID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	competency, err := GetCompetencyForCoursePhase(c, coursePhaseID, competencyID)
	if err != nil {
		if errors.Is(err, assessmentSchemas.ErrSchemaNotAccessible) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, competencyDTO.GetCompetencyDTOsFromDBModels([]db.Competency{competency})[0])
}

// listCompetenciesByCategory godoc
// @Summary List competencies by category
// @Description List competencies for a category.
// @Tags competencies
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param categoryID path string true "Category ID"
// @Success 200 {array} competencyDTO.Competency
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/competency/category/{categoryID} [get]
func listCompetenciesByCategory(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	categoryID, err := uuid.Parse(c.Param("categoryID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	competencies, err := ListCompetenciesByCategoryForCoursePhase(c, coursePhaseID, categoryID)
	if err != nil {
		if errors.Is(err, assessmentSchemas.ErrSchemaNotAccessible) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, competencyDTO.GetCompetencyDTOsFromDBModels(competencies))
}

// createCompetency godoc
// @Summary Create competency
// @Description Create a competency for the course phase.
// @Tags competencies
// @Accept json
// @Param coursePhaseID path string true "Course phase ID"
// @Param competency body competencyDTO.CreateCompetencyRequest true "Competency payload"
// @Success 201 {string} string "Created"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/competency [post]
func createCompetency(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var req competencyDTO.CreateCompetencyRequest
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = CreateCompetency(c, coursePhaseID, req)
	if err != nil {
		if errors.Is(err, assessmentSchemas.ErrSchemaNotAccessible) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusCreated)
}

// updateCompetency godoc
// @Summary Update competency
// @Description Update a competency.
// @Tags competencies
// @Accept json
// @Param coursePhaseID path string true "Course phase ID"
// @Param competencyID path string true "Competency ID"
// @Param competency body competencyDTO.UpdateCompetencyRequest true "Competency payload"
// @Success 200 {string} string "OK"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/competency/{competencyID} [put]
func updateCompetency(c *gin.Context) {
	competencyID, err := uuid.Parse(c.Param("competencyID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var req competencyDTO.UpdateCompetencyRequest
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = UpdateCompetency(c, competencyID, coursePhaseID, req)
	if err != nil {
		if errors.Is(err, assessmentSchemas.ErrSchemaNotAccessible) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

// deleteCompetency godoc
// @Summary Delete competency
// @Description Delete a competency.
// @Tags competencies
// @Param coursePhaseID path string true "Course phase ID"
// @Param competencyID path string true "Competency ID"
// @Success 200 {string} string "OK"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/competency/{competencyID} [delete]
func deleteCompetency(c *gin.Context) {
	competencyID, err := uuid.Parse(c.Param("competencyID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = DeleteCompetency(c, competencyID, coursePhaseID)
	if err != nil {
		if errors.Is(err, assessmentSchemas.ErrSchemaNotAccessible) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
