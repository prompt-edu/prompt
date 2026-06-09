package assessmentDTO

import (
	"time"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
)

// CreateOrUpdateAssessmentRequest is the JSON payload for the per-competency
// assessment upsert endpoint. Author and AuthorID are populated server-side
// from the JWT and MUST NOT be set by the client (json:"-").
type CreateOrUpdateAssessmentRequest struct {
	CourseParticipationID uuid.UUID                `json:"courseParticipationID"`
	CoursePhaseID         uuid.UUID                `json:"coursePhaseID"`
	CompetencyID          uuid.UUID                `json:"competencyID"`
	ScoreLevel            scoreLevelDTO.ScoreLevel `json:"scoreLevel"`
	AssessedAt            *time.Time               `json:"assessedAt,omitempty"`

	Author   string `json:"-"`
	AuthorID string `json:"-"`
}
