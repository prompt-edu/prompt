package copy

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	log "github.com/sirupsen/logrus"
)

// auditCopyAction names both the copy route and the event its handler records.
const auditCopyAction = "Copied course phase"

type CopyService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

func NewCopyService(queries db.Queries, conn *pgxpool.Pool) *CopyService {
	return &CopyService{
		queries: queries,
		conn:    conn,
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

func (s *CopyService) CopyPhase(ctx context.Context, sourceCoursePhaseID, targetCoursePhaseID uuid.UUID) error {
	if sourceCoursePhaseID == targetCoursePhaseID {
		return nil
	}

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := s.queries.WithTx(tx)

	skills, err := qtx.GetSkillsByCoursePhase(ctx, sourceCoursePhaseID)
	if err != nil {
		return err
	}

	// Copy skills to the new course phase
	for _, skill := range skills {
		err := qtx.CreateSkill(ctx, db.CreateSkillParams{
			ID:            uuid.New(),
			Name:          skill.Name,
			CoursePhaseID: targetCoursePhaseID,
		})
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("could not commit phase copy: ", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
