package copy

import (
	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

// InterviewCopyHandler implements promptTypes.PhaseCopyHandler. Interview slots and
// reviews are tied to the phase they were created for, so nothing is carried over.
type InterviewCopyHandler struct{}

func (h *InterviewCopyHandler) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
	// Nothing is carried over, so an audit entry would claim a copy that did not happen.
	audit.Suppress(c)
	return nil
}
