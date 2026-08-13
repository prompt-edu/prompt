package courseMailing

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/courseMailing/courseMailingDTO"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/utils"
	log "github.com/sirupsen/logrus"
)

// setupCourseMailingRouter registers the course-level mail campaign endpoints.
// @Summary Course Mailing Endpoints
// @Description Endpoints for managing per-course mail campaigns
// @Tags course_mailing
// @Security BearerAuth
func setupCourseMailingRouter(router *gin.RouterGroup, authMiddleware func() gin.HandlerFunc, permissionIDMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	editRoles := []string{permissionValidation.PromptAdmin, permissionValidation.PromptLecturer, permissionValidation.CourseLecturer, permissionValidation.CourseEditor}
	sendRoles := []string{permissionValidation.PromptAdmin, permissionValidation.PromptLecturer, permissionValidation.CourseLecturer}

	campaigns := router.Group("/courses/:uuid/mail-campaigns", authMiddleware())
	campaigns.GET("", permissionIDMiddleware(editRoles...), listCampaigns)
	campaigns.POST("", permissionIDMiddleware(editRoles...), createCampaign)
	campaigns.GET("/:campaignID", permissionIDMiddleware(editRoles...), getCampaign)
	campaigns.PUT("/:campaignID", permissionIDMiddleware(editRoles...), updateCampaign)
	campaigns.DELETE("/:campaignID", permissionIDMiddleware(editRoles...), deleteCampaign)
	campaigns.POST("/:campaignID/copy", permissionIDMiddleware(editRoles...), copyCampaign)
	campaigns.GET("/:campaignID/recipients-preview", permissionIDMiddleware(editRoles...), previewRecipients)
	campaigns.POST("/:campaignID/test", permissionIDMiddleware(editRoles...), testSendCampaign)
	campaigns.POST("/:campaignID/send", permissionIDMiddleware(sendRoles...), sendCampaign)
	campaigns.POST("/:campaignID/resend-failed", permissionIDMiddleware(sendRoles...), resendFailedCampaign)
}

func actorFromContext(c *gin.Context) courseMailingDTO.Actor {
	name := strings.TrimSpace(c.GetString(keycloakTokenVerifier.CtxFirstName) + " " + c.GetString(keycloakTokenVerifier.CtxLastName))
	return courseMailingDTO.Actor{
		ID:    c.GetString(keycloakTokenVerifier.CtxUserID),
		Email: c.GetString(keycloakTokenVerifier.CtxUserEmail),
		Name:  name,
	}
}

func parseCourseAndCampaign(c *gin.Context) (uuid.UUID, uuid.UUID, error) {
	courseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}
	campaignID, err := uuid.Parse(c.Param("campaignID"))
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}
	return courseID, campaignID, nil
}

// listCampaigns godoc
// @Summary List mail campaigns for a course
// @Tags course_mailing
// @Produce json
// @Param uuid path string true "Course UUID"
// @Success 200 {array} courseMailingDTO.MailCampaign
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/mail-campaigns [get]
func listCampaigns(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	campaigns, err := CourseMailingServiceSingleton.ListCampaigns(c, courseID)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, campaigns)
}

// createCampaign godoc
// @Summary Create a mail campaign draft
// @Tags course_mailing
// @Accept json
// @Produce json
// @Param uuid path string true "Course UUID"
// @Param campaign body courseMailingDTO.MailCampaignRequest true "Campaign"
// @Success 201 {object} courseMailingDTO.MailCampaign
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/mail-campaigns [post]
func createCampaign(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	req, err := bindCampaignRequest(c)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	created, err := CourseMailingServiceSingleton.CreateCampaign(c, courseID, actorFromContext(c), req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, courseMailingDTO.MailCampaignFromModel(created))
}

