package interviewAssignmentDTO

import "github.com/google/uuid"

type CreateInterviewAssignmentAdminRequest struct {
	InterviewSlotID       uuid.UUID `json:"interview_slot_id" binding:"required"`
	CourseParticipationID uuid.UUID `json:"course_participation_id" binding:"required"`
}
