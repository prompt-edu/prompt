package categoryAssessment

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/categoryAssessment/categoryAssessmentDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	log "github.com/sirupsen/logrus"
)

// setupCategoryAssessmentRouter sets up category-assessment endpoints.
// @Summary Category Assessment Endpoints
// @Description Manage per-category free-text comments for student assessments.
// @Tags categoryAssessments
// @Security BearerAuth
func setupCategoryAssessmentRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	r := routerGroup.Group("/category-assessment")

	r.POST("", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer, promptSDK.CourseEditor), createOrUpdateCategoryAssessment)
	r.DELETE("/:categoryAssessmentID", authMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer), deleteCategoryAssessment)
}

// createOrUpdateCategoryAssessment godoc
// @Summary Create or update category assessment comment
// @Tags categoryAssessments
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param body body categoryAssessmentDTO.CreateOrUpdateCategoryAssessmentRequest true "Payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/category-assessment [post]
func createOrUpdateCategoryAssessment(c *gin.Context) {
	var req categoryAssessmentDTO.CreateOrUpdateCategoryAssessmentRequest
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	if err := CreateOrUpdateCategoryAssessment(c, req); err != nil {
		if errors.Is(err, assessmentCompletion.ErrAssessmentCompleted) ||
			errors.Is(err, coursePhaseConfig.ErrNotStarted) ||
			errors.Is(err, coursePhaseConfig.ErrDeadlinePassed) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Category assessment created/updated successfully"})
}

// deleteCategoryAssessment godoc
// @Summary Delete category assessment
// @Tags categoryAssessments
// @Param coursePhaseID path string true "Course phase ID"
// @Param categoryAssessmentID path string true "Category assessment ID"
// @Success 200 {string} string "OK"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/category-assessment/{categoryAssessmentID} [delete]
func deleteCategoryAssessment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("categoryAssessmentID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	if err := DeleteCategoryAssessment(c, id); err != nil {
		if errors.Is(err, assessmentCompletion.ErrAssessmentCompleted) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.String(http.StatusOK, "OK")
}

func handleError(c *gin.Context, statusCode int, err error) {
	log.Error(err)
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
