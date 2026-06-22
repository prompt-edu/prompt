package mailing

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/mailing/mailingDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

var sendManualMailFn = SendManualMailToParticipants

// setupMailingRouter sets up the mailing endpoints
// @Summary Mailing Endpoints
// @Description Endpoints for sending status mails
// @Tags mailing
// @Security BearerAuth
func setupMailingRouter(router *gin.RouterGroup, authMiddleware func() gin.HandlerFunc, permissionRoleMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	mailing := router.Group("/mailing", authMiddleware())
	mailing.PUT("/:coursePhaseID", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer, permissionValidation.CourseLecturer), sendStatusMailManualTrigger)
	mailing.POST("/:coursePhaseID/manual", permissionRoleMiddleware(permissionValidation.PromptAdmin, permissionValidation.PromptLecturer, permissionValidation.CourseLecturer), sendManualMailTrigger)
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

// sendManualMailTrigger godoc
// @Summary Manually trigger custom mails for a selected recipient list
// @Description Sends mails to the provided course participations using the provided template and placeholders
// @Tags mailing
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Param request body mailingDTO.SendManualMailRequest true "Manual mail request"
// @Success 200 {object} mailingDTO.ManualMailReport
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /mailing/{coursePhaseID}/manual [post]
func sendManualMailTrigger(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	var request mailingDTO.SendManualMailRequest
	if err := c.BindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	report, err := sendManualMailFn(
		c,
		coursePhaseID,
		request,
	)
	if err != nil {
		if errors.Is(err, ErrManualMailValidation) {
			handleError(c, http.StatusBadRequest, err)
			return
		}
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, report)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, utils.ErrorResponse{
		Error: err.Error(),
	})
}
