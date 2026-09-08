package copy

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt-sdk/audit"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
)

// auditCopyAction names the copy route and the event its handler records, so
// both describe the same action in the audit log.
const auditCopyAction = "Copied course phase"

type selfTeamCopyHandler struct{}

func (h *selfTeamCopyHandler) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
	// Core probes this endpoint with source == target to find out whether it
	// exists, so such a request is not a copy and belongs in no audit log.
	if req.SourceCoursePhaseID == req.TargetCoursePhaseID {
		audit.Suppress(c)
		return nil
	}

	recordCopyAudit(c, req)

	return nil
}

// recordCopyAudit scopes the event to the target phase. The route sits outside
// :coursePhaseID, so an automatically captured event would carry no phase and
// never reach the course audit log. A blank target is left to that automatic
// entry rather than pinning the log to the nil phase.
func recordCopyAudit(c *gin.Context, req promptTypes.PhaseCopyRequest) {
	if req.TargetCoursePhaseID == uuid.Nil {
		return
	}
	audit.Record(c, audit.Event{
		Action:        auditCopyAction,
		EntityType:    "coursePhase",
		EntityID:      req.TargetCoursePhaseID.String(),
		CoursePhaseID: req.TargetCoursePhaseID.String(),
		Metadata:      map[string]any{"sourceCoursePhaseID": req.SourceCoursePhaseID.String()},
	})
}
