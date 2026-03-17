package interviewSlotDTO

import "github.com/google/uuid"

type SelfParticipationResponse struct {
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
}
