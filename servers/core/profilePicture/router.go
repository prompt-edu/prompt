package profilePicture

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt/servers/core/profilePicture/profilePictureDTO"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

// RegisterRoutes mounts the profile picture endpoints. Every logged-in user may read every
// picture; uploading and deleting only ever touch the caller's own picture.
func RegisterRoutes(routerGroup *gin.RouterGroup, service *ProfilePictureService, authMiddleware func() gin.HandlerFunc) {
	router := routerGroup.Group("/profile-pictures", authMiddleware())
	// Lookups and presigning change nothing, so they stay out of the audit log.
	router.POST("/lookup", audit.Skip(), service.lookupProfilePictures)
	router.GET("/me", service.getOwnProfilePicture)
	router.POST("/me/presign", audit.Skip(), service.presignProfilePictureUpload)
	router.POST("/me/complete", audit.Describe("Uploaded profile picture"), service.completeProfilePictureUpload)
	router.DELETE("/me", audit.Describe("Deleted profile picture"), service.deleteOwnProfilePicture)
}

// lookupProfilePictures godoc
// @Summary Look up profile pictures
// @Description Resolve presigned download URLs for the profile pictures of several users, students, or course participations. Ids without a picture are omitted.
// @Tags profilePictures
// @Accept json
// @Produce json
// @Param request body profilePictureDTO.LookupRequest true "Ids to look up"
// @Success 200 {object} profilePictureDTO.ProfilePictureURLs
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /profile-pictures/lookup [post]
func (s *ProfilePictureService) lookupProfilePictures(c *gin.Context) {
	var req profilePictureDTO.LookupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	urls, err := s.LookupPictureURLs(c, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, urls)
}

// getOwnProfilePicture godoc
// @Summary Get own profile picture
// @Description Get a presigned download URL for the profile picture of the logged-in user
// @Tags profilePictures
// @Produce json
// @Success 200 {object} profilePictureDTO.ProfilePicture
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /profile-pictures/me [get]
func (s *ProfilePictureService) getOwnProfilePicture(c *gin.Context) {
	userID, err := utils.GetUserUUIDFromContext(c)
	if err != nil {
		handleError(c, http.StatusUnauthorized, err)
		return
	}

	picture, err := s.GetOwnPicture(c, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, picture)
}

// presignProfilePictureUpload godoc
// @Summary Presign a profile picture upload
// @Description Get a presigned URL to upload a cropped JPEG profile picture for the logged-in user
// @Tags profilePictures
// @Produce json
// @Success 200 {object} profilePictureDTO.PresignedUpload
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /profile-pictures/me/presign [post]
func (s *ProfilePictureService) presignProfilePictureUpload(c *gin.Context) {
	userID, err := utils.GetUserUUIDFromContext(c)
	if err != nil {
		handleError(c, http.StatusUnauthorized, err)
		return
	}

	presigned, err := s.PresignUpload(c, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, presigned)
}

// completeProfilePictureUpload godoc
// @Summary Complete a profile picture upload
// @Description Validate an uploaded picture and make it the profile picture of the logged-in user, replacing the previous one
// @Tags profilePictures
// @Accept json
// @Produce json
// @Param request body profilePictureDTO.CompleteUpload true "Storage key returned by the presign endpoint"
// @Success 200 {object} profilePictureDTO.ProfilePicture
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /profile-pictures/me/complete [post]
func (s *ProfilePictureService) completeProfilePictureUpload(c *gin.Context) {
	userID, err := utils.GetUserUUIDFromContext(c)
	if err != nil {
		handleError(c, http.StatusUnauthorized, err)
		return
	}

	var req profilePictureDTO.CompleteUpload
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	picture, err := s.CompleteUpload(c, Uploader{
		UserID:          userID,
		UniversityLogin: utils.GetUniversityLoginFromContext(c),
		Email:           utils.GetUserEmailFromContext(c),
	}, req.StorageKey)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, picture)
}

// deleteOwnProfilePicture godoc
// @Summary Delete own profile picture
// @Description Delete the profile picture of the logged-in user
// @Tags profilePictures
// @Success 204
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /profile-pictures/me [delete]
func (s *ProfilePictureService) deleteOwnProfilePicture(c *gin.Context) {
	userID, err := utils.GetUserUUIDFromContext(c)
	if err != nil {
		handleError(c, http.StatusUnauthorized, err)
		return
	}

	if err := s.DeleteOwnPicture(c, userID); err != nil {
		handleServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		handleError(c, http.StatusBadRequest, err)
	case errors.Is(err, ErrNotFound):
		handleError(c, http.StatusNotFound, err)
	default:
		handleError(c, http.StatusInternalServerError, err)
	}
}

func handleError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, utils.ErrorResponse{
		Error: err.Error(),
	})
}
