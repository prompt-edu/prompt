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
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	log "github.com/sirupsen/logrus"
)

type CategoryAssessmentService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var CategoryAssessmentServiceSingleton *CategoryAssessmentService

func CreateOrUpdateCategoryAssessment(ctx context.Context, req categoryAssessmentDTO.CreateOrUpdateCategoryAssessmentRequest) error {
	tx, err := CategoryAssessmentServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := CategoryAssessmentServiceSingleton.queries.WithTx(tx)

	if err := assessmentCompletion.CheckAssessmentIsEditable(ctx, qtx, req.CourseParticipationID, req.CoursePhaseID); err != nil {
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
		log.Error("could not commit category assessment: ", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
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

func ListCategoryAssessmentsByCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]db.CategoryAssessment, error) {
	items, err := CategoryAssessmentServiceSingleton.queries.ListCategoryAssessmentsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not list category assessments by course phase: ", err)
		return nil, errors.New("could not list category assessments by course phase")
	}
	return items, nil
}

func DeleteCategoryAssessment(ctx context.Context, id uuid.UUID) error {
	tx, err := CategoryAssessmentServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := CategoryAssessmentServiceSingleton.queries.WithTx(tx)

	existing, err := qtx.GetCategoryAssessment(ctx, id)
	if err != nil {
		log.Info("category assessment not found, nothing to delete: ", err)
		return nil
	}

	if err := assessmentCompletion.CheckAssessmentIsEditable(ctx, qtx, existing.CourseParticipationID, existing.CoursePhaseID); err != nil {
		return err
	}

	if err := qtx.DeleteCategoryAssessment(ctx, id); err != nil {
		log.Error("could not delete category assessment: ", err)
		return errors.New("could not delete category assessment")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("could not commit category assessment deletion: ", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
