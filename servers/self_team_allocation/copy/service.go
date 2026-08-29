package copy

import (
	"github.com/gin-gonic/gin"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
)

type SelfTeamCopyHandler struct{}

func (h *SelfTeamCopyHandler) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
	return nil
}
