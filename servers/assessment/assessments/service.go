package assessments

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/actionItem"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/actionItem/actionItemDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion/assessmentCompletionDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/categoryAssessment"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/categoryAssessment/categoryAssessmentDTO"
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
var ErrInvalidScoreLevel = errors.New("validation failed: scoreLevel is required and must be valid")
var ErrUnsupportedAssessmentExportFormat = errors.New("unsupported assessment export format")
var ErrAssessmentNotInPhase = errors.New("assessment does not belong to this course phase")
var ErrAssessmentNotFound = errors.New("assessment not found")

const AssessmentExportFormatJSON = "json"

func CreateOrUpdateAssessment(ctx context.Context, req assessmentDTO.CreateOrUpdateAssessmentRequest) error {
	if req.ScoreLevel == "" {
		return ErrInvalidScoreLevel
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
		Author:                req.Author,
		AuthorID:              req.AuthorID,
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

	categoryAssessments, err := categoryAssessment.ListCategoryAssessmentsByStudentInPhase(ctx, courseParticipationID, coursePhaseID)
	if err != nil {
		log.Error("could not get category assessments for student in phase: ", err)
		return assessmentDTO.StudentAssessment{}, errors.New("could not get category assessments for student in phase")
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
		CategoryAssessments:   categoryAssessmentDTO.GetCategoryAssessmentDTOsFromDBModels(categoryAssessments),
		AssessmentCompletion:  completion,
		StudentScore:          studentScore,
		Evaluations:           evaluations,
	}, nil
}

func ExportStudentAssessment(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID, format string) (assessmentDTO.AssessmentExport, error) {
	if format != AssessmentExportFormatJSON {
		return assessmentDTO.AssessmentExport{}, ErrUnsupportedAssessmentExportFormat
	}

	assessments, err := ListAssessmentsByStudentInPhase(ctx, courseParticipationID, coursePhaseID)
	if err != nil {
		log.Error("could not get assessments for export: ", err)
		return assessmentDTO.AssessmentExport{}, errors.New("could not get assessments for export")
	}

	catAssessments, err := categoryAssessment.ListCategoryAssessmentsByStudentInPhase(ctx, courseParticipationID, coursePhaseID)
	if err != nil {
		log.Error("could not get category assessments for export: ", err)
		return assessmentDTO.AssessmentExport{}, errors.New("could not get category assessments for export")
	}

	completion := assessmentCompletionDTO.AssessmentCompletion{}
	completionExists, err := assessmentCompletion.CheckAssessmentCompletionExists(ctx, courseParticipationID, coursePhaseID)
	if err != nil {
		log.Error("could not check assessment completion existence: ", err)
		return assessmentDTO.AssessmentExport{}, errors.New("could not check assessment completion existence")
	}
	if completionExists {
		dbCompletion, err := assessmentCompletion.GetAssessmentCompletion(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not get assessment completion: ", err)
			return assessmentDTO.AssessmentExport{}, errors.New("could not get assessment completion")
		}
		completion = assessmentCompletionDTO.MapDBAssessmentCompletionToAssessmentCompletionDTO(dbCompletion)
	}

	studentScore := scoreLevelDTO.StudentScore{
		ScoreLevel:   scoreLevelDTO.ScoreLevelVeryBad,
		ScoreNumeric: pgtype.Float8{Float64: 0.0, Valid: true},
	}
	if len(assessments) > 0 {
		studentScore, err = scoreLevel.GetStudentScore(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not get score level: ", err)
			return assessmentDTO.AssessmentExport{}, errors.New("could not get score level")
		}
	}

	actionItems, err := getAssessmentExportActionItems(ctx, coursePhaseID, courseParticipationID)
	if err != nil {
		log.Error("could not get assessment export action items: ", err)
		return assessmentDTO.AssessmentExport{}, errors.New("could not get assessment export action items")
	}

	return assessmentDTO.AssessmentExport{
		ExportedAt:            time.Now().UTC(),
		CoursePhaseID:         coursePhaseID,
		CourseParticipationID: courseParticipationID,
		StudentAssessment: assessmentDTO.StudentAssessment{
			CourseParticipationID: courseParticipationID,
			Assessments:           assessmentDTO.GetAssessmentDTOsFromDBModels(assessments),
			CategoryAssessments:   categoryAssessmentDTO.GetCategoryAssessmentDTOsFromDBModels(catAssessments),
			AssessmentCompletion:  completion,
			StudentScore:          studentScore,
			Evaluations:           []evaluationDTO.Evaluation{},
		},
		ActionItems: actionItems,
	}, nil
}

func getAssessmentExportActionItems(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID) ([]actionItemDTO.ActionItem, error) {
	actionItems, err := AssessmentServiceSingleton.queries.ListActionItemsForStudentInPhase(ctx, db.ListActionItemsForStudentInPhaseParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		return nil, err
	}

	return actionItemDTO.GetActionItemDTOsFromDBModels(actionItems), nil
}

func GetStudentAssessmentResults(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID, config coursePhaseConfigDTO.CoursePhaseConfig) (assessmentDTO.StudentAssessmentResults, error) {
	var results assessmentDTO.StudentAssessmentResults
	var err error

	assessments := []db.Assessment{}
	categoryAssessments := []db.CategoryAssessment{}
	if config.GradingSheetVisible {
		assessments, err = ListAssessmentsByStudentInPhase(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not get assessments for student in phase: ", err)
			return results, errors.New("could not get assessments for student in phase")
		}
		categoryAssessments, err = categoryAssessment.ListCategoryAssessmentsByStudentInPhase(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not get category assessments for student in phase: ", err)
			return results, errors.New("could not get category assessments for student in phase")
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

	var studentScore *scoreLevelDTO.StudentScore
	if config.GradingSheetVisible && len(assessments) > 0 {
		score, err := scoreLevel.GetStudentScore(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not get score level: ", err)
			return results, errors.New("could not get score level")
		}
		studentScore = &score
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
		CategoryAssessments:   categoryAssessmentDTO.GetCategoryAssessmentDTOsFromDBModels(categoryAssessments),
		AssessmentCompletion:  assessmentCompletionDTO.MapDBAssessmentCompletionToAssessmentCompletionDTO(completion),
		StudentScore:          studentScore,
		PeerEvaluationResults: peerEvalResults,
		SelfEvaluationResults: selfEvalResults,
		ActionItems:           actionItems,
	}

	return results, nil
}

func DeleteAssessment(ctx context.Context, id, coursePhaseID uuid.UUID) error {
	tx, err := AssessmentServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := AssessmentServiceSingleton.queries.WithTx(tx)

	assessment, err := qtx.GetAssessment(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAssessmentNotFound
		}
		log.Error("could not get assessment for deletion: ", err)
		return fmt.Errorf("could not get assessment: %w", err)
	}

	if assessment.CoursePhaseID != coursePhaseID {
		return ErrAssessmentNotInPhase
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
