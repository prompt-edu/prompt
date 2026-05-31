package categoryAssessment

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/categoryAssessment/categoryAssessmentDTO"
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
}

// createOrUpdateCategoryAssessment godoc
// @Summary Create or update category assessment comment
// @Description Upserts the free-text comment for a (category, student) within the course phase. The author identity is taken from the authenticated JWT and any client-sent author fields are ignored.
// @Tags categoryAssessments
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param body body categoryAssessmentDTO.CreateOrUpdateCategoryAssessmentRequest true "Payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/category-assessment [post]
func createOrUpdateCategoryAssessment(c *gin.Context) {
	var req categoryAssessmentDTO.CreateOrUpdateCategoryAssessmentRequest
	if err := c.BindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	tokenUser, ok := keycloakTokenVerifier.GetTokenUser(c)
	if !ok {
		handleError(c, http.StatusUnauthorized, errors.New("authenticated user not found in context"))
		return
	}
	// Server is the source of truth for author identity; ignore anything the client supplied.
	req.Author = tokenUser.FirstName + " " + tokenUser.LastName
	req.AuthorID = tokenUser.ID

	if err := CreateOrUpdateCategoryAssessment(c, req); err != nil {
		if errors.Is(err, ErrNotEditable) {
			handleError(c, http.StatusForbidden, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Category assessment created/updated successfully"})
}

func handleError(c *gin.Context, statusCode int, err error) {
	log.Error(err)
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
