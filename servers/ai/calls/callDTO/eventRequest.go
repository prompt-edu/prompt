package callDTO

import "github.com/google/uuid"

type EventRequest struct {
	Type         string     `json:"type" binding:"required,oneof=shown accepted edited rejected"`
	EditDistance *int       `json:"editDistance" binding:"omitempty,min=0"`
	TargetID     *uuid.UUID `json:"targetId"`
}

// Data is what the event stores next to its type. TargetID names what the phase applied the output to.
func (r EventRequest) Data() any {
	return struct {
		EditDistance *int       `json:"editDistance,omitempty"`
		TargetID     *uuid.UUID `json:"targetId,omitempty"`
	}{r.EditDistance, r.TargetID}
}