// getCampaign godoc
// @Summary Get a mail campaign with recipients
// @Tags course_mailing
// @Produce json
// @Param uuid path string true "Course UUID"
// @Param campaignID path string true "Campaign UUID"
// @Success 200 {object} courseMailingDTO.MailCampaignDetail
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/mail-campaigns/{campaignID} [get]
func getCampaign(c *gin.Context) {
	courseID, campaignID, err := parseCourseAndCampaign(c)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	detail, err := CourseMailingServiceSingleton.GetCampaignDetail(c, courseID, campaignID)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

// updateCampaign godoc
// @Summary Update a mail campaign
// @Tags course_mailing
// @Accept json
// @Produce json
// @Param uuid path string true "Course UUID"
// @Param campaignID path string true "Campaign UUID"
// @Param campaign body courseMailingDTO.MailCampaignRequest true "Campaign"
// @Success 200 {object} courseMailingDTO.MailCampaign
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/mail-campaigns/{campaignID} [put]
func updateCampaign(c *gin.Context) {
	courseID, campaignID, err := parseCourseAndCampaign(c)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	req, err := bindCampaignRequest(c)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	updated, err := CourseMailingServiceSingleton.UpdateCampaign(c, courseID, campaignID, actorFromContext(c), req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, courseMailingDTO.MailCampaignFromModel(updated))
}

// deleteCampaign godoc
// @Summary Delete a mail campaign
// @Tags course_mailing
// @Param uuid path string true "Course UUID"
// @Param campaignID path string true "Campaign UUID"
// @Success 200 {string} string "OK"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/mail-campaigns/{campaignID} [delete]
func deleteCampaign(c *gin.Context) {
	courseID, campaignID, err := parseCourseAndCampaign(c)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	if err := CourseMailingServiceSingleton.DeleteCampaign(c, courseID, campaignID); err != nil {
		handleServiceError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

// copyCampaign godoc
// @Summary Copy a mail campaign into a new draft
// @Tags course_mailing
// @Produce json
// @Param uuid path string true "Course UUID"
// @Param campaignID path string true "Campaign UUID"
// @Success 201 {object} courseMailingDTO.MailCampaign
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/mail-campaigns/{campaignID}/copy [post]
func copyCampaign(c *gin.Context) {
	courseID, campaignID, err := parseCourseAndCampaign(c)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	copied, err := CourseMailingServiceSingleton.CopyCampaign(c, courseID, campaignID, actorFromContext(c))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, courseMailingDTO.MailCampaignFromModel(copied))
}

// previewRecipients godoc
// @Summary Preview the live recipient list for a campaign
// @Tags course_mailing
// @Produce json
// @Param uuid path string true "Course UUID"
// @Param campaignID path string true "Campaign UUID"
// @Success 200 {object} courseMailingDTO.RecipientPreview
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/mail-campaigns/{campaignID}/recipients-preview [get]
func previewRecipients(c *gin.Context) {
	courseID, campaignID, err := parseCourseAndCampaign(c)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	preview, err := CourseMailingServiceSingleton.PreviewRecipients(c, courseID, campaignID)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, preview)
}

// testSendCampaign godoc
// @Summary Send a test copy of a campaign to the current user
// @Tags course_mailing
// @Param uuid path string true "Course UUID"
// @Param campaignID path string true "Campaign UUID"
// @Success 200 {string} string "OK"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/mail-campaigns/{campaignID}/test [post]
func testSendCampaign(c *gin.Context) {
	courseID, campaignID, err := parseCourseAndCampaign(c)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	if err := CourseMailingServiceSingleton.TestSend(c, courseID, campaignID, actorFromContext(c)); err != nil {
		handleServiceError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

// sendCampaign godoc
// @Summary Send a mail campaign to all matching recipients
// @Tags course_mailing
// @Produce json
// @Param uuid path string true "Course UUID"
// @Param campaignID path string true "Campaign UUID"
// @Success 202 {object} courseMailingDTO.SendResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/mail-campaigns/{campaignID}/send [post]
func sendCampaign(c *gin.Context) {
	courseID, campaignID, err := parseCourseAndCampaign(c)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	count, err := CourseMailingServiceSingleton.SendCampaign(c, courseID, campaignID, actorFromContext(c))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, courseMailingDTO.SendResponse{RecipientCount: count})
}

// resendFailedCampaign godoc
// @Summary Resend a campaign to its failed recipients only
// @Tags course_mailing
// @Produce json
// @Param uuid path string true "Course UUID"
// @Param campaignID path string true "Campaign UUID"
// @Success 202 {object} courseMailingDTO.SendResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/mail-campaigns/{campaignID}/resend-failed [post]
func resendFailedCampaign(c *gin.Context) {
	courseID, campaignID, err := parseCourseAndCampaign(c)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	count, err := CourseMailingServiceSingleton.ResendFailed(c, courseID, campaignID, actorFromContext(c))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, courseMailingDTO.SendResponse{RecipientCount: count})
}

func bindCampaignRequest(c *gin.Context) (courseMailingDTO.MailCampaignRequest, error) {
	var req courseMailingDTO.MailCampaignRequest
	// ShouldBindJSON does not write a response on error, so the handler owns the
	// error mapping (BindJSON would abort with its own 400 and double-write).
	if err := c.ShouldBindJSON(&req); err != nil {
		return courseMailingDTO.MailCampaignRequest{}, err
	}
	return req, nil
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, utils.ErrorResponse{Error: err.Error()})
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		handleError(c, http.StatusNotFound, err)
	case errors.Is(err, ErrSendInProgress):
		handleError(c, http.StatusConflict, err)
	case errors.Is(err, ErrNoRecipients):
		handleError(c, http.StatusUnprocessableEntity, err)
	case errors.Is(err, ErrValidation):
		handleError(c, http.StatusBadRequest, err)
	default:
		// Do not leak wrapped database/internal error text to the client.
		log.Error(err)
		handleError(c, http.StatusInternalServerError, errors.New("internal server error"))
	}
}
