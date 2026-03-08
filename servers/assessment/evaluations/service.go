package evaluations

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationDTO"
	log "github.com/sirupsen/logrus"
)

type EvaluationService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var EvaluationServiceSingleton *EvaluationService

func CreateOrUpdateEvaluation(ctx context.Context, coursePhaseID uuid.UUID, req evaluationDTO.CreateOrUpdateEvaluationRequest) error {
	tx, err := EvaluationServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := EvaluationServiceSingleton.queries.WithTx(tx)

	err = evaluationCompletion.CheckEvaluationIsEditable(ctx, qtx, req.CourseParticipationID, coursePhaseID, req.AuthorCourseParticipationID, req.Type)
	if err != nil {
		return err
	}

	err = qtx.CreateOrUpdateEvaluation(ctx, db.CreateOrUpdateEvaluationParams{
		CourseParticipationID:       req.CourseParticipationID,
		CoursePhaseID:               coursePhaseID,
		CompetencyID:                req.CompetencyID,
		ScoreLevel:                  scoreLevelDTO.MapDTOtoDBScoreLevel(req.ScoreLevel),
		AuthorCourseParticipationID: req.AuthorCourseParticipationID,
		Type:                        assessmentType.MapDTOtoDBAssessmentType(req.Type),
	})
	if err != nil {
		log.Error("could not create or update evaluation: ", err)
		return errors.New("could not create or update evaluation")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func DeleteEvaluation(ctx context.Context, id uuid.UUID) error {
	tx, err := EvaluationServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := EvaluationServiceSingleton.queries.WithTx(tx)

	evaluation, err := qtx.GetEvaluationByID(ctx, id)
	if err != nil {
		log.Error("could not get evaluation by ID: ", err)
		return errors.New("could not get evaluation by ID")
	}

	err = evaluationCompletion.CheckEvaluationIsEditable(ctx, qtx, evaluation.CourseParticipationID, evaluation.CoursePhaseID, evaluation.AuthorCourseParticipationID, assessmentType.MapDBAssessmentTypeToDTO(evaluation.Type))
	if err != nil {
		return err
	}

	err = qtx.DeleteEvaluation(ctx, id)
	if err != nil {
		log.Error("could not delete evaluation: ", err)
		return errors.New("could not delete evaluation")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func GetEvaluationsByPhase(ctx context.Context, coursePhaseID uuid.UUID) ([]evaluationDTO.Evaluation, error) {
	evaluations, err := EvaluationServiceSingleton.queries.GetEvaluationsByPhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not get evaluations by phase: ", err)
		return nil, errors.New("could not get evaluations by phase")
	}
	return evaluationDTO.MapToEvaluationDTOs(evaluations), nil
}

func GetEvaluationsForParticipantInPhase(ctx context.Context, courseParticipationID uuid.UUID, coursePhaseID uuid.UUID) ([]evaluationDTO.Evaluation, error) {
	evaluations, err := EvaluationServiceSingleton.queries.GetEvaluationsForParticipantInPhase(ctx, db.GetEvaluationsForParticipantInPhaseParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not get evaluations for participant in phase: ", err)
		return nil, errors.New("could not get evaluations for participant in phase")
	}
	return evaluationDTO.MapToEvaluationDTOs(evaluations), nil
}

func GetEvaluationsForTutorInPhase(ctx context.Context, courseParticipationID uuid.UUID, coursePhaseID uuid.UUID) ([]evaluationDTO.Evaluation, error) {
	evaluations, err := EvaluationServiceSingleton.queries.GetEvaluationsForTutorInPhase(ctx, db.GetEvaluationsForTutorInPhaseParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not get evaluations for tutor in phase: ", err)
		return nil, errors.New("could not get evaluations for tutor in phase")
	}
	return evaluationDTO.MapToEvaluationDTOs(evaluations), nil
}

func GetEvaluationsForAuthorInPhase(ctx context.Context, courseParticipationID uuid.UUID, coursePhaseID uuid.UUID) ([]evaluationDTO.Evaluation, error) {
	evaluations, err := EvaluationServiceSingleton.queries.GetEvaluationsForAuthorInPhase(ctx, db.GetEvaluationsForAuthorInPhaseParams{
		AuthorCourseParticipationID: courseParticipationID,
		CoursePhaseID:               coursePhaseID,
	})
	if err != nil {
		log.Error("could not get evaluations for author in phase: ", err)
		return nil, errors.New("could not get evaluations for author in phase")
	}
	return evaluationDTO.MapToEvaluationDTOs(evaluations), nil
}

func GetEvaluationByID(ctx context.Context, id uuid.UUID) (evaluationDTO.Evaluation, error) {
	evaluation, err := EvaluationServiceSingleton.queries.GetEvaluationByID(ctx, id)
	if err != nil {
		log.Error("could not get evaluation by ID: ", err)
		return evaluationDTO.Evaluation{}, errors.New("could not get evaluation by ID")
	}
	return evaluationDTO.MapDBToEvaluationDTO(evaluation), nil
}
