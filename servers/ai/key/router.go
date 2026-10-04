package key

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	log "github.com/sirupsen/logrus"
)

const minKeyLength = 8

type setKeyRequest struct {
	Key string `json:"key" binding:"required,max=512"`
}

// PromptLecturer is never allowed: the SDK admits it for every course phase.
func RegisterRoutes(coursePhaseAPI *gin.RouterGroup, service *Service) {
	keyAPI := coursePhaseAPI.Group("/key", promptSDK.AuthenticationMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer))
	keyAPI.GET("", service.getKey)
	keyAPI.PUT("", service.setKey)
	keyAPI.DELETE("", service.deleteKey)
}

func (s *Service) getKey(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, sdkUtils.ErrorResponse{Error: "invalid course phase id"})
		return
	}
	status, err := s.Get(c.Request.Context(), coursePhaseID)
	if err != nil {
		log.WithError(err).Error("Could not load the phase key")
		c.JSON(http.StatusInternalServerError, sdkUtils.ErrorResponse{Error: "could not load the key"})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (s *Service) setKey(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, sdkUtils.ErrorResponse{Error: "invalid course phase id"})
		return
	}
	var request setKeyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, sdkUtils.ErrorResponse{Error: "a key of at most 512 characters is required"})
		return
	}
	logosKey := strings.TrimSpace(request.Key)
	if len(logosKey) < minKeyLength || strings.ContainsFunc(logosKey, func(r rune) bool { return r < '!' || r > '~' }) {
		c.JSON(http.StatusBadRequest, sdkUtils.ErrorResponse{Error: "the key must be at least 8 printable ASCII characters"})
		return
	}
	user, _ := keycloakTokenVerifier.GetTokenUser(c)
	status, err := s.Set(c.Request.Context(), coursePhaseID, logosKey, user.ID, c.GetHeader("Authorization"))
	if err != nil {
		log.WithError(err).Error("Could not store the phase key")
		c.JSON(http.StatusInternalServerError, sdkUtils.ErrorResponse{Error: "could not store the key"})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (s *Service) deleteKey(c *gin.Context) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, sdkUtils.ErrorResponse{Error: "invalid course phase id"})
		return
	}
	if err := s.Delete(c.Request.Context(), coursePhaseID); err != nil {
		log.WithError(err).Error("Could not delete the phase key")
		c.JSON(http.StatusInternalServerError, sdkUtils.ErrorResponse{Error: "could not delete the key"})
		return
	}
	c.Status(http.StatusNoContent)
}
