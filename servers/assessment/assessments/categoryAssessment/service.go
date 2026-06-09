package categoryAssessment

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/categoryAssessment/categoryAssessmentDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	log "github.com/sirupsen/logrus"
)

type CategoryAssessmentService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var CategoryAssessmentServiceSingleton *CategoryAssessmentService

// ErrNotEditable is returned when the assessment is not editable yet, past
// the deadline, or already finalized. The router maps this to 403.
var ErrNotEditable = errors.New("category assessment is not editable")

func wrapEditabilityError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, assessmentCompletion.ErrAssessmentCompleted) ||
		errors.Is(err, coursePhaseConfig.ErrNotStarted) ||
		errors.Is(err, coursePhaseConfig.ErrDeadlinePassed) {
		return fmt.Errorf("%w: %w", ErrNotEditable, err)
	}
	return err
}

func CreateOrUpdateCategoryAssessment(ctx context.Context, req categoryAssessmentDTO.CreateOrUpdateCategoryAssessmentRequest) error {
	tx, err := CategoryAssessmentServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := CategoryAssessmentServiceSingleton.queries.WithTx(tx)

	if err := wrapEditabilityError(assessmentCompletion.CheckAssessmentIsEditable(ctx, qtx, req.CourseParticipationID, req.CoursePhaseID)); err != nil {
		return err
	}

	if err := qtx.CreateOrUpdateCategoryAssessment(ctx, db.CreateOrUpdateCategoryAssessmentParams{
		CategoryID:            req.CategoryID,
		CoursePhaseID:         req.CoursePhaseID,
		CourseParticipationID: req.CourseParticipationID,
		Comment:               req.Comment,
		Author:                req.Author,
		AuthorID:              req.AuthorID,
	}); err != nil {
		log.Error("could not create or update category assessment: ", err)
		return errors.New("could not create or update category assessment")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("could not commit category assessment transaction: ", err)
		return errors.New("could not commit category assessment")
	}
	return nil
}

func ListCategoryAssessmentsByStudentInPhase(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) ([]db.CategoryAssessment, error) {
	items, err := CategoryAssessmentServiceSingleton.queries.ListCategoryAssessmentsByStudentInPhase(ctx, db.ListCategoryAssessmentsByStudentInPhaseParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not list category assessments for student in phase: ", err)
		return nil, errors.New("could not list category assessments for student in phase")
	}
	return items, nil
}
