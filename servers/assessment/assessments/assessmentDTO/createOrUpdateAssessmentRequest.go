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
	AssessedAt            *time.Time               `json:"assessedAt,omitempty"`
	Author                string                   `json:"author"`
	AuthorID              string                   `json:"authorID"`
}
