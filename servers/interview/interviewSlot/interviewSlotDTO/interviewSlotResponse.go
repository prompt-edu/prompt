package interviewSlotDTO

import (
	"time"

	"github.com/google/uuid"
)

type InterviewSlotResponse struct {
	ID            uuid.UUID        `json:"id"`
	CoursePhaseID uuid.UUID        `json:"coursePhaseId"`
	StartTime     time.Time        `json:"startTime"`
	EndTime       time.Time        `json:"endTime"`
	Location      *string          `json:"location"`
	Capacity      int32            `json:"capacity"`
	AssignedCount int64            `json:"assignedCount"`
	Assignments   []AssignmentInfo `json:"assignments"`
	CreatedAt     time.Time        `json:"createdAt"`
	UpdatedAt     time.Time        `json:"updatedAt"`
}
