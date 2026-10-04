package actionItemDTO

import "github.com/google/uuid"

type CreateActionItemRequest struct {
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	Action                string    `json:"action"`
	Author                string    `json:"author"`
}
