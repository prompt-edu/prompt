package execution

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	log "github.com/sirupsen/logrus"
)

// RegisterRoutes mounts execution endpoints on the given router group.
func RegisterRoutes(rg *gin.RouterGroup, svc *Service) {
	rg.GET("/instances", listInstances(svc))
	rg.POST("/execute", triggerExecution(svc))
	rg.GET("/execute/preview", previewExecution(svc))
	rg.POST("/instances/:instanceID/retry", retryInstance(svc))
	rg.DELETE("/instances/:instanceID", deleteInstance(svc))
}

// RegisterStudentRoutes mounts the endpoints a student of the phase calls. The group
// must only admit students: the handlers answer for the caller's own participation.
func RegisterStudentRoutes(rg *gin.RouterGroup, svc *Service) {
	rg.GET("/my-resources", listMyResources(svc))
}

// isPreconditionError reports whether a trigger was refused because of how the phase is
// set up, which the lecturer can fix, rather than because something broke.
func isPreconditionError(err error) bool {
	return errors.Is(err, ErrProviderNotConfigured) ||
		errors.Is(err, ErrNothingConfigured) ||
		errors.Is(err, ErrSemesterTagMissing) ||
		errors.Is(err, ErrTeamsNotWired)
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
// @Success 202 {object} TriggerSummary
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
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
		summary, err := svc.TriggerExecution(c.Request.Context(), c.GetHeader("Authorization"), coursePhaseID)
		if err != nil {
			switch {
			case errors.Is(err, ErrExecutionInProgress):
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			case isPreconditionError(err):
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			default:
				log.WithError(err).Error("trigger execution")
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusAccepted, summary)
	}
}

// previewExecution godoc
// @Summary Preview infrastructure provisioning
// @Description Runs the checks of a trigger and resolves its targets without writing anything, and reports how many instances a trigger would create, retry and leave alone.
// @Tags execution
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {object} ProvisioningPreview
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/execute/preview [get]
func previewExecution(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}
		preview, err := svc.PreviewExecution(c.Request.Context(), c.GetHeader("Authorization"), coursePhaseID)
		if err != nil {
			if isPreconditionError(err) {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			log.WithError(err).Error("preview execution")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, preview)
	}
}

// listMyResources godoc
// @Summary List the caller's resources
// @Description Lists every resource config of the phase as the calling student sees it: the instances provisioned for them or their team, with whether they were granted access, and configs that have provisioned nothing for them yet. Error details are never included.
// @Tags execution
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {array} MyResourceResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/my-resources [get]
func listMyResources(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}
		courseParticipationID, err := keycloakTokenVerifier.GetUserCourseParticipationID(c)
		if err != nil {
			c.JSON(keycloakTokenVerifier.GetUserCourseParticipationIDErrorStatus(err), gin.H{"error": err.Error()})
			return
		}
		resources, err := svc.ListMyResources(c.Request.Context(), coursePhaseID, courseParticipationID)
		if err != nil {
			log.WithError(err).Error("list my resources")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load your resources"})
			return
		}
		c.JSON(http.StatusOK, resources)
	}
}

// retryInstance godoc
// @Summary Retry failed resource instance
// @Description Resets a failed or partial resource instance to pending and starts asynchronous provisioning.
// @Tags execution
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param instanceID path string true "Resource instance ID"
// @Success 202 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
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
		if err := svc.RetryInstance(c.Request.Context(), c.GetHeader("Authorization"), coursePhaseID, instanceID); err != nil {
			switch {
			case errors.Is(err, ErrInstanceNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			case errors.Is(err, ErrInstanceNotRetryable):
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			default:
				log.WithError(err).Error("retry instance")
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
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
