package evaluations

import (
	"testing"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
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

func tutorEvaluation(competencyID, authorID uuid.UUID, level scoreLevelDTO.ScoreLevel) evaluationDTO.Evaluation {
	evaluation := peerEvaluation(competencyID, authorID, level)
	evaluation.Type = assessmentType.Tutor
	return evaluation
}

func tutorFeedback(authorID uuid.UUID, feedbackType db.FeedbackType, text string) db.FeedbackItem {
	return db.FeedbackItem{ID: uuid.New(), AuthorCourseParticipationID: authorID, FeedbackType: feedbackType, FeedbackText: text}
}

func TestBuildTutorEvaluationResultsAveragesAndCountsEvaluators(t *testing.T) {
	sharedCompetency := uuid.New()
	loneCompetency := uuid.New()
	firstAuthor := uuid.New()
	secondAuthor := uuid.New()

	results := BuildTutorEvaluationResults(uuid.New(), uuid.New(), []evaluationDTO.Evaluation{
		tutorEvaluation(sharedCompetency, firstAuthor, scoreLevelDTO.ScoreLevelVeryGood),
		tutorEvaluation(sharedCompetency, secondAuthor, scoreLevelDTO.ScoreLevelOk),
		tutorEvaluation(loneCompetency, firstAuthor, scoreLevelDTO.ScoreLevelBad),
	}, nil)

	assert.Equal(t, 2, results.EvaluatorCount)
	assert.Len(t, results.Results, 1, "a competency rated by a single student must be omitted")
	assert.Equal(t, sharedCompetency, results.Results[0].CompetencyID)
	assert.InDelta(t, 2.0, results.Results[0].AverageScoreNumeric, 0.0001)
	assert.NotNil(t, results.FeedbackItems)
	assert.Empty(t, results.FeedbackItems)
}

func TestBuildTutorEvaluationResultsHidesFeedbackOfASingleAuthor(t *testing.T) {
	author := uuid.New()

	results := BuildTutorEvaluationResults(uuid.New(), uuid.New(), nil, []db.FeedbackItem{
		tutorFeedback(author, db.FeedbackTypePositive, "Great sessions"),
		tutorFeedback(author, db.FeedbackTypeNegative, "Often late"),
	})

	assert.Empty(t, results.FeedbackItems, "feedback written by one student only would identify them")
}

func TestBuildTutorEvaluationResultsSortsFeedbackWithoutAuthors(t *testing.T) {
	firstAuthor := uuid.New()
	secondAuthor := uuid.New()

	results := BuildTutorEvaluationResults(uuid.New(), uuid.New(), nil, []db.FeedbackItem{
		tutorFeedback(firstAuthor, db.FeedbackTypePositive, "Patient"),
		tutorFeedback(secondAuthor, db.FeedbackTypeNegative, "Unclear deadlines"),
		tutorFeedback(firstAuthor, db.FeedbackTypeNegative, "Slow replies"),
		tutorFeedback(secondAuthor, db.FeedbackTypePositive, "Helpful reviews"),
	})

	assert.Equal(t, []evaluationDTO.AnonymousFeedbackItem{
		{FeedbackType: db.FeedbackTypeNegative, FeedbackText: "Slow replies"},
		{FeedbackType: db.FeedbackTypeNegative, FeedbackText: "Unclear deadlines"},
		{FeedbackType: db.FeedbackTypePositive, FeedbackText: "Helpful reviews"},
		{FeedbackType: db.FeedbackTypePositive, FeedbackText: "Patient"},
	}, results.FeedbackItems)
}

func TestBuildTutorEvaluationResultsHidesAFeedbackTypeOfASingleAuthor(t *testing.T) {
	firstAuthor := uuid.New()
	secondAuthor := uuid.New()

	// Two students wrote feedback, but only one of them wrote negative feedback.
	results := BuildTutorEvaluationResults(uuid.New(), uuid.New(), nil, []db.FeedbackItem{
		tutorFeedback(firstAuthor, db.FeedbackTypePositive, "Patient"),
		tutorFeedback(secondAuthor, db.FeedbackTypePositive, "Helpful reviews"),
		tutorFeedback(secondAuthor, db.FeedbackTypeNegative, "Unclear deadlines"),
	})

	assert.Equal(t, []evaluationDTO.AnonymousFeedbackItem{
		{FeedbackType: db.FeedbackTypePositive, FeedbackText: "Helpful reviews"},
		{FeedbackType: db.FeedbackTypePositive, FeedbackText: "Patient"},
	}, results.FeedbackItems, "a feedback type written by one student only would identify them")
}
