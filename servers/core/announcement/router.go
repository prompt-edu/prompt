package announcement

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt-sdk/audit"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/core/announcement/announcementDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

func RegisterRoutes(routerGroup *gin.RouterGroup, service *AnnouncementService, authMiddleware func() gin.HandlerFunc) {
	setupAnnouncementRouter(routerGroup, service, authMiddleware, permissionValidation.CheckAccessControlByRole)
}

func setupAnnouncementRouter(router *gin.RouterGroup, s *AnnouncementService, authMiddleware func() gin.HandlerFunc, permissionRoleMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	router.GET("/announcements/active", s.getActiveAnnouncements)

	admin := router.Group("/announcements", authMiddleware(), permissionRoleMiddleware(permissionValidation.PromptAdmin))
	admin.GET("", s.getAnnouncements)
	admin.POST("", audit.Describe("Created announcement"), s.createAnnouncement)
	admin.PUT("/:id", audit.Describe("Updated announcement"), s.updateAnnouncement)
	admin.DELETE("/:id", audit.Describe("Deleted announcement"), s.deleteAnnouncement)
}

// getActiveAnnouncements godoc
// @Summary Get active announcements
// @Description Get all enabled announcements whose schedule includes the current time. Public endpoint.
// @Tags announcements
// @Produce json
// @Success 200 {object} []announcementDTO.Announcement
// @Failure 500 {object} utils.ErrorResponse
// @Router /announcements/active [get]
func (s *AnnouncementService) getActiveAnnouncements(c *gin.Context) {
	announcements, err := s.ListActiveAnnouncements(c)
	if err != nil {
		utils.RespondWithDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, announcements)
}

// getAnnouncements godoc
// @Summary Get all announcements
// @Description Get all announcements for administration. Expired announcements are only included when requested.
// @Tags announcements
// @Produce json
// @Param includeExpired query bool false "Include expired announcements"
// @Success 200 {object} []announcementDTO.Announcement
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /announcements [get]
func (s *AnnouncementService) getAnnouncements(c *gin.Context) {
	includeExpired, err := strconv.ParseBool(c.DefaultQuery("includeExpired", "false"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	announcements, err := s.ListAnnouncements(c, includeExpired)
	if err != nil {
		utils.RespondWithDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, announcements)
}

// createAnnouncement godoc
// @Summary Create an announcement
// @Description Create a new announcement banner
// @Tags announcements
// @Accept json
// @Produce json
// @Param announcement body announcementDTO.UpsertAnnouncement true "Announcement to create"
// @Success 201 {object} announcementDTO.Announcement
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /announcements [post]
func (s *AnnouncementService) createAnnouncement(c *gin.Context) {
	var request announcementDTO.UpsertAnnouncement
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	request = normalizeAnnouncement(request)
	if err := validateAnnouncement(request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	announcement, err := s.CreateAnnouncement(c, request)
	if err != nil {
		utils.RespondWithDBError(c, err)
		return
	}
	c.JSON(http.StatusCreated, announcement)
}

// updateAnnouncement godoc
// @Summary Update an announcement
// @Description Update an existing announcement banner, including whether it is enabled
// @Tags announcements
// @Accept json
// @Produce json
// @Param id path string true "Announcement UUID"
// @Param announcement body announcementDTO.UpsertAnnouncement true "Updated announcement"
// @Success 200 {object} announcementDTO.Announcement
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /announcements/{id} [put]
func (s *AnnouncementService) updateAnnouncement(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	var request announcementDTO.UpsertAnnouncement
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}
	request = normalizeAnnouncement(request)
	if err := validateAnnouncement(request); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	announcement, err := s.UpdateAnnouncement(c, id, request)
	if err != nil {
		utils.RespondWithDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, announcement)
}

// deleteAnnouncement godoc
// @Summary Delete an announcement
// @Description Permanently delete an announcement banner
// @Tags announcements
// @Param id path string true "Announcement UUID"
// @Success 204
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /announcements/{id} [delete]
func (s *AnnouncementService) deleteAnnouncement(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if err := s.DeleteAnnouncement(c, id); err != nil {
		utils.RespondWithDBError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, sdkUtils.ErrorResponse{
		Error: err.Error(),
	})
}
