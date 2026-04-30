package resourceconfig

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// RegisterRoutes mounts resource config endpoints on the given router group.
func RegisterRoutes(rg *gin.RouterGroup, svc *Service) {
	rg.GET("/resource-configs", listResourceConfigs(svc))
	rg.POST("/resource-configs", createResourceConfig(svc))
	rg.GET("/resource-configs/:resourceConfigID", getResourceConfig(svc))
	rg.PUT("/resource-configs/:resourceConfigID", updateResourceConfig(svc))
	rg.DELETE("/resource-configs/:resourceConfigID", deleteResourceConfig(svc))
}

func listResourceConfigs(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}
		configs, err := svc.ListResourceConfigs(c.Request.Context(), coursePhaseID)
		if err != nil {
			log.WithError(err).Error("list resource configs")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, configs)
	}
}

func createResourceConfig(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}
		var req CreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		resp, err := svc.CreateResourceConfig(c.Request.Context(), coursePhaseID, req)
		if err != nil {
			log.WithError(err).Error("create resource config")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, resp)
	}
}

func getResourceConfig(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}
		id, err := uuid.Parse(c.Param("resourceConfigID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resourceConfigID"})
			return
		}
		resp, err := svc.GetResourceConfig(c.Request.Context(), coursePhaseID, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "resource config not found"})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

func updateResourceConfig(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}
		id, err := uuid.Parse(c.Param("resourceConfigID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resourceConfigID"})
			return
		}
		var req UpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		resp, err := svc.UpdateResourceConfig(c.Request.Context(), coursePhaseID, id, req)
		if err != nil {
			log.WithError(err).Error("update resource config")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

func deleteResourceConfig(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}
		id, err := uuid.Parse(c.Param("resourceConfigID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resourceConfigID"})
			return
		}
		if err := svc.DeleteResourceConfig(c.Request.Context(), coursePhaseID, id); err != nil {
			log.WithError(err).Error("delete resource config")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNoContent, nil)
	}
}
