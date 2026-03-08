package mailing

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/mailing/mailingDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

// setupMailingRouter sets up the mailing endpoints
// @Summary Mailing Endpoints
// @Description Endpoints for sending status mails
// @Tags mailing
// @Security BearerAuth
func setupMailingRouter(router *gin.RouterGroup, authMiddleware func() gin.HandlerFunc, permissionRoleMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	mailing := router.Group("/mailing", authMiddleware())
	mailing.PUT("/:coursePhaseID", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer, permissionValidation.CourseLecturer), sendStatusMailManualTrigger)
}

// sendStatusMailManualTrigger godoc
// @Summary Manually trigger status mail for a course phase
// @Description Sends a status mail for a given course phase ID
// @Tags mailing
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param mailingInfo body mailingDTO.SendStatusMail true "Mailing info"
// @Success 200 {object} mailingDTO.MailingReport
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /mailing/{coursePhaseID} [put]
func sendStatusMailManualTrigger(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var mailingInfo mailingDTO.SendStatusMail
	if err := c.BindJSON(&mailingInfo); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	response, err := SendStatusMailManualTrigger(c, coursePhaseID, mailingInfo.StatusMailToBeSend)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, utils.ErrorResponse{
		Error: err.Error(),
	})
}
