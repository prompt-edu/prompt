package evaluationCompletion

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationCompletion/evaluationCompletionDTO"
	log "github.com/sirupsen/logrus"
)

type EvaluationCompletionService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var EvaluationCompletionServiceSingleton *EvaluationCompletionService

func CheckEvaluationIsEditable(ctx context.Context, qtx *db.Queries, courseParticipationID, coursePhaseID, authorCourseParticipationID uuid.UUID, evaluationType assessmentType.AssessmentType) error {
	switch evaluationType {
	case assessmentType.Self:
		open, err := coursePhaseConfig.IsSelfEvaluationOpen(ctx, coursePhaseID)
		if err != nil {
			return err
		}
		if !open {
			return coursePhaseConfig.ErrNotStarted
		}
	case assessmentType.Peer:
		open, err := coursePhaseConfig.IsPeerEvaluationOpen(ctx, coursePhaseID)
		if err != nil {
			return err
		}
		if !open {
			return coursePhaseConfig.ErrNotStarted
		}
	case assessmentType.Tutor:
		open, err := coursePhaseConfig.IsTutorEvaluationOpen(ctx, coursePhaseID)
		if err != nil {
			return err
		}
		if !open {
			return coursePhaseConfig.ErrNotStarted
		}
	}

	exists, err := qtx.CheckEvaluationCompletionExists(ctx, db.CheckEvaluationCompletionExistsParams{
		CourseParticipationID:       courseParticipationID,
		CoursePhaseID:               coursePhaseID,
		AuthorCourseParticipationID: authorCourseParticipationID,
	})
	if err != nil {
		log.Error("could not check evaluation completion existence: ", err)
		return errors.New("could not check evaluation completion existence")
	}
	if exists {
		completion, err := qtx.GetEvaluationCompletion(ctx, db.GetEvaluationCompletionParams{
			CourseParticipationID:       courseParticipationID,
			CoursePhaseID:               coursePhaseID,
			AuthorCourseParticipationID: authorCourseParticipationID,
		})
		if err != nil {
			log.Error("could not get evaluation completion: ", err)
			return errors.New("could not get evaluation completion")
		}

		if completion.Completed {
			log.Error("evaluation completion already exists and is marked as completed")
			return errors.New("evaluation completion already exists and is marked as completed")
		}
	}
	return nil
}

