package callDTO

import "github.com/google/uuid"

type EventRequest struct {
	Type         string     `json:"type" binding:"required,oneof=shown accepted edited rejected"`
	EditDistance *int       `json:"editDistance" binding:"omitempty,min=0"`
	ActionItemID *uuid.UUID `json:"actionItemId"`
}

// Data is what the event stores next to its type.
func (r EventRequest) Data() any {
	return struct {
		EditDistance *int       `json:"editDistance,omitempty"`
		ActionItemID *uuid.UUID `json:"actionItemId,omitempty"`
	}{r.EditDistance, r.ActionItemID}
}
