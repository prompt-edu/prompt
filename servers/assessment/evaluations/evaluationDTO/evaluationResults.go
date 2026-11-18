package evaluationDTO

import (
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

// AggregatedEvaluationResult represents anonymized, averaged evaluation results for a competency.
type AggregatedEvaluationResult struct {
	CompetencyID        uuid.UUID `json:"competencyID"`
	AverageScoreNumeric float64   `json:"averageScoreNumeric"`
}

// SelfEvaluationResult is a student's own submitted score for a competency.
type SelfEvaluationResult struct {
	CompetencyID uuid.UUID                `json:"competencyID"`
	ScoreLevel   scoreLevelDTO.ScoreLevel `json:"scoreLevel"`
}

// StudentEvaluationResults is what a student may see on an evaluation-only phase once results are
// released. Peer scores are averaged and carry no author or rater count.
type StudentEvaluationResults struct {
	CourseParticipationID uuid.UUID                    `json:"courseParticipationID"`
	CoursePhaseID         uuid.UUID                    `json:"coursePhaseID"`
	SelfResults           []SelfEvaluationResult       `json:"selfResults"`
	PeerResults           []AggregatedEvaluationResult `json:"peerResults"`
}

// AnonymousFeedbackItem is a feedback text without anything that points to its author.
type AnonymousFeedbackItem struct {
	FeedbackType db.FeedbackType `json:"feedbackType"`
	FeedbackText string          `json:"feedbackText"`
}

// TutorEvaluationResults is what a tutor may see about the tutor evaluations written about them.
type TutorEvaluationResults struct {
	TutorID        uuid.UUID                    `json:"tutorID"`
	CoursePhaseID  uuid.UUID                    `json:"coursePhaseID"`
	EvaluatorCount int                          `json:"evaluatorCount"`
	Results        []AggregatedEvaluationResult `json:"results"`
	FeedbackItems  []AnonymousFeedbackItem      `json:"feedbackItems"`
}
