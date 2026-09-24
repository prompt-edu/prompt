package copy

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
)

// auditCopyAction names both the copy route and the event its handler records.
const auditCopyAction = "Copied course phase"

type CopyService struct {
	queries db.Queries
}

func NewCopyService(queries db.Queries) *CopyService {
	return &CopyService{
		queries: queries,
	}
}

// HandlePhaseCopy implements promptTypes.PhaseCopyHandler.
func (s *CopyService) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
	// Core probes whether this service supports copying by posting a copy of a phase onto
	// itself, so that probe must not reach the audit log.
	if req.SourceCoursePhaseID == req.TargetCoursePhaseID {
		audit.Suppress(c)
		return nil
	}
	recordCopyAudit(c, req)

	return s.CopyPhase(c.Request.Context(), req.SourceCoursePhaseID, req.TargetCoursePhaseID)
}

// recordCopyAudit scopes the event to the target phase: the route sits outside :coursePhaseID,
// so an automatically captured event would carry no phase and never reach the course audit log.
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

// CopyPhase copies the certificate template and the student page text into the target phase, see
// db/query/coursePhaseCopy.sql for what is deliberately left behind. The copy is a single
// statement, so it needs no transaction.
func (s *CopyService) CopyPhase(ctx context.Context, sourceCoursePhaseID, targetCoursePhaseID uuid.UUID) error {
	if sourceCoursePhaseID == targetCoursePhaseID {
		return nil
	}

	return s.queries.CopyCoursePhaseConfig(ctx, db.CopyCoursePhaseConfigParams{
		SourceCoursePhaseID: sourceCoursePhaseID,
		TargetCoursePhaseID: targetCoursePhaseID,
	})
}
