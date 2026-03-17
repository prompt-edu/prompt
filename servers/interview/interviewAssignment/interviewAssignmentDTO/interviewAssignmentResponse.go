package interviewAssignmentDTO

import (
	"time"

	"github.com/google/uuid"
	interviewSlotDTO "github.com/prompt-edu/prompt/servers/interview/interviewSlot/interviewSlotDTO"
)

type InterviewAssignmentResponse struct {
	ID                    uuid.UUID                               `json:"id"`
	InterviewSlotID       uuid.UUID                               `json:"interview_slot_id"`
	CourseParticipationID uuid.UUID                               `json:"course_participation_id"`
	AssignedAt            time.Time                               `json:"assigned_at"`
	SlotDetails           *interviewSlotDTO.InterviewSlotResponse `json:"slot_details,omitempty"`
}
