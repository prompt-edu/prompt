package config

import (
	"github.com/gin-gonic/gin"
)

// configHandler implements promptTypes.PhaseConfigHandler. The interview phase has
// nothing to configure, so the handler is stateless and reports an empty config.
type configHandler struct{}

func (h *configHandler) HandlePhaseConfig(c *gin.Context) (map[string]bool, error) {
	return map[string]bool{}, nil
}
