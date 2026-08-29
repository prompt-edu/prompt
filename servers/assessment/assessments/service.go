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
	"github.com/prompt-edu/prompt/servers/assessment/assessments/actionItem/actionItemDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion/assessmentCompletionDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/categoryAssessment/categoryAssessmentDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig/coursePhaseConfigDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationDTO"
	"github.com/prompt-edu/prompt/servers/assessment/utils"
	log "github.com/sirupsen/logrus"
)

type assessmentCompletionProvider interface {
	CheckAssessmentIsEditable(ctx context.Context, qtx *db.Queries, courseParticipationID, coursePhaseID uuid.UUID) error
	CheckAssessmentCompletionExists(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) (bool, error)
	GetAssessmentCompletion(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) (db.AssessmentCompletion, error)
}

type categoryAssessmentProvider interface {
	ListCategoryAssessmentsByStudentInPhase(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) ([]db.CategoryAssessment, error)
}

type actionItemProvider interface {
	ListActionItemsForStudentInPhase(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) ([]actionItemDTO.ActionItem, error)
}

type scoreLevelProvider interface {
	GetStudentScore(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) (scoreLevelDTO.StudentScore, error)
}

type AssessmentService struct {
	queries              db.Queries
	conn                 *pgxpool.Pool
	assessmentCompletion assessmentCompletionProvider
	categoryAssessment   categoryAssessmentProvider
	actionItem           actionItemProvider
	scoreLevel           scoreLevelProvider
}

func NewAssessmentService(
	queries db.Queries,
	conn *pgxpool.Pool,
	assessmentCompletion assessmentCompletionProvider,
	categoryAssessment categoryAssessmentProvider,
	actionItem actionItemProvider,
	scoreLevel scoreLevelProvider,
) *AssessmentService {
	return &AssessmentService{
		queries:              queries,
		conn:                 conn,
		assessmentCompletion: assessmentCompletion,
		categoryAssessment:   categoryAssessment,
		actionItem:           actionItem,
		scoreLevel:           scoreLevel,
	}
}

var ErrInvalidScoreLevel = errors.New("validation failed: scoreLevel is required and must be valid")
var ErrUnsupportedAssessmentExportFormat = errors.New("unsupported assessment export format")
var ErrAssessmentNotInPhase = errors.New("assessment does not belong to this course phase")
var ErrAssessmentNotFound = errors.New("assessment not found")

const AssessmentExportFormatJSON = "json"

func (s *AssessmentService) CreateOrUpdateAssessment(ctx context.Context, req assessmentDTO.CreateOrUpdateAssessmentRequest) error {
	if req.ScoreLevel == "" {
		return ErrInvalidScoreLevel
	}

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := s.queries.WithTx(tx)

	err = s.assessmentCompletion.CheckAssessmentIsEditable(ctx, qtx, req.CourseParticipationID, req.CoursePhaseID)
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

func (s *AssessmentService) GetAssessment(ctx context.Context, id uuid.UUID) (db.Assessment, error) {
	assessment, err := s.queries.GetAssessment(ctx, id)
	if err != nil {
		log.Error("could not get assessment: ", err)
		return db.Assessment{}, errors.New("could not get assessment")
	}
	return assessment, nil
}

func (s *AssessmentService) ListAssessmentsByCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]db.Assessment, error) {
	assessments, err := s.queries.ListAssessmentsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not get assessments by course phase: ", err)
		return nil, errors.New("could not get assessments by course phase")
	}
	return assessments, nil
}

func (s *AssessmentService) ListAssessmentsByStudentInPhase(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) ([]db.Assessment, error) {
	assessments, err := s.queries.ListAssessmentsByStudentInPhase(ctx, db.ListAssessmentsByStudentInPhaseParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not get assessments for student in phase: ", err)
		return nil, errors.New("could not get assessments for student in phase")
	}
	return assessments, nil
}

