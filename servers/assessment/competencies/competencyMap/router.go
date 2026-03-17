package competencyMap

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/competencies/competencyMap/competencyMapDTO"
	log "github.com/sirupsen/logrus"
)

// setupCompetencyMapRouter sets up competency mapping endpoints.
// @Summary Competency Mapping Endpoints
// @Description Manage competency mappings.
// @Tags competency_mappings
// @Security BearerAuth
func setupCompetencyMapRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	competencyMapRouter := routerGroup.Group("/competency-mappings")

	competencyMapRouter.POST("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), createCompetencyMapping)
	competencyMapRouter.DELETE("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), deleteCompetencyMapping)
	competencyMapRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getAllCompetencyMappings)
	competencyMapRouter.GET("/from/:fromCompetencyID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getCompetencyMappings)
	competencyMapRouter.GET("/to/:toCompetencyID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getReverseCompetencyMappings)
	competencyMapRouter.GET("/evaluations/:fromCompetencyID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getEvaluationsByMappedCompetency)
}

// createCompetencyMapping godoc
// @Summary Create competency mapping
// @Description Create a mapping between competencies.
// @Tags competency_mappings
// @Accept json
// @Param coursePhaseID path string true "Course phase ID"
// @Param mapping body competencyMapDTO.CompetencyMapping true "Competency mapping payload"
// @Success 201 {object} map[string]string "Created"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/competency-mappings [post]
func createCompetencyMapping(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var req competencyMapDTO.CompetencyMapping
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = CreateCompetencyMapping(c, coursePhaseID, req)
	if err != nil {
		if errors.Is(err, ErrCompetencyNotInCoursePhase) {
			handleError(c, http.StatusBadRequest, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Competency mapping created successfully"})
}

// deleteCompetencyMapping godoc
// @Summary Delete competency mapping
// @Description Delete a mapping between competencies.
// @Tags competency_mappings
// @Accept json
// @Param coursePhaseID path string true "Course phase ID"
// @Param mapping body competencyMapDTO.CompetencyMapping true "Competency mapping payload"
// @Success 200 {object} map[string]string "OK"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/competency-mappings [delete]
func deleteCompetencyMapping(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var req competencyMapDTO.CompetencyMapping
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = DeleteCompetencyMapping(c, coursePhaseID, req)
	if err != nil {
		if errors.Is(err, ErrCompetencyNotInCoursePhase) {
			handleError(c, http.StatusBadRequest, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Competency mapping deleted successfully"})
}

// getAllCompetencyMappings godoc
// @Summary List competency mappings
// @Description List all competency mappings.
// @Tags competency_mappings
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} competencyMapDTO.CompetencyMapping
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/competency-mappings [get]
func getAllCompetencyMappings(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	mappings, err := GetAllCompetencyMappings(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, competencyMapDTO.GetCompetencyMappingsFromDBModels(mappings))
}

// getCompetencyMappings godoc
// @Summary List mappings from a competency
// @Description List mappings from the given competency.
// @Tags competency_mappings
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param fromCompetencyID path string true "Source competency ID"
// @Success 200 {array} competencyMapDTO.CompetencyMapping
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/competency-mappings/from/{fromCompetencyID} [get]
func getCompetencyMappings(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	fromCompetencyID, err := uuid.Parse(c.Param("fromCompetencyID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	mappings, err := GetCompetencyMappings(c, coursePhaseID, fromCompetencyID)
	if err != nil {
		if errors.Is(err, ErrCompetencyNotInCoursePhase) {
			handleError(c, http.StatusBadRequest, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, competencyMapDTO.GetCompetencyMappingsFromDBModels(mappings))
}

// getReverseCompetencyMappings godoc
// @Summary List mappings to a competency
// @Description List mappings to the given competency.
// @Tags competency_mappings
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param toCompetencyID path string true "Target competency ID"
// @Success 200 {array} competencyMapDTO.CompetencyMapping
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/competency-mappings/to/{toCompetencyID} [get]
func getReverseCompetencyMappings(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	toCompetencyID, err := uuid.Parse(c.Param("toCompetencyID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	mappings, err := GetReverseCompetencyMappings(c, coursePhaseID, toCompetencyID)
	if err != nil {
		if errors.Is(err, ErrCompetencyNotInCoursePhase) {
			handleError(c, http.StatusBadRequest, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, competencyMapDTO.GetCompetencyMappingsFromDBModels(mappings))
}

// getEvaluationsByMappedCompetency godoc
// @Summary List evaluations by mapped competency
// @Description List evaluations using the mapped competency relationships.
// @Tags competency_mappings
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param fromCompetencyID path string true "Source competency ID"
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/competency-mappings/evaluations/{fromCompetencyID} [get]
func getEvaluationsByMappedCompetency(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	fromCompetencyID, err := uuid.Parse(c.Param("fromCompetencyID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	evaluations, err := GetEvaluationsByMappedCompetency(c, coursePhaseID, fromCompetencyID)
	if err != nil {
		if errors.Is(err, ErrCompetencyNotInCoursePhase) {
			handleError(c, http.StatusBadRequest, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	// Note: You might want to convert these to DTOs if evaluation DTOs exist
	c.JSON(http.StatusOK, evaluations)
}

func handleError(c *gin.Context, statusCode int, err error) {
	log.Error(err)
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
