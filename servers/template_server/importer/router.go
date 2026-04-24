package importer

import (
	"net/http"

	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
)

func setupImporterRouter(routerGroup *gin.RouterGroup, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	importerRouter := routerGroup.Group("/import")

	// Current implementation is restricted to Prompt Admins because the importer
	// creates a course and immediately configures newly created phases. A regular
	// lecturer token may not yet contain the freshly assigned course role.
	importerRouter.POST(
		"/prompt-ready",
		authMiddleware(promptSDK.PromptAdmin),
		importPromptReadyTemplate,
	)
}

// importPromptReadyTemplate godoc
// @Summary Import a prompt-ready Team Allocation template
// @Description Create a template course, create Application and Team Allocation phases, wire applicationAnswers to Team Allocation, and initialize questions/skills/teams from a prompt-ready JSON payload.
// @Tags importer
// @Accept json
// @Produce json
// @Param payload body PromptReadyTemplate true "Prompt-ready template payload"
// @Success 201 {object} ImportPromptReadyTemplateResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /import/prompt-ready [post]
func importPromptReadyTemplate(c *gin.Context) {
	var payload PromptReadyTemplate
	if err := c.BindJSON(&payload); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	result, err := ImporterServiceSingleton.ImportPromptReadyTemplate(
		c.Request.Context(),
		c.GetHeader("Authorization"),
		payload,
	)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
