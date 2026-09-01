package copy

import (
	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

// interviewCopyHandler implements promptTypes.PhaseCopyHandler. Interview slots and
// reviews are tied to the phase they were created for, so nothing is carried over.
type interviewCopyHandler struct{}

func (h *interviewCopyHandler) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
	return nil
}