func CreateOrUpdateEvaluationCompletion(ctx context.Context, req evaluationCompletionDTO.EvaluationCompletion) error {
	err := CheckEvaluationIsEditable(ctx, &EvaluationCompletionServiceSingleton.queries, req.CourseParticipationID, req.CoursePhaseID, req.AuthorCourseParticipationID, req.Type)
	if err != nil {
		return err
	}

	tx, err := EvaluationCompletionServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := EvaluationCompletionServiceSingleton.queries.WithTx(tx)

	err = qtx.CreateOrUpdateEvaluationCompletion(ctx, db.CreateOrUpdateEvaluationCompletionParams{
		CourseParticipationID:       req.CourseParticipationID,
		CoursePhaseID:               req.CoursePhaseID,
		AuthorCourseParticipationID: req.AuthorCourseParticipationID,
		CompletedAt:                 pgtype.Timestamptz{Time: time.Now(), Valid: true},
		Completed:                   req.Completed,
	})
	if err != nil {
		log.Error("could not create or update evaluation completion: ", err)
		return errors.New("could not create or update evaluation completion")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("could not commit evaluation completion: ", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func MarkEvaluationAsCompleted(ctx context.Context, req evaluationCompletionDTO.EvaluationCompletion) error {
	err := CheckEvaluationIsEditable(ctx, &EvaluationCompletionServiceSingleton.queries, req.CourseParticipationID, req.CoursePhaseID, req.AuthorCourseParticipationID, req.Type)
	if err != nil {
		return err
	}

	// Check if there are remaining evaluations before marking as completed
	remainingEvaluations, err := EvaluationCompletionServiceSingleton.queries.CountRemainingEvaluationsForStudent(ctx, db.CountRemainingEvaluationsForStudentParams{
		CourseParticipationID:       req.CourseParticipationID,
		AuthorCourseParticipationID: req.AuthorCourseParticipationID,
		CoursePhaseID:               req.CoursePhaseID,
		Column4:                     assessmentType.MapDTOtoDBAssessmentType(req.Type),
	})
	if err != nil {
		log.Error("could not check remaining evaluations: ", err)
		return errors.New("could not check remaining evaluations")
	}

	if remainingEvaluations > 0 {
		log.Warnf("cannot mark evaluation as completed: %d evaluations still remaining", remainingEvaluations)
		return fmt.Errorf("cannot mark evaluation as completed: %d evaluations still remaining", remainingEvaluations)
	}

	tx, err := EvaluationCompletionServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := EvaluationCompletionServiceSingleton.queries.WithTx(tx)

	err = qtx.MarkEvaluationAsFinished(ctx, db.MarkEvaluationAsFinishedParams{
		CourseParticipationID:       req.CourseParticipationID,
		CoursePhaseID:               req.CoursePhaseID,
		AuthorCourseParticipationID: req.AuthorCourseParticipationID,
		CompletedAt:                 pgtype.Timestamptz{Time: time.Now(), Valid: true},
		Type:                        assessmentType.MapDTOtoDBAssessmentType(req.Type),
	})
	if err != nil {
		log.Error("could not mark evaluation as finished: ", err)
		return errors.New("could not mark evaluation as finished")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("could not commit evaluation completion: ", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func UnmarkEvaluationAsCompleted(ctx context.Context, courseParticipationID, coursePhaseID, authorCourseParticipationID uuid.UUID) error {
	// Get the evaluation completion to determine its type
	dbCompletion, err := EvaluationCompletionServiceSingleton.queries.GetEvaluationCompletion(ctx, db.GetEvaluationCompletionParams{
		CourseParticipationID:       courseParticipationID,
		CoursePhaseID:               coursePhaseID,
		AuthorCourseParticipationID: authorCourseParticipationID,
	})
	if err != nil {
		log.Error("could not get evaluation completion: ", err)
		return errors.New("could not get evaluation completion")
	}

	completion := evaluationCompletionDTO.MapDBEvaluationCompletionToEvaluationCompletionDTO(dbCompletion)
	switch completion.Type {
	case assessmentType.Self:
		deadlinePassed, err := coursePhaseConfig.IsSelfEvaluationDeadlinePassed(ctx, coursePhaseID)
		if err != nil {
			return err
		}
		if deadlinePassed {
			return coursePhaseConfig.ErrDeadlinePassed
		}
	case assessmentType.Peer:
		deadlinePassed, err := coursePhaseConfig.IsPeerEvaluationDeadlinePassed(ctx, coursePhaseID)
		if err != nil {
			return err
		}
		if deadlinePassed {
			return coursePhaseConfig.ErrDeadlinePassed
		}
	case assessmentType.Tutor:
		deadlinePassed, err := coursePhaseConfig.IsTutorEvaluationDeadlinePassed(ctx, coursePhaseID)
		if err != nil {
			return err
		}
		if deadlinePassed {
			return coursePhaseConfig.ErrDeadlinePassed
		}
	}

	err = EvaluationCompletionServiceSingleton.queries.UnmarkEvaluationAsFinished(ctx, db.UnmarkEvaluationAsFinishedParams{
		CourseParticipationID:       courseParticipationID,
		CoursePhaseID:               coursePhaseID,
		AuthorCourseParticipationID: authorCourseParticipationID,
	})
	if err != nil {
		log.Error("could not unmark evaluation as finished: ", err)
		return errors.New("could not unmark evaluation as finished")
	}
	return nil
}

func DeleteEvaluationCompletion(ctx context.Context, courseParticipationID, coursePhaseID, authorCourseParticipationID uuid.UUID) error {
	err := EvaluationCompletionServiceSingleton.queries.DeleteEvaluationCompletion(ctx, db.DeleteEvaluationCompletionParams{
		CourseParticipationID:       courseParticipationID,
		CoursePhaseID:               coursePhaseID,
		AuthorCourseParticipationID: authorCourseParticipationID,
	})
	if err != nil {
		log.Error("could not delete evaluation completion: ", err)
		return errors.New("could not delete evaluation completion")
	}
	return nil
}

func ListEvaluationCompletionsByCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]db.EvaluationCompletion, error) {
	completions, err := EvaluationCompletionServiceSingleton.queries.GetEvaluationCompletionsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not get evaluation completions by course phase: ", err)
		return nil, errors.New("could not get evaluation completions by course phase")
	}
	return completions, nil
}

func GetEvaluationCompletionForParticipantInPhase(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) ([]db.EvaluationCompletion, error) {
	completions, err := EvaluationCompletionServiceSingleton.queries.GetEvaluationCompletionsForParticipantInPhase(ctx, db.GetEvaluationCompletionsForParticipantInPhaseParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not get evaluation completions for participant in phase: ", err)
		return nil, errors.New("could not get evaluation completions for participant in phase")
	}
	return completions, nil
}

func GetEvaluationCompletionsForAuthorInPhase(ctx context.Context, authorCourseParticipationID, coursePhaseID uuid.UUID) ([]db.EvaluationCompletion, error) {
	completions, err := EvaluationCompletionServiceSingleton.queries.GetEvaluationCompletionsForAuthorInPhase(ctx, db.GetEvaluationCompletionsForAuthorInPhaseParams{
		AuthorCourseParticipationID: authorCourseParticipationID,
		CoursePhaseID:               coursePhaseID,
	})
	if err != nil {
		log.Error("could not get evaluation completions for author in phase: ", err)
		return nil, errors.New("could not get evaluation completions for author in phase")
	}
	return completions, nil
}
