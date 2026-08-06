package evaluations

import (
	"testing"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationDTO"
	"github.com/stretchr/testify/assert"
)

func peerEvaluation(competencyID, authorID uuid.UUID, level scoreLevelDTO.ScoreLevel) evaluationDTO.Evaluation {
	return evaluationDTO.Evaluation{
		ID:                          uuid.New(),
		CompetencyID:                competencyID,
		ScoreLevel:                  level,
		AuthorCourseParticipationID: authorID,
		Type:                        assessmentType.Peer,
	}
}

func TestAggregateEvaluationsAveragesMatchingType(t *testing.T) {
	competencyID := uuid.New()
	evals := []evaluationDTO.Evaluation{
		peerEvaluation(competencyID, uuid.New(), scoreLevelDTO.ScoreLevelVeryGood), // 1
		peerEvaluation(competencyID, uuid.New(), scoreLevelDTO.ScoreLevelOk),       // 3
		{
			ID:                          uuid.New(),
			CompetencyID:                competencyID,
			ScoreLevel:                  scoreLevelDTO.ScoreLevelBad,
			AuthorCourseParticipationID: uuid.New(),
			Type:                        assessmentType.Self,
		},
	}

	results := AggregateEvaluations(evals, assessmentType.Peer, 2)

	assert.Len(t, results, 1)
	assert.Equal(t, competencyID, results[0].CompetencyID)
	assert.InDelta(t, 2.0, results[0].AverageScoreNumeric, 0.0001, "self evaluation must not be averaged in")
}

func TestAggregateEvaluationsSuppressesSingleRater(t *testing.T) {
	loneCompetency := uuid.New()
	sharedCompetency := uuid.New()
	firstAuthor := uuid.New()
	secondAuthor := uuid.New()

	evals := []evaluationDTO.Evaluation{
		peerEvaluation(loneCompetency, firstAuthor, scoreLevelDTO.ScoreLevelVeryGood),
		peerEvaluation(sharedCompetency, firstAuthor, scoreLevelDTO.ScoreLevelGood),
		peerEvaluation(sharedCompetency, secondAuthor, scoreLevelDTO.ScoreLevelOk),
	}

	results := AggregateEvaluations(evals, assessmentType.Peer, 2)

	assert.Len(t, results, 1, "a competency rated by a single peer must be omitted")
	assert.Equal(t, sharedCompetency, results[0].CompetencyID)
}

func TestAggregateEvaluationsCountsDistinctAuthors(t *testing.T) {
	competencyID := uuid.New()
	author := uuid.New()

	// Duplicate rows from one author must not pass the floor
	evals := []evaluationDTO.Evaluation{
		peerEvaluation(competencyID, author, scoreLevelDTO.ScoreLevelGood),
		peerEvaluation(competencyID, author, scoreLevelDTO.ScoreLevelOk),
	}

	assert.Empty(t, AggregateEvaluations(evals, assessmentType.Peer, 2))
	assert.Len(t, AggregateEvaluations(evals, assessmentType.Peer, 1), 1,
		"the assessment path keeps its unfiltered behavior")
}
