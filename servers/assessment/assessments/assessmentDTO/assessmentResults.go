package assessmentDTO

import (
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/actionItem/actionItemDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion/assessmentCompletionDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
)

// AggregatedEvaluationResult represents anonymized, averaged evaluation results for a competency.
type AggregatedEvaluationResult struct {
	CompetencyID        uuid.UUID `json:"competencyID"`
	AverageScoreNumeric float64   `json:"averageScoreNumeric"`
}

// StudentAssessmentResults bundles all information a student may see once results are released.
type StudentAssessmentResults struct {
	CourseParticipationID uuid.UUID                                    `json:"courseParticipationID"`
	CoursePhaseID         uuid.UUID                                    `json:"coursePhaseID"`
	AssessmentCompletion  assessmentCompletionDTO.AssessmentCompletion `json:"assessmentCompletion"`
	Assessments           []Assessment                                 `json:"assessments"`
	StudentScore          scoreLevelDTO.StudentScore                   `json:"studentScore"`
	PeerEvaluationResults []AggregatedEvaluationResult                 `json:"peerEvaluationResults"`
	SelfEvaluationResults []AggregatedEvaluationResult                 `json:"selfEvaluationResults"`
	ActionItems           []actionItemDTO.ActionItem                   `json:"actionItems,omitempty"`
}
