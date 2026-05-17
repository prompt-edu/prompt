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
	rg.POST("/provider-configs/:providerType/validate", validateProviderConfig(svc))
	rg.GET("/provider-configs/:providerType/fields", getAuthFields())
}

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
