package evaluations

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig/coursePhaseConfigDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationDTO"
)

// MinPeerRaters keeps a peer average from revealing a single named teammate's exact score. Every
// student-facing peer average takes this floor, on evaluation-only and assessment-enabled phases alike.
const MinPeerRaters = 2

// MinTutorRaters does the same for the tutor evaluation results a tutor sees about themselves: the
// tutor knows their team, so a single rater would be identifiable.
const MinTutorRaters = 2

// AggregateEvaluations averages the scores of one evaluation type per competency. A competency is
// only reported when at least minDistinctAuthors different people rated it.
func AggregateEvaluations(evals []evaluationDTO.Evaluation, targetType assessmentType.AssessmentType, minDistinctAuthors int) []evaluationDTO.AggregatedEvaluationResult {
	type accumulator struct {
		sum     float64
		count   int
		authors map[uuid.UUID]struct{}
	}

	aggregated := make(map[uuid.UUID]*accumulator)
	for _, eval := range evals {
		if eval.Type != targetType {
			continue
		}
		current, exists := aggregated[eval.CompetencyID]
		if !exists {
			current = &accumulator{authors: make(map[uuid.UUID]struct{})}
			aggregated[eval.CompetencyID] = current
		}
		current.sum += scoreLevelDTO.MapScoreLevelToNumber(eval.ScoreLevel)
		current.count++
		current.authors[eval.AuthorCourseParticipationID] = struct{}{}
	}

	results := make([]evaluationDTO.AggregatedEvaluationResult, 0, len(aggregated))
	for competencyID, acc := range aggregated {
		if acc.count == 0 || len(acc.authors) < minDistinctAuthors {
			continue
		}
		results = append(results, evaluationDTO.AggregatedEvaluationResult{
			CompetencyID:        competencyID,
			AverageScoreNumeric: acc.sum / float64(acc.count),
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].CompetencyID.String() < results[j].CompetencyID.String()
	})

	return results
}

// GetStudentEvaluationResults bundles what a student may see on an evaluation-only phase: their own
// submitted self-evaluation plus anonymized peer averages.
func (s *EvaluationService) GetStudentEvaluationResults(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID, config coursePhaseConfigDTO.CoursePhaseConfig) (evaluationDTO.StudentEvaluationResults, error) {
	results := evaluationDTO.StudentEvaluationResults{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
		SelfResults:           []evaluationDTO.SelfEvaluationResult{},
		PeerResults:           []evaluationDTO.AggregatedEvaluationResult{},
	}

	evals, err := s.GetEvaluationsForParticipantInPhase(ctx, courseParticipationID, coursePhaseID)
	if err != nil {
		return results, err
	}

	if config.SelfEvaluationEnabled {
		for _, eval := range evals {
			if eval.Type != assessmentType.Self {
				continue
			}
			results.SelfResults = append(results.SelfResults, evaluationDTO.SelfEvaluationResult{
				CompetencyID: eval.CompetencyID,
				ScoreLevel:   eval.ScoreLevel,
			})
		}
		sort.Slice(results.SelfResults, func(i, j int) bool {
			return results.SelfResults[i].CompetencyID.String() < results.SelfResults[j].CompetencyID.String()
		})
	}

	if config.PeerEvaluationEnabled {
		results.PeerResults = AggregateEvaluations(evals, assessmentType.Peer, MinPeerRaters)
	}

	return results, nil
}

// BuildTutorEvaluationResults anonymizes the tutor evaluations written about one tutor: averages
// per competency rated by at least MinTutorRaters students, and the feedback texts of a type
// (positive or negative) only once at least MinTutorRaters students wrote that type. The floor
// applies per type because the tutor sees the types apart, so a type written by one student only
// would point to them. The texts are sorted by type and text, so their order does not hint at who
// wrote them or when.
func BuildTutorEvaluationResults(coursePhaseID, tutorID uuid.UUID, evals []evaluationDTO.Evaluation, feedbackItems []db.FeedbackItem) evaluationDTO.TutorEvaluationResults {
	evaluators := make(map[uuid.UUID]struct{})
	for _, eval := range evals {
		if eval.Type == assessmentType.Tutor {
			evaluators[eval.AuthorCourseParticipationID] = struct{}{}
		}
	}

	results := evaluationDTO.TutorEvaluationResults{
		TutorID:        tutorID,
		CoursePhaseID:  coursePhaseID,
		EvaluatorCount: len(evaluators),
		Results:        AggregateEvaluations(evals, assessmentType.Tutor, MinTutorRaters),
		FeedbackItems:  []evaluationDTO.AnonymousFeedbackItem{},
	}

	authorsByType := make(map[db.FeedbackType]map[uuid.UUID]struct{})
	for _, item := range feedbackItems {
		if authorsByType[item.FeedbackType] == nil {
			authorsByType[item.FeedbackType] = make(map[uuid.UUID]struct{})
		}
		authorsByType[item.FeedbackType][item.AuthorCourseParticipationID] = struct{}{}
	}

	for _, item := range feedbackItems {
		if len(authorsByType[item.FeedbackType]) < MinTutorRaters {
			continue
		}
		results.FeedbackItems = append(results.FeedbackItems, evaluationDTO.AnonymousFeedbackItem{
			FeedbackType: item.FeedbackType,
			FeedbackText: item.FeedbackText,
		})
	}
	sort.Slice(results.FeedbackItems, func(i, j int) bool {
		a, b := results.FeedbackItems[i], results.FeedbackItems[j]
		if a.FeedbackType != b.FeedbackType {
			return a.FeedbackType < b.FeedbackType
		}
		return a.FeedbackText < b.FeedbackText
	})

	return results
}

// GetTutorEvaluationResults loads the tutor evaluations and tutor feedback written about one
// tutor and anonymizes them with BuildTutorEvaluationResults.
func (s *EvaluationService) GetTutorEvaluationResults(ctx context.Context, coursePhaseID, tutorID uuid.UUID) (evaluationDTO.TutorEvaluationResults, error) {
	evals, err := s.GetEvaluationsForTutorInPhase(ctx, tutorID, coursePhaseID)
	if err != nil {
		return evaluationDTO.TutorEvaluationResults{}, err
	}

	feedbackItems, err := s.queries.ListFeedbackItemsForTutorInPhase(ctx, db.ListFeedbackItemsForTutorInPhaseParams{
		CourseParticipationID: tutorID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		return evaluationDTO.TutorEvaluationResults{}, fmt.Errorf("could not get feedback items for tutor in phase: %w", err)
	}

	return BuildTutorEvaluationResults(coursePhaseID, tutorID, evals, feedbackItems), nil
}
