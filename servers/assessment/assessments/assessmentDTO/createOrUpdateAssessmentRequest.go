package assessmentDTO

import (
	"time"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
)

type CreateOrUpdateAssessmentRequest struct {
	CourseParticipationID uuid.UUID                `json:"courseParticipationID"`
	CoursePhaseID         uuid.UUID                `json:"coursePhaseID"`
	CompetencyID          uuid.UUID                `json:"competencyID"`
	ScoreLevel            scoreLevelDTO.ScoreLevel `json:"scoreLevel"`
	Comment               string                   `json:"comment"`
	Examples              string                   `json:"examples"`
	AssessedAt            *time.Time               `json:"assessedAt,omitempty"`
	Author                string                   `json:"author"`
}
