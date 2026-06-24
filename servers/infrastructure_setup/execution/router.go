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

// listInstances godoc
// @Summary List resource instances
// @Description Lists provisioned resource instances and their lifecycle status for a course phase.
// @Tags execution
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} ResourceInstanceResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/instances [get]
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

// triggerExecution godoc
// @Summary Trigger infrastructure provisioning
// @Description Creates pending resource instances for configured targets and starts asynchronous provisioning.
// @Tags execution
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 202 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/execute [post]
func triggerExecution(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}
		if err := svc.TriggerExecution(c.Request.Context(), c.GetHeader("Authorization"), coursePhaseID); err != nil {
			log.WithError(err).Error("trigger execution")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"message": "execution started"})
	}
}

// retryInstance godoc
// @Summary Retry failed resource instance
// @Description Resets a failed resource instance to pending and starts asynchronous provisioning.
// @Tags execution
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param instanceID path string true "Resource instance ID"
// @Success 202 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/instances/{instanceID}/retry [post]
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
		if err := svc.RetryFailedInstanceWithAuth(c.Request.Context(), c.GetHeader("Authorization"), coursePhaseID, instanceID); err != nil {
			log.WithError(err).Error("retry instance")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"message": "retry started"})
	}
}

// deleteInstance godoc
// @Summary Delete resource instance
// @Description Deletes a resource instance row in PROMPT. External provider resources are not deleted.
// @Tags execution
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param instanceID path string true "Resource instance ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/instances/{instanceID} [delete]
func deleteInstance(svc *Service) gin.HandlerFunc {
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
		if err := svc.DeleteInstance(c.Request.Context(), coursePhaseID, instanceID); err != nil {
			log.WithError(err).Error("delete instance")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNoContent, nil)
	}
}
