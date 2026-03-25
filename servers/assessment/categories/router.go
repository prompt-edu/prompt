package categories

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/categories/categoryDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	log "github.com/sirupsen/logrus"
)

// setupCategoryRouter sets up category endpoints.
// @Summary Category Endpoints
// @Description Manage assessment categories.
// @Tags categories
// @Security BearerAuth
func setupCategoryRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	categoryRouter := routerGroup.Group("/category")

	categoryRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getAllCategories)
	categoryRouter.GET("/assessment/with-competencies", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), getCategoriesWithCompetencies)
	categoryRouter.GET("/self/with-competencies", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), getSelfEvaluationCategoriesWithCompetencies)
	categoryRouter.GET("/peer/with-competencies", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), getPeerEvaluationCategoriesWithCompetencies)
	categoryRouter.GET("/tutor/with-competencies", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor, promptSDK.CourseStudent), getTutorEvaluationCategoriesWithCompetencies)

	categoryRouter.POST("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), createCategory)
	categoryRouter.PUT("/:categoryID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), updateCategory)
	categoryRouter.DELETE("/:categoryID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), deleteCategory)
}

// getAllCategories godoc
// @Summary List categories
// @Description List all categories for the course phase.
// @Tags categories
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} db.Category
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/category [get]
func getAllCategories(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	categories, err := ListCategoriesForCoursePhase(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, categories)
}

// createCategory godoc
// @Summary Create category
// @Description Create a new category for the course phase.
// @Tags categories
// @Accept json
// @Param coursePhaseID path string true "Course phase ID"
// @Param category body categoryDTO.CreateCategoryRequest true "Category payload"
// @Success 201 {string} string "Created"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/category [post]
func createCategory(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var request categoryDTO.CreateCategoryRequest
	if err := c.BindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	err = CreateCategory(c, coursePhaseID, request)
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

// updateCategory godoc
// @Summary Update category
// @Description Update a category for the course phase.
// @Tags categories
// @Accept json
// @Param coursePhaseID path string true "Course phase ID"
// @Param categoryID path string true "Category ID"
// @Param category body categoryDTO.UpdateCategoryRequest true "Category payload"
// @Success 200 {string} string "OK"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/category/{categoryID} [put]
func updateCategory(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	categoryID, err := uuid.Parse(c.Param("categoryID"))
	if err != nil {
		log.Error("Error parsing categoryID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var request categoryDTO.UpdateCategoryRequest
	if err := c.BindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = UpdateCategory(c, categoryID, coursePhaseID, request)
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

// deleteCategory godoc
// @Summary Delete category
// @Description Delete a category from the course phase.
// @Tags categories
// @Param coursePhaseID path string true "Course phase ID"
// @Param categoryID path string true "Category ID"
// @Success 200 {string} string "OK"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/category/{categoryID} [delete]
func deleteCategory(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	categoryID, err := uuid.Parse(c.Param("categoryID"))
	if err != nil {
		log.Error("Error parsing categoryID: ", err)
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = DeleteCategory(c, categoryID, coursePhaseID)
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

// getCategoriesWithCompetencies godoc
// @Summary List categories with competencies
// @Description List assessment categories with competencies for the course phase.
// @Tags categories
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} categoryDTO.CategoryWithCompetencies
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/category/assessment/with-competencies [get]
func getCategoriesWithCompetencies(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	config, err := coursePhaseConfig.GetCoursePhaseConfig(c, coursePhaseID)
	if err != nil {
		log.Error("Error getting course phase config: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	if isStudent(c) && (!config.ResultsReleased || !config.GradingSheetVisible) {
		handleError(c, http.StatusForbidden, fmt.Errorf("assessment results are not released yet"))
		return
	}

	result, err := GetCategoriesWithCompetencies(c, config.AssessmentSchemaID)
	if err != nil {
		log.Error("Error getting categories with competencies: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// getSelfEvaluationCategoriesWithCompetencies godoc
// @Summary List self-evaluation categories with competencies
// @Description List self-evaluation categories with competencies for the course phase.
// @Tags categories
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} categoryDTO.CategoryWithCompetencies
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/category/self/with-competencies [get]
func getSelfEvaluationCategoriesWithCompetencies(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	config, err := coursePhaseConfig.GetCoursePhaseConfig(c, coursePhaseID)
	if err != nil {
		log.Error("Error getting course phase config: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	result, err := GetCategoriesWithCompetencies(c, config.SelfEvaluationSchema)
	if err != nil {
		log.Error("Error getting self evaluation categories with competencies: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// getPeerEvaluationCategoriesWithCompetencies godoc
// @Summary List peer-evaluation categories with competencies
// @Description List peer-evaluation categories with competencies for the course phase.
// @Tags categories
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} categoryDTO.CategoryWithCompetencies
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/category/peer/with-competencies [get]
func getPeerEvaluationCategoriesWithCompetencies(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	config, err := coursePhaseConfig.GetCoursePhaseConfig(c, coursePhaseID)
	if err != nil {
		log.Error("Error getting course phase config: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	result, err := GetCategoriesWithCompetencies(c, config.PeerEvaluationSchema)
	if err != nil {
		log.Error("Error getting peer evaluation categories with competencies: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// getTutorEvaluationCategoriesWithCompetencies godoc
// @Summary List tutor-evaluation categories with competencies
// @Description List tutor-evaluation categories with competencies for the course phase.
// @Tags categories
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} categoryDTO.CategoryWithCompetencies
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/category/tutor/with-competencies [get]
func getTutorEvaluationCategoriesWithCompetencies(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	config, err := coursePhaseConfig.GetCoursePhaseConfig(c, coursePhaseID)
	if err != nil {
		log.Error("Error getting course phase config: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	result, err := GetCategoriesWithCompetencies(c, config.TutorEvaluationSchema)
	if err != nil {
		log.Error("Error getting tutor evaluation categories with competencies: ", err)
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}

func isStudent(c *gin.Context) bool {
	userRolesRaw, exists := c.Get("userRoles")
	if !exists {
		return false
	}
	userRoles, ok := userRolesRaw.(map[string]bool)
	if !ok {
		return false
	}
	return userRoles[promptSDK.CourseStudent]
}
