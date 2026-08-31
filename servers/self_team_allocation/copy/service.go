package copy

import (
	"github.com/gin-gonic/gin"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
)

type selfTeamCopyHandler struct{}

func (h *selfTeamCopyHandler) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
	return nil
}
