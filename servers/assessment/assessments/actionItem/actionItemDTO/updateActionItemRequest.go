package actionItemDTO

import "github.com/google/uuid"

type UpdateActionItemRequest struct {
	ID                    uuid.UUID `json:"id"`
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	Action                string    `json:"action"`
	Author                string    `json:"author"`
}
