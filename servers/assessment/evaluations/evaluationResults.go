package evaluations

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig/coursePhaseConfigDTO"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationDTO"
)

// MinPeerRaters keeps a peer average from revealing a single named teammate's exact score. Every
// student-facing peer average takes this floor, on evaluation-only and assessment-enabled phases alike.
const MinPeerRaters = 2

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
