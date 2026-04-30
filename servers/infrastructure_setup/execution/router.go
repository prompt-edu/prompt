package execution

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// RegisterRoutes mounts execution endpoints on the given router group.
func RegisterRoutes(rg *gin.RouterGroup, svc *Service) {
	rg.GET("/instances", listInstances(svc))
	rg.POST("/execute", triggerExecution(svc))
	rg.POST("/instances/:instanceID/retry", retryInstance(svc))
	rg.DELETE("/instances/:instanceID", deleteInstance(svc))
}

func listInstances(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}
		instances, err := svc.ListInstances(c.Request.Context(), coursePhaseID)
		if err != nil {
			log.WithError(err).Error("list instances")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, instances)
	}
}

func triggerExecution(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}
		if err := svc.TriggerExecution(c.Request.Context(), coursePhaseID); err != nil {
			log.WithError(err).Error("trigger execution")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"message": "execution started"})
	}
}

func retryInstance(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}
		instanceID, err := uuid.Parse(c.Param("instanceID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid instanceID"})
			return
		}
		if err := svc.RetryFailedInstance(c.Request.Context(), coursePhaseID, instanceID); err != nil {
			log.WithError(err).Error("retry instance")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"message": "retry started"})
	}
}

func deleteInstance(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		instanceID, err := uuid.Parse(c.Param("instanceID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid instanceID"})
			return
		}
		if err := svc.DeleteInstance(c.Request.Context(), instanceID); err != nil {
			log.WithError(err).Error("delete instance")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNoContent, nil)
	}
}
