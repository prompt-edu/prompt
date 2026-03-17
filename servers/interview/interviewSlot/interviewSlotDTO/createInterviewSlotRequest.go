package interviewSlotDTO

import "time"

type CreateInterviewSlotRequest struct {
	StartTime time.Time `json:"startTime" binding:"required"`
	EndTime   time.Time `json:"endTime" binding:"required"`
	Location  *string   `json:"location"`
	Capacity  int32     `json:"capacity" binding:"required,min=1"`
}
