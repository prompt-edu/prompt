package interviewSlotDTO

import (
	"time"

	"github.com/google/uuid"
)

type AssignmentInfo struct {
	ID                    uuid.UUID    `json:"id"`
	CourseParticipationID uuid.UUID    `json:"courseParticipationId"`
	AssignedAt            time.Time    `json:"assignedAt"`
	Student               *StudentInfo `json:"student,omitempty"`
}