func (s *AssessmentService) GetStudentAssessment(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID) (assessmentDTO.StudentAssessment, error) {
	assessments, err := s.ListAssessmentsByStudentInPhase(ctx, courseParticipationID, coursePhaseID)
	if err != nil {
		log.Error("could not get assessments for student in phase: ", err)
		return assessmentDTO.StudentAssessment{}, errors.New("could not get assessments for student in phase")
	}

	categoryAssessments, err := s.categoryAssessment.ListCategoryAssessmentsByStudentInPhase(ctx, courseParticipationID, coursePhaseID)
	if err != nil {
		log.Error("could not get category assessments for student in phase: ", err)
		return assessmentDTO.StudentAssessment{}, errors.New("could not get category assessments for student in phase")
	}

	var completion = assessmentCompletionDTO.AssessmentCompletion{}
	var studentScore = scoreLevelDTO.StudentScore{
		ScoreLevel:   scoreLevelDTO.ScoreLevelVeryBad,
		ScoreNumeric: pgtype.Float8{Float64: 0.0, Valid: true},
	}

	exists, err := s.assessmentCompletion.CheckAssessmentCompletionExists(ctx, courseParticipationID, coursePhaseID)
	if err != nil {
		log.Error("could not check assessment completion existence: ", err)
		return assessmentDTO.StudentAssessment{}, errors.New("could not check assessment completion existence")
	}

	if exists {
		dbAssessmentCompletion, err := s.assessmentCompletion.GetAssessmentCompletion(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not get assessment completion: ", err)
			return assessmentDTO.StudentAssessment{}, errors.New("could not get assessment completion")
		}
		completion = assessmentCompletionDTO.MapDBAssessmentCompletionToAssessmentCompletionDTO(dbAssessmentCompletion)
	}

	if len(assessments) > 0 {
		studentScore, err = s.scoreLevel.GetStudentScore(ctx, courseParticipationID, coursePhaseID)
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

func (s *AssessmentService) ExportStudentAssessment(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID, format string) (assessmentDTO.AssessmentExport, error) {
	if format != AssessmentExportFormatJSON {
		return assessmentDTO.AssessmentExport{}, ErrUnsupportedAssessmentExportFormat
	}

	assessments, err := s.ListAssessmentsByStudentInPhase(ctx, courseParticipationID, coursePhaseID)
	if err != nil {
		log.Error("could not get assessments for export: ", err)
		return assessmentDTO.AssessmentExport{}, errors.New("could not get assessments for export")
	}

	catAssessments, err := s.categoryAssessment.ListCategoryAssessmentsByStudentInPhase(ctx, courseParticipationID, coursePhaseID)
	if err != nil {
		log.Error("could not get category assessments for export: ", err)
		return assessmentDTO.AssessmentExport{}, errors.New("could not get category assessments for export")
	}

	completion := assessmentCompletionDTO.AssessmentCompletion{}
	completionExists, err := s.assessmentCompletion.CheckAssessmentCompletionExists(ctx, courseParticipationID, coursePhaseID)
	if err != nil {
		log.Error("could not check assessment completion existence: ", err)
		return assessmentDTO.AssessmentExport{}, errors.New("could not check assessment completion existence")
	}
	if completionExists {
		dbCompletion, err := s.assessmentCompletion.GetAssessmentCompletion(ctx, courseParticipationID, coursePhaseID)
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
		studentScore, err = s.scoreLevel.GetStudentScore(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not get score level: ", err)
			return assessmentDTO.AssessmentExport{}, errors.New("could not get score level")
		}
	}

	actionItems, err := s.getAssessmentExportActionItems(ctx, coursePhaseID, courseParticipationID)
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

func (s *AssessmentService) getAssessmentExportActionItems(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID) ([]actionItemDTO.ActionItem, error) {
	actionItems, err := s.queries.ListActionItemsForStudentInPhase(ctx, db.ListActionItemsForStudentInPhaseParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		return nil, err
	}

	return actionItemDTO.GetActionItemDTOsFromDBModels(actionItems), nil
}

func (s *AssessmentService) GetStudentAssessmentResults(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID, config coursePhaseConfigDTO.CoursePhaseConfig) (assessmentDTO.StudentAssessmentResults, error) {
	var results assessmentDTO.StudentAssessmentResults
	var err error

	assessments := []db.Assessment{}
	categoryAssessments := []db.CategoryAssessment{}
	if config.GradingSheetVisible {
		assessments, err = s.ListAssessmentsByStudentInPhase(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not get assessments for student in phase: ", err)
			return results, errors.New("could not get assessments for student in phase")
		}
		categoryAssessments, err = s.categoryAssessment.ListCategoryAssessmentsByStudentInPhase(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not get category assessments for student in phase: ", err)
			return results, errors.New("could not get category assessments for student in phase")
		}
	}

	completion := db.AssessmentCompletion{}
	exists, err := s.assessmentCompletion.CheckAssessmentCompletionExists(ctx, courseParticipationID, coursePhaseID)
	if err != nil {
		log.Error("could not check assessment completion existence: ", err)
		return results, errors.New("could not check assessment completion existence")
	}
	if exists {
		completion, err = s.assessmentCompletion.GetAssessmentCompletion(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not get assessment completion: ", err)
			return results, errors.New("could not get assessment completion")
		}
	}

	var studentScore *scoreLevelDTO.StudentScore
	if config.GradingSheetVisible && len(assessments) > 0 {
		score, err := s.scoreLevel.GetStudentScore(ctx, courseParticipationID, coursePhaseID)
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
		actionItems, err = s.actionItem.ListActionItemsForStudentInPhase(ctx, courseParticipationID, coursePhaseID)
		if err != nil {
			log.Error("could not list action items for student in phase: ", err)
			return results, errors.New("could not list action items for student in phase")
		}
	}

	peerEvalResults := []assessmentDTO.AggregatedEvaluationResult{}
	selfEvalResults := []assessmentDTO.AggregatedEvaluationResult{}
	if config.GradingSheetVisible {
		peerEvalResults = evaluations.AggregateEvaluations(evals, assessmentType.Peer, evaluations.MinPeerRaters)
		// A self-evaluation only ever has its own author, so it takes no anonymity floor
		selfEvalResults = evaluations.AggregateEvaluations(evals, assessmentType.Self, 1)
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

func (s *AssessmentService) DeleteAssessment(ctx context.Context, id, coursePhaseID uuid.UUID) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := s.queries.WithTx(tx)

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

	err = s.assessmentCompletion.CheckAssessmentIsEditable(ctx, qtx, assessment.CourseParticipationID, assessment.CoursePhaseID)
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
