package phaseconfig

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/phaseconfig/phaseconfigDTO"
	log "github.com/sirupsen/logrus"
)

// RegisterRoutes mounts infrastructure setup phase config endpoints.
func RegisterRoutes(rg *gin.RouterGroup, svc *Service) {
	rg.GET("/setup-config", getConfig(svc))
	rg.PUT("/setup-config", upsertConfig(svc))
}

// getConfig godoc
// @Summary Get infrastructure setup configuration
// @Description Returns the semester tag configuration for an infrastructure setup course phase. Upstream team and participation data is wired via the standard phase configurator.
// @Tags infrastructure-setup-config
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Success 200 {object} phaseconfigDTO.Response
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/setup-config [get]
func getConfig(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}
		resp, err := svc.Get(c.Request.Context(), coursePhaseID)
		if err != nil {
			log.WithError(err).Error("get infrastructure setup config")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// upsertConfig godoc
// @Summary Create or update infrastructure setup configuration
// @Description Saves the semester tag configuration for an infrastructure setup course phase.
// @Tags infrastructure-setup-config
// @Accept json
// @Produce json
// @Param coursePhaseID path string true "Course phase ID"
// @Param config body phaseconfigDTO.UpsertRequest true "Infrastructure setup configuration"
// @Success 200 {object} phaseconfigDTO.Response
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /course_phase/{coursePhaseID}/setup-config [put]
func upsertConfig(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coursePhaseID"})
			return
		}

		var req phaseconfigDTO.UpsertRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := svc.Upsert(c.Request.Context(), coursePhaseID, req)
		if err != nil {
			log.WithError(err).Error("upsert infrastructure setup config")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}
