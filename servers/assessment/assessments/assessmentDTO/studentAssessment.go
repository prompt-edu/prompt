package assessmentDTO

import (
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion/assessmentCompletionDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationDTO"
)

type StudentAssessment struct {
	CourseParticipationID uuid.UUID                                    `json:"courseParticipationID"`
	Assessments           []Assessment                                 `json:"assessments"`
	AssessmentCompletion  assessmentCompletionDTO.AssessmentCompletion `json:"assessmentCompletion"`
	StudentScore          scoreLevelDTO.StudentScore                   `json:"studentScore"`
	Evaluations           []evaluationDTO.Evaluation                   `json:"evaluations"`
}
