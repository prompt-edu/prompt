package interviewAssignmentDTO

import "github.com/google/uuid"

type CreateInterviewAssignmentRequest struct {
	InterviewSlotID uuid.UUID `json:"interview_slot_id" binding:"required"`
}
