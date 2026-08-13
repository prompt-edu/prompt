package auditLogDTO

import (
	"encoding/json"
	"time"
)

// AuditEntry is the JSON representation of one audit log row returned to the UI.
type AuditEntry struct {
	ID            string          `json:"id"`
	CreatedAt     time.Time       `json:"createdAt"`
	ActorID       string          `json:"actorID,omitempty"`
	ActorName     string          `json:"actorName"`
	ActorEmail    string          `json:"actorEmail"`
	ActorRoles    []string        `json:"actorRoles"`
	ActorRole     string          `json:"actorRole"`
	Action        string          `json:"action"`
	ActionKey     string          `json:"actionKey"`
	Outcome       string          `json:"outcome"`
	EntityType    string          `json:"entityType,omitempty"`
	EntityID      string          `json:"entityID,omitempty"`
	EntityName    string          `json:"entityName,omitempty"`
	CourseID      string          `json:"courseID,omitempty"`
	CoursePhaseID string          `json:"coursePhaseID,omitempty"`
	SourceService string          `json:"sourceService"`
	HTTPMethod    string          `json:"httpMethod,omitempty"`
	HTTPPath      string          `json:"httpPath,omitempty"`
	HTTPStatus    int             `json:"httpStatus,omitempty"`
	Metadata      json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
}

// Cursor is the keyset position for fetching the next page (older entries).
type Cursor struct {
	CreatedAt time.Time `json:"createdAt"`
	ID        string    `json:"id"`
}

// AuditLogPage is one page of audit entries plus the cursor for the next page
// (nil when there are no more).
type AuditLogPage struct {
	Entries    []AuditEntry `json:"entries"`
	NextCursor *Cursor      `json:"nextCursor"`
}

// ListFilters holds the server-side query filters parsed from the request.
type ListFilters struct {
	CourseID        string // empty => all courses (admin view)
	ActorRole       string
	Outcome         string
	ActionKey       string
	EntityType      string
	SourceService   string
	CoursePhaseID   string
	Search          string
	From            *time.Time
	To              *time.Time
	CursorCreatedAt *time.Time
	CursorID        string
	Limit           int
}
