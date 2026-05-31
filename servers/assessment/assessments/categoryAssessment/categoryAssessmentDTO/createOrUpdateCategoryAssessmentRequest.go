package categoryAssessmentDTO

import (
	"github.com/google/uuid"
)

// CreateOrUpdateCategoryAssessmentRequest is the JSON payload for the upsert
// endpoint. Author and AuthorID are populated server-side from the JWT and
// MUST NOT be set by the client (they are intentionally not exported in JSON).
type CreateOrUpdateCategoryAssessmentRequest struct {
	CategoryID            uuid.UUID `json:"categoryID"`
	CoursePhaseID         uuid.UUID `json:"coursePhaseID"`
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	Comment               string    `json:"comment"`

	Author   string `json:"-"`
	AuthorID string `json:"-"`
}
