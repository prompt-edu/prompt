package assessmentCompletion

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion/assessmentCompletionDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/utils"
	log "github.com/sirupsen/logrus"
)

type AssessmentCompletionService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var AssessmentCompletionServiceSingleton *AssessmentCompletionService

var ErrAssessmentCompleted = errors.New("assessment already completed")

func CheckAssessmentCompletionExists(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) (bool, error) {
	exists, err := AssessmentCompletionServiceSingleton.queries.CheckAssessmentCompletionExists(ctx, db.CheckAssessmentCompletionExistsParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not check assessment completion existence: ", err)
		return false, errors.New("could not check assessment completion existence")
	}
	return exists, nil
}

func CheckAssessmentIsEditable(ctx context.Context, qtx *db.Queries, courseParticipationID, coursePhaseID uuid.UUID) error {
	open, err := coursePhaseConfig.IsAssessmentOpen(ctx, coursePhaseID)
	if err != nil {
		return err
	}
	if !open {
		return coursePhaseConfig.ErrNotStarted
	}

	exists, err := qtx.CheckAssessmentCompletionExists(ctx, db.CheckAssessmentCompletionExistsParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not check assessment completion existence: ", err)
		return errors.New("could not check assessment completion existence")
	}
	if exists {
		completion, err := qtx.GetAssessmentCompletion(ctx, db.GetAssessmentCompletionParams{
			CourseParticipationID: courseParticipationID,
			CoursePhaseID:         coursePhaseID,
		})
		if err != nil {
			log.Error("could not get assessment completion: ", err)
			return errors.New("could not get assessment completion")
		}

		if completion.Completed {
			return ErrAssessmentCompleted
		}
	}
	return nil
}

func CountRemainingAssessmentsForStudent(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) (db.CountRemainingAssessmentsForStudentRow, error) {
	remainingAssessments, err := AssessmentCompletionServiceSingleton.queries.CountRemainingAssessmentsForStudent(ctx, db.CountRemainingAssessmentsForStudentParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not count remaining assessments: ", err)
		return db.CountRemainingAssessmentsForStudentRow{}, errors.New("could not count remaining assessments")
	}
	return remainingAssessments, nil
}

func GetAllGrades(ctx context.Context, coursePhaseID uuid.UUID) ([]assessmentCompletionDTO.GradeWithParticipation, error) {
	grades, err := AssessmentCompletionServiceSingleton.queries.GetAllGrades(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not get grades by course phase: ", err)
		return nil, errors.New("could not get grades by course phase")
	}

	return assessmentCompletionDTO.GetGradesWithParticipationFromDBGradesWithParticipation(grades), nil
}

func GetStudentGrade(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) (float64, error) {
	grade, err := AssessmentCompletionServiceSingleton.queries.GetStudentGrade(ctx, db.GetStudentGradeParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		log.Error("could not get student grade: ", err)
		return 0, errors.New("could not get student grade")
	}
	if !grade.Valid {
		return 0, nil
	}
	return utils.MapNumericToFloat64(grade), nil
}

func CreateOrUpdateAssessmentCompletion(ctx context.Context, req assessmentCompletionDTO.AssessmentCompletion) error {
	err := CheckAssessmentIsEditable(ctx, &AssessmentCompletionServiceSingleton.queries, req.CourseParticipationID, req.CoursePhaseID)
	if err != nil {
		return err
	}

	tx, err := AssessmentCompletionServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := AssessmentCompletionServiceSingleton.queries.WithTx(tx)

	err = qtx.CreateOrUpdateAssessmentCompletion(ctx, db.CreateOrUpdateAssessmentCompletionParams{
		CourseParticipationID: req.CourseParticipationID,
		CoursePhaseID:         req.CoursePhaseID,
		CompletedAt:           pgtype.Timestamptz{Time: time.Now(), Valid: true},
		Author:                req.Author,
		Comment:               req.Comment,
		GradeSuggestion:       utils.MapFloat64ToNumeric(req.GradeSuggestion),
		Completed:             req.Completed,
	})
	if err != nil {
		log.Error("could not create or update assessment completion: ", err)
		return errors.New("could not create or update assessment completion")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("could not commit assessment completion: ", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func MarkAssessmentAsCompleted(ctx context.Context, req assessmentCompletionDTO.AssessmentCompletion) error {
	err := CheckAssessmentIsEditable(ctx, &AssessmentCompletionServiceSingleton.queries, req.CourseParticipationID, req.CoursePhaseID)
	if err != nil {
		return err
	}

	remaining, err := CountRemainingAssessmentsForStudent(ctx, req.CourseParticipationID, req.CoursePhaseID)
	if err != nil {
		log.Error("could not count remaining assessments: ", err)
		return errors.New("could not count remaining assessments")
	}
	if remaining.RemainingAssessments > 0 {
		log.Error("cannot mark assessment as completed, remaining assessments exist")
		return errors.New("cannot mark assessment as completed, remaining assessments exist")
	}

	tx, err := AssessmentCompletionServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := AssessmentCompletionServiceSingleton.queries.WithTx(tx)

	err = qtx.MarkAssessmentAsFinished(ctx, db.MarkAssessmentAsFinishedParams{
		CourseParticipationID: req.CourseParticipationID,
		CoursePhaseID:         req.CoursePhaseID,
		CompletedAt:           pgtype.Timestamptz{Time: time.Now(), Valid: true},
		Author:                req.Author,
	})
	if err != nil {
		log.Error("could not create assessment completion: ", err)
		return errors.New("could not create assessment completion")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("could not commit assessment completion: ", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func UnmarkAssessmentAsCompleted(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) error {
	deadlinePassed, err := coursePhaseConfig.IsAssessmentDeadlinePassed(ctx, coursePhaseID)
	if err != nil {
		return err
	}
	if deadlinePassed {
		return coursePhaseConfig.ErrDeadlinePassed
	}

	err = AssessmentCompletionServiceSingleton.queries.UnmarkAssessmentAsFinished(ctx, db.UnmarkAssessmentAsFinishedParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not unmark assessment as finished: ", err)
		return errors.New("could not unmark assessment as finished")
	}
	return nil
}

func DeleteAssessmentCompletion(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) error {
	err := AssessmentCompletionServiceSingleton.queries.DeleteAssessmentCompletion(ctx, db.DeleteAssessmentCompletionParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not delete assessment completion: ", err)
		return errors.New("could not delete assessment completion")
	}
	return nil
}

func ListAssessmentCompletionsByCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]db.AssessmentCompletion, error) {
	completions, err := AssessmentCompletionServiceSingleton.queries.GetAssessmentCompletionsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not get completions by course phase: ", err)
		return nil, errors.New("could not get completions by course phase")
	}
	return completions, nil
}

func GetAssessmentCompletion(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) (db.AssessmentCompletion, error) {
	completion, err := AssessmentCompletionServiceSingleton.queries.GetAssessmentCompletion(ctx, db.GetAssessmentCompletionParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not get assessment completion: ", err)
		return db.AssessmentCompletion{}, errors.New("could not get assessment completion")
	}
	return completion, nil
}
