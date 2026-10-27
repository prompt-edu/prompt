package actionItemDTO

import "github.com/google/uuid"

// CreateActionItemRequest is the JSON payload for creating an action item.
// Author is populated server-side from the JWT and MUST NOT be set by the
// client (it is intentionally not exported in JSON).
type CreateActionItemRequest struct {
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	Action                string    `json:"action"`

	Author string `json:"-"`
}
