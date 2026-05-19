package providerconfig

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// RegisterRoutes mounts provider config endpoints on the given router group.
// All routes require the coursePhaseID path parameter.
func RegisterRoutes(rg *gin.RouterGroup, svc *Service) {
	rg.GET("/provider-configs", listProviderConfigs(svc))
	rg.PUT("/provider-configs", upsertProviderConfig(svc))
	rg.DELETE("/provider-configs/:providerType", deleteProviderConfig(svc))
	rg.POST("/provider-configs/:providerType/validate", validateProviderConfig(svc))
	rg.GET("/provider-configs/:providerType/fields", getAuthFields())
}

// deleteProviderConfig godoc
// @Summary Delete provider configuration
// @Description Removes stored credentials for a provider type. Resource configurations and instances for that provider cascade-delete.
// @Tags provider-configs
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param providerType path string true "Provider type" Enums(gitlab, slack, outline, rancher, keycloak)
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/provider-configs/{providerType} [delete]
func deleteProviderConfig(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}

		providerType := c.Param("providerType")
		if _, err := GetAuthFields(providerType); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := svc.DeleteProviderConfig(c.Request.Context(), coursePhaseID, providerType); err != nil {
			log.WithError(err).Error("delete provider config")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNoContent, nil)
	}
}

// listProviderConfigs godoc
// @Summary List provider configurations
// @Description Lists configured infrastructure providers for a course phase. Credentials are redacted.
// @Tags provider-configs
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} ProviderConfigResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/provider-configs [get]
func listProviderConfigs(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}

		configs, err := svc.ListProviderConfigs(c.Request.Context(), coursePhaseID)
		if err != nil {
			log.WithError(err).Error("list provider configs")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, configs)
	}
}

// upsertProviderConfig godoc
// @Summary Create or update provider configuration
// @Description Encrypts and stores credentials for an infrastructure provider.
// @Tags provider-configs
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param providerConfig body UpsertRequest true "Provider configuration"
// @Success 200 {object} ProviderConfigResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/provider-configs [put]
func upsertProviderConfig(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}

		var req UpsertRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := svc.UpsertProviderConfig(c.Request.Context(), coursePhaseID, req)
		if err != nil {
			log.WithError(err).Error("upsert provider config")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// validateProviderConfig godoc
// @Summary Validate provider credentials
// @Description Decrypts the stored provider credentials and validates them against the external provider.
// @Tags provider-configs
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param providerType path string true "Provider type" Enums(gitlab, slack, outline, rancher, keycloak)
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/provider-configs/{providerType}/validate [post]
func validateProviderConfig(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}

		providerType := c.Param("providerType")
		if err := svc.ValidateProviderConfig(c.Request.Context(), coursePhaseID, providerType); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"valid": true})
	}
}

// getAuthFields godoc
// @Summary Get provider credential fields
// @Description Returns the credential fields required to configure a provider type.
// @Tags provider-configs
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param providerType path string true "Provider type" Enums(gitlab, slack, outline, rancher, keycloak)
// @Success 200 {array} AuthField
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/provider-configs/{providerType}/fields [get]
func getAuthFields() gin.HandlerFunc {
	return func(c *gin.Context) {
		providerType := c.Param("providerType")
		fields, err := GetAuthFields(providerType)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, fields)
	}
}
