package evaluationDTO

import (
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
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
