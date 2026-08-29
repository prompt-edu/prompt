package copy

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	log "github.com/sirupsen/logrus"
)

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

type assessmentCopyHandler struct {
	service *CopyService
}

func (h *assessmentCopyHandler) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
	if req.SourceCoursePhaseID == req.TargetCoursePhaseID {
		return nil
	}

	ctx := c.Request.Context()

	tx, err := h.service.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := h.service.queries.WithTx(tx)

	// Get the course phase config from the source course phase
	sourceConfig, err := qtx.GetCoursePhaseConfig(ctx, req.SourceCoursePhaseID)
	if err != nil {
		log.WithError(err).Error("Failed to get source course phase config")
		return err
	}

	// Copying a disabled source must not hide grades a non-empty target already holds
	if !sourceConfig.AssessmentEnabled {
		hasData, err := qtx.PhaseHasAssessmentData(ctx, req.TargetCoursePhaseID)
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
		CoursePhaseID:            req.TargetCoursePhaseID,
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
