package copy

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	log "github.com/sirupsen/logrus"
)

// AuditCopyAction names the copy route and the event its handler records, so both
// describe the same action in the audit log.
const AuditCopyAction = "Copied assessment phase"

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
		Action:        AuditCopyAction,
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

	// Get the course phase config from the source course phase
	sourceConfig, err := qtx.GetCoursePhaseConfig(ctx, sourceCoursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to get source course phase config")
		return err
	}

	// Copying a disabled source must not hide grades a non-empty target already holds
	if !sourceConfig.AssessmentEnabled {
		hasData, err := qtx.PhaseHasAssessmentData(ctx, targetCoursePhaseID)
		if err != nil {
			log.WithError(err).Error("Failed to check target course phase for assessment data")
			return err
		}
		if hasData.Bool {
			return coursePhaseConfig.ErrCannotDisableAssessmentWithData
		}
	}

	// Create a new course phase config for the target course phase with the same parameters
	params := db.CreateOrUpdateCoursePhaseConfigParams{
		AssessmentSchemaID:       sourceConfig.AssessmentSchemaID,
		CoursePhaseID:            targetCoursePhaseID,
		Start:                    sourceConfig.Start,
		Deadline:                 sourceConfig.Deadline,
		SelfEvaluationEnabled:    sourceConfig.SelfEvaluationEnabled,
		SelfEvaluationSchema:     sourceConfig.SelfEvaluationSchema,
		SelfEvaluationStart:      sourceConfig.SelfEvaluationStart,
		SelfEvaluationDeadline:   sourceConfig.SelfEvaluationDeadline,
		PeerEvaluationEnabled:    sourceConfig.PeerEvaluationEnabled,
		PeerEvaluationSchema:     sourceConfig.PeerEvaluationSchema,
		PeerEvaluationStart:      sourceConfig.PeerEvaluationStart,
		PeerEvaluationDeadline:   sourceConfig.PeerEvaluationDeadline,
		TutorEvaluationEnabled:   sourceConfig.TutorEvaluationEnabled,
		TutorEvaluationSchema:    sourceConfig.TutorEvaluationSchema,
		TutorEvaluationStart:     sourceConfig.TutorEvaluationStart,
		TutorEvaluationDeadline:  sourceConfig.TutorEvaluationDeadline,
		EvaluationResultsVisible: sourceConfig.EvaluationResultsVisible,
		GradeSuggestionVisible:   pgtype.Bool{Bool: sourceConfig.GradeSuggestionVisible, Valid: true},
		ActionItemsVisible:       pgtype.Bool{Bool: sourceConfig.ActionItemsVisible, Valid: true},
		GradingSheetVisible:      pgtype.Bool{Bool: sourceConfig.GradingSheetVisible, Valid: true},
		AssessmentEnabled:        sourceConfig.AssessmentEnabled,
	}

	err = qtx.CreateOrUpdateCoursePhaseConfig(ctx, params)
	if err != nil {
		log.WithError(err).Error("Failed to create course phase config for target course phase")
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("could not commit phase copy: ", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
