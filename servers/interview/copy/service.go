package copy

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

// auditCopyAction labels the copy route on both the denied and the completed
// path, so filtering the audit log by action finds every attempt.
const auditCopyAction = "Copied course phase"

// interviewCopyHandler implements promptTypes.PhaseCopyHandler. Interview slots and
// reviews are tied to the phase they were created for, so nothing is carried over.
type interviewCopyHandler struct{}

func (h *interviewCopyHandler) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
	recordCopyAudit(c, req)

	return nil
}

// recordCopyAudit scopes the event to the target phase. The route sits outside
// :coursePhaseID, so an automatically captured event would carry no phase and
// never reach the course audit log. A blank target is left to that automatic
// entry rather than pinning the log to the nil phase.
func recordCopyAudit(c *gin.Context, req promptTypes.PhaseCopyRequest) {
	// Core probes this endpoint with source == target to find out whether it
	// exists, so such a request is not a copy and belongs in no audit log.
	if req.SourceCoursePhaseID == req.TargetCoursePhaseID {
		audit.Suppress(c)
		return
	}
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
