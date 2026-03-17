package assessments

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/actionItem"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/actionItem/actionItemDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion/assessmentCompletionDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig/coursePhaseConfigDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationDTO"
	"github.com/prompt-edu/prompt/servers/assessment/utils"
	log "github.com/sirupsen/logrus"
)

type AssessmentService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var AssessmentServiceSingleton *AssessmentService
var ErrValidationFailed = errors.New("validation failed: comment and examples are required for Strongly Disagree, Disagree, and Neutral scores")
var ErrInvalidScoreLevel = errors.New("validation failed: scoreLevel is required and must be valid")

func CreateOrUpdateAssessment(ctx context.Context, req assessmentDTO.CreateOrUpdateAssessmentRequest) error {
	// Validate scoreLevel is not empty
	if req.ScoreLevel == "" {
		return ErrInvalidScoreLevel
	}

	// Validate that comment and examples are provided for low score levels
	if isLowScoreLevel(req.ScoreLevel) {
		if req.Comment == "" || req.Examples == "" {
			return ErrValidationFailed
		}
	}

	tx, err := AssessmentServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := AssessmentServiceSingleton.queries.WithTx(tx)

	err = assessmentCompletion.CheckAssessmentIsEditable(ctx, qtx, req.CourseParticipationID, req.CoursePhaseID)
	if err != nil {
		return err
	}

	err = qtx.CreateOrUpdateAssessment(ctx, db.CreateOrUpdateAssessmentParams{
		CourseParticipationID: req.CourseParticipationID,
		CoursePhaseID:         req.CoursePhaseID,
		CompetencyID:          req.CompetencyID,
		ScoreLevel:            scoreLevelDTO.MapDTOtoDBScoreLevel(req.ScoreLevel),
		Examples:              req.Examples,
		Comment:               pgtype.Text{String: req.Comment, Valid: true},
		Author:                req.Author,
	})
	if err != nil {
		log.Error("could not create or update assessment: ", err)
		return errors.New("could not create or update assessment")
	}
	if err := tx.Commit(ctx); err != nil {
		log.Error("could not commit assessment creation/update: ", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func GetAssessment(ctx context.Context, id uuid.UUID) (db.Assessment, error) {
	assessment, err := AssessmentServiceSingleton.queries.GetAssessment(ctx, id)
	if err != nil {
		log.Error("could not get assessment: ", err)
		return db.Assessment{}, errors.New("could not get assessment")
	}
	return assessment, nil
}

func ListAssessmentsByCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]db.Assessment, error) {
	assessments, err := AssessmentServiceSingleton.queries.ListAssessmentsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not get assessments by course phase: ", err)
		return nil, errors.New("could not get assessments by course phase")
	}
	return assessments, nil
}

func ListAssessmentsByStudentInPhase(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) ([]db.Assessment, error) {
	assessments, err := AssessmentServiceSingleton.queries.ListAssessmentsByStudentInPhase(ctx, db.ListAssessmentsByStudentInPhaseParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not get assessments for student in phase: ", err)
		return nil, errors.New("could not get assessments for student in phase")
	}
	return assessments, nil
}

func GetStudentAssessment(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID) (assessmentDTO.StudentAssessment, error) {
	assessments, err := ListAssessmentsByStudentInPhase(ctx, courseParticipationID, coursePhaseID)
	if err != nil {
		log.Error("could not get assessments for student in phase: ", err)
		return assessmentDTO.StudentAssessment{}, errors.New("could not get assessments for student in phase")
	}

	var completion = assessmentCompletionDTO.AssessmentCompletion{}
	var studentScore = scoreLevelDTO.StudentScore{
		ScoreLevel:   scoreLevelDTO.ScoreLevelVeryBad,
		ScoreNumeric: pgtype.Float8{Float64: 0.0, Valid: true},
	}

	exists, err := assessmentCompletion.CheckAssessmentCompletionExists(ctx, courseParticipationID, coursePhaseID)
	if err != nil {
		log.Error("could not check assessment completion existence: ", err)
		return assessmentDTO.StudentAssessment{}, errors.New("could not check assessment completion existence")
	}

	if exists {
		dbAssessmentCompletion, err := assessmentCompletion.GetAssessmentCompletion(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not get assessment completion: ", err)
			return assessmentDTO.StudentAssessment{}, errors.New("could not get assessment completion")
		}
		completion = assessmentCompletionDTO.MapDBAssessmentCompletionToAssessmentCompletionDTO(dbAssessmentCompletion)
	}

	if len(assessments) > 0 {
		studentScore, err = scoreLevel.GetStudentScore(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not get score level: ", err)
			return assessmentDTO.StudentAssessment{}, errors.New("could not get score level")
		}
	}

	evaluations, err := evaluations.GetEvaluationsForParticipantInPhase(ctx, courseParticipationID, coursePhaseID)
	if err != nil {
		log.Error("could not get evaluations: ", err)
		return assessmentDTO.StudentAssessment{}, errors.New("could not get evaluations")
	}

	if evaluations == nil {
		evaluations = []evaluationDTO.Evaluation{}
	}

	return assessmentDTO.StudentAssessment{
		CourseParticipationID: courseParticipationID,
		Assessments:           assessmentDTO.GetAssessmentDTOsFromDBModels(assessments),
		AssessmentCompletion:  completion,
		StudentScore:          studentScore,
		Evaluations:           evaluations,
	}, nil
}

