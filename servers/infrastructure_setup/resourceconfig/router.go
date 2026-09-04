package resourceconfig

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/resourceconfig/resourceconfigDTO"
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

// listResourceConfigs godoc
// @Summary List resource configurations
// @Description Lists all resource configurations for an infrastructure setup course phase.
// @Tags resource-configs
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} resourceconfigDTO.ResourceConfigResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/resource-configs [get]
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

// createResourceConfig godoc
// @Summary Create resource configuration
// @Description Creates a resource configuration describing what to provision per team or per student.
// @Tags resource-configs
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param resourceConfig body resourceconfigDTO.CreateRequest true "Resource configuration"
// @Success 201 {object} resourceconfigDTO.ResourceConfigResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/resource-configs [post]
func createResourceConfig(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}
		var req resourceconfigDTO.CreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		resp, err := svc.CreateResourceConfig(c.Request.Context(), coursePhaseID, req)
		if err != nil {
			respondWithError(c, "create resource config", err)
			return
		}
		c.JSON(http.StatusCreated, resp)
	}
}

// getResourceConfig godoc
// @Summary Get resource configuration
// @Description Retrieves a single resource configuration by ID.
// @Tags resource-configs
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param resourceConfigID path string true "Resource configuration ID"
// @Success 200 {object} resourceconfigDTO.ResourceConfigResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/resource-configs/{resourceConfigID} [get]
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
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "resource config not found"})
				return
			}
			log.WithError(err).Error("failed to get resource config")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get resource config"})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// updateResourceConfig godoc
// @Summary Update resource configuration
// @Description Updates a resource configuration.
// @Tags resource-configs
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param resourceConfigID path string true "Resource configuration ID"
// @Param resourceConfig body resourceconfigDTO.UpdateRequest true "Resource configuration"
// @Success 200 {object} resourceconfigDTO.ResourceConfigResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/resource-configs/{resourceConfigID} [put]
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
		var req resourceconfigDTO.UpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		resp, err := svc.UpdateResourceConfig(c.Request.Context(), coursePhaseID, id, req)
		if err != nil {
			respondWithError(c, "update resource config", err)
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// deleteResourceConfig godoc
// @Summary Delete resource configuration
// @Description Deletes a resource configuration and cascades to its resource instances. A configuration with provisioned instances requires confirm=true, because the instances are PROMPT's only record of the external resources, which are never deleted.
// @Tags resource-configs
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param resourceConfigID path string true "Resource configuration ID"
// @Param confirm query bool false "Confirm dropping the records of provisioned instances"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/resource-configs/{resourceConfigID} [delete]
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
		confirmed, err := strconv.ParseBool(c.DefaultQuery("confirm", "false"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "confirm must be a boolean"})
			return
		}
		if err := svc.DeleteResourceConfig(c.Request.Context(), coursePhaseID, id, confirmed); err != nil {
			respondWithError(c, "delete resource config", err)
			return
		}
		c.JSON(http.StatusNoContent, nil)
	}
}

// respondWithError answers 400 for a rejected request, 409 for one that needs
// confirmation, and 500 for anything else.
func respondWithError(c *gin.Context, action string, err error) {
	if errors.Is(err, ErrValidation) || errors.Is(err, ErrProviderNotConfigured) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, ErrConfirmationRequired) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	log.WithError(err).Error(action)
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
