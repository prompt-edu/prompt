package ai

import (
	"net/http"

	"github.com/gin-gonic/gin"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution"
)

const ServiceName = "AI"

type Status struct {
	Enabled bool `json:"enabled"`
}

func Enabled() bool {
	return sdkUtils.GetEnv("AI_ENABLED", "") == "true"
}

func ServerURL() string {
	if !Enabled() {
		return ""
	}
	if sdkUtils.GetEnv("ENVIRONMENT", "development") == "development" {
		return "http://localhost:8092/ai/api"
	}
	return resolution.NormaliseHost(sdkUtils.GetEnv("CORE_HOST", "localhost:8080")) + "/ai/api"
}

func RegisterRoutes(api *gin.RouterGroup, authMiddleware func() gin.HandlerFunc) {
	api.GET("/ai/status", authMiddleware(), getStatus)
}

// getStatus godoc
// @Summary AI status
// @Description Reports whether the AI features are enabled for this deployment.
// @Tags ai
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ai.Status
// @Router /ai/status [get]
func getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, Status{Enabled: Enabled()})
}
