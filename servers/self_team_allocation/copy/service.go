package copy

import (
	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt-sdk/audit"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
)

type SelfTeamCopyHandler struct{}

func (h *SelfTeamCopyHandler) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
	// Teams, assignments, tutors and the timeframe are all scoped to the phase
	// they were created for, so nothing is carried over and an audit entry would
	// claim a copy that did not happen.
	audit.Suppress(c)
	return nil
}
