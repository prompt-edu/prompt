package interviewSlotDTO

// MaxBatchInterviewSlots bounds a single batch so one request cannot create an unbounded schedule.
const MaxBatchInterviewSlots = 100

type CreateInterviewSlotsBatchRequest struct {
	Slots []CreateInterviewSlotRequest `json:"slots" binding:"required,min=1,max=100,dive"`
}
