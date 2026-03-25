package assessmentSchemas

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas/assessmentSchemaDTO"
	log "github.com/sirupsen/logrus"
)

// SetupAssessmentSchemaRouter sets up assessment schema endpoints.
// @Summary Assessment Schema Endpoints
// @Description Manage assessment schemas.
// @Tags assessment_schemas
// @Security BearerAuth
func SetupAssessmentSchemaRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	schemaRouter := routerGroup.Group("/assessment-schema")

	schemaRouter.GET("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getAllAssessmentSchemas)
	schemaRouter.GET("/:schemaID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), getAssessmentSchema)
	schemaRouter.GET("/:schemaID/has-assessment-data", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), checkSchemaHasAssessmentData)
	schemaRouter.POST("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), createAssessmentSchema)
	schemaRouter.PUT("/:schemaID", authMiddleware(promptSDK.PromptAdmin), updateAssessmentSchema)
	schemaRouter.DELETE("/:schemaID", authMiddleware(promptSDK.PromptAdmin), deleteAssessmentSchema)
}

// getAllAssessmentSchemas godoc
// @Summary List assessment schemas
// @Description List all assessment schemas.
// @Tags assessment_schemas
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} assessmentSchemaDTO.AssessmentSchema
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/assessment-schema [get]
func getAllAssessmentSchemas(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	schemas, err := ListAssessmentSchemasForCoursePhase(c, coursePhaseID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, schemas)
}

// getAssessmentSchema godoc
// @Summary Get assessment schema
// @Description Get an assessment schema by ID.
// @Tags assessment_schemas
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param schemaID path string true "Schema ID"
// @Success 200 {object} assessmentSchemaDTO.AssessmentSchema
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/assessment-schema/{schemaID} [get]
func getAssessmentSchema(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	schemaID, err := uuid.Parse(c.Param("schemaID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	schema, err := GetAssessmentSchemaForCoursePhase(c, coursePhaseID, schemaID)
	if err != nil {
		if errors.Is(err, ErrSchemaNotAccessible) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, schema)
}

// checkSchemaHasAssessmentData godoc
// @Summary Check schema assessment data
// @Description Check whether a schema has assessment data for the course phase.
// @Tags assessment_schemas
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param schemaID path string true "Schema ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/assessment-schema/{schemaID}/has-assessment-data [get]
func checkSchemaHasAssessmentData(c *gin.Context) {
	schemaID, err := uuid.Parse(c.Param("schemaID"))
	if err != nil {
		log.WithError(err).Error("Failed to parse schema ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schema ID"})
		return
	}

	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		log.WithError(err).Error("Failed to parse course phase ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course phase ID"})
		return
	}

	isAccessible, err := CheckSchemaAccessibleForCoursePhase(c, coursePhaseID, schemaID)
	if err != nil {
		log.WithError(err).Error("Failed to check schema accessibility")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check assessment data"})
		return
	}
	if !isAccessible {
		c.JSON(http.StatusForbidden, gin.H{"error": ErrSchemaNotAccessible.Error()})
		return
	}

	hasData, err := CheckPhaseHasAssessmentData(c, coursePhaseID, schemaID)
	if err != nil {
		log.WithError(err).Error("Failed to check if schema has assessment data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check assessment data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"hasAssessmentData": hasData})
}

// createAssessmentSchema godoc
// @Summary Create assessment schema
// @Description Create a new assessment schema.
// @Tags assessment_schemas
// @Accept json
// @Param coursePhaseID path string true "Course phase ID"
// @Param schema body assessmentSchemaDTO.CreateAssessmentSchemaRequest true "Assessment schema payload"
// @Success 201 {object} assessmentSchemaDTO.AssessmentSchema
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/assessment-schema [post]
func createAssessmentSchema(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var request assessmentSchemaDTO.CreateAssessmentSchemaRequest
	if err := c.BindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	schema, err := CreateAssessmentSchemaForCoursePhase(c, coursePhaseID, request)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, schema)
}

// updateAssessmentSchema godoc
// @Summary Update assessment schema
// @Description Update an assessment schema.
// @Tags assessment_schemas
// @Accept json
// @Param coursePhaseID path string true "Course phase ID"
// @Param schemaID path string true "Schema ID"
// @Param schema body assessmentSchemaDTO.UpdateAssessmentSchemaRequest true "Assessment schema payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/assessment-schema/{schemaID} [put]
func updateAssessmentSchema(c *gin.Context) {
	schemaID, err := uuid.Parse(c.Param("schemaID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var request assessmentSchemaDTO.UpdateAssessmentSchemaRequest
	if err := c.BindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = UpdateAssessmentSchema(c, schemaID, request)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Assessment schema updated successfully"})
}

// deleteAssessmentSchema godoc
// @Summary Delete assessment schema
// @Description Delete an assessment schema.
// @Tags assessment_schemas
// @Param coursePhaseID path string true "Course phase ID"
// @Param schemaID path string true "Schema ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/assessment-schema/{schemaID} [delete]
func deleteAssessmentSchema(c *gin.Context) {
	schemaID, err := uuid.Parse(c.Param("schemaID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	err = DeleteAssessmentSchema(c, schemaID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Assessment schema deleted successfully"})
}

func handleError(c *gin.Context, statusCode int, err error) {
	log.WithError(err).Error("Error in assessment schema handler")
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
