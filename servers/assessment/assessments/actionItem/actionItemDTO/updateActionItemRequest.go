package actionItemDTO

import "github.com/google/uuid"

// UpdateActionItemRequest is the JSON payload for updating an action item.
// Author is populated server-side from the JWT and MUST NOT be set by the
// client (it is intentionally not exported in JSON).
type UpdateActionItemRequest struct {
	ID                    uuid.UUID `json:"id"`
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	Action                string    `json:"action"`

	Author string `json:"-"`
}