func GetStudentAssessmentResults(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID, config coursePhaseConfigDTO.CoursePhaseConfig) (assessmentDTO.StudentAssessmentResults, error) {
	var results assessmentDTO.StudentAssessmentResults
	var err error

	assessments := []db.Assessment{}
	if config.GradingSheetVisible {
		assessments, err = ListAssessmentsByStudentInPhase(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not get assessments for student in phase: ", err)
			return results, errors.New("could not get assessments for student in phase")
		}
	}

	completion := db.AssessmentCompletion{}
	exists, err := assessmentCompletion.CheckAssessmentCompletionExists(ctx, courseParticipationID, coursePhaseID)
	if err != nil {
		log.Error("could not check assessment completion existence: ", err)
		return results, errors.New("could not check assessment completion existence")
	}
	if exists {
		completion, err = assessmentCompletion.GetAssessmentCompletion(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not get assessment completion: ", err)
			return results, errors.New("could not get assessment completion")
		}
	}

	studentScore := scoreLevelDTO.StudentScore{
		ScoreLevel:   scoreLevelDTO.ScoreLevelVeryBad,
		ScoreNumeric: pgtype.Float8{Float64: 0.0, Valid: true},
	}
	if len(assessments) > 0 {
		studentScore, err = scoreLevel.GetStudentScore(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not get score level: ", err)
			return results, errors.New("could not get score level")
		}
	}

	var evals []evaluationDTO.Evaluation
	if config.GradingSheetVisible {
		evals, err = evaluations.GetEvaluationsForParticipantInPhase(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not get evaluations for participant in phase: ", err)
			return results, errors.New("could not get evaluations for participant in phase")
		}
	}

	var actionItems []actionItemDTO.ActionItem
	if config.ActionItemsVisible {
		actionItems, err = actionItem.ListActionItemsForStudentInPhase(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not list action items for student in phase: ", err)
			return results, errors.New("could not list action items for student in phase")
		}
	}

	peerEvalResults := []assessmentDTO.AggregatedEvaluationResult{}
	selfEvalResults := []assessmentDTO.AggregatedEvaluationResult{}
	if config.GradingSheetVisible {
		peerEvalResults = aggregateEvaluations(evals, assessmentType.Peer)
		selfEvalResults = aggregateEvaluations(evals, assessmentType.Self)
	}

	if !config.GradeSuggestionVisible {
		completion.GradeSuggestion = utils.MapFloat64ToNumeric(0.0)
		completion.Comment = ""
	}

	results = assessmentDTO.StudentAssessmentResults{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
		Assessments:           assessmentDTO.GetAssessmentDTOsFromDBModels(assessments),
		AssessmentCompletion:  assessmentCompletionDTO.MapDBAssessmentCompletionToAssessmentCompletionDTO(completion),
		StudentScore:          studentScore,
		PeerEvaluationResults: peerEvalResults,
		SelfEvaluationResults: selfEvalResults,
		ActionItems:           actionItems,
	}

	return results, nil
}

func DeleteAssessment(ctx context.Context, id uuid.UUID) error {
	tx, err := AssessmentServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := AssessmentServiceSingleton.queries.WithTx(tx)

	assessment, err := qtx.GetAssessment(ctx, id)
	if err != nil {
		log.Info("assessment not found, nothing to delete: ", err)
		return nil
	}

	err = assessmentCompletion.CheckAssessmentIsEditable(ctx, qtx, assessment.CourseParticipationID, assessment.CoursePhaseID)
	if err != nil {
		return err
	}

	err = qtx.DeleteAssessment(ctx, id)
	if err != nil {
		log.Error("could not delete assessment: ", err)
		return errors.New("could not delete assessment")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("could not commit assessment deletion: ", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func isLowScoreLevel(scoreLevel scoreLevelDTO.ScoreLevel) bool {
	return scoreLevel == scoreLevelDTO.ScoreLevelVeryBad ||
		scoreLevel == scoreLevelDTO.ScoreLevelBad ||
		scoreLevel == scoreLevelDTO.ScoreLevelOk
}

func aggregateEvaluations(evals []evaluationDTO.Evaluation, targetType assessmentType.AssessmentType) []assessmentDTO.AggregatedEvaluationResult {
	type accumulator struct {
		sum   float64
		count int
	}

	aggregated := make(map[uuid.UUID]accumulator)
	for _, eval := range evals {
		if eval.Type != targetType {
			continue
		}
		num := scoreLevelDTO.MapScoreLevelToNumber(eval.ScoreLevel)
		current := aggregated[eval.CompetencyID]
		current.sum += num
		current.count++
		aggregated[eval.CompetencyID] = current
	}

	results := make([]assessmentDTO.AggregatedEvaluationResult, 0, len(aggregated))
	for competencyID, acc := range aggregated {
		if acc.count == 0 {
			continue
		}
		avg := acc.sum / float64(acc.count)
		results = append(results, assessmentDTO.AggregatedEvaluationResult{
			CompetencyID:        competencyID,
			AverageScoreNumeric: avg,
		})
	}

	return results
}
