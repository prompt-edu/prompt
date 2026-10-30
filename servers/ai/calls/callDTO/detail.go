package callDTO

import "github.com/google/uuid"

const (
	ContentAvailable   = "available"
	ContentRestricted  = "restricted"
	ContentUnavailable = "unavailable"
)

type Detail struct {
	Call
	ContentState string      `json:"contentState"`
	Content      *Content    `json:"content,omitempty"`
	Subjects     []uuid.UUID `json:"subjects"`
	Events       []Event     `json:"events"`
}
