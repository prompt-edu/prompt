package calls

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/ai/db/sqlc"
)

const (
	EventShown         = "shown"
	EventAccepted      = "accepted"
	EventEdited        = "edited"
	EventRejected      = "rejected"
	EventContentViewed = "content_viewed"

	ContentAvailable   = "available"
	ContentRestricted  = "restricted"
	ContentUnavailable = "unavailable"
)

// Call is the audit metadata of one model call.
type Call struct {
	ID               uuid.UUID       `json:"id"`
	ActorID          string          `json:"actorId"`
	ActorRole        string          `json:"actorRole"`
	Feature          string          `json:"feature"`
	Template         *string         `json:"template"`
	TemplateVersion  *string         `json:"templateVersion"`
	RequestedModel   *string         `json:"requestedModel"`
	ServedModel      *string         `json:"servedModel"`
	Provider         string          `json:"provider"`
	Params           json.RawMessage `json:"params"`
	ContextHash      *string         `json:"contextHash"`
	ResponseHash     *string         `json:"responseHash"`
	Outcome          string          `json:"outcome"`
	HTTPStatus       *int32          `json:"httpStatus"`
	FinishReason     *string         `json:"finishReason"`
	ErrorCode        *string         `json:"errorCode"`
	PromptTokens     *int32          `json:"promptTokens"`
	CompletionTokens *int32          `json:"completionTokens"`
	Streamed         bool            `json:"streamed"`
	ServerVersion    string          `json:"serverVersion"`
	RequestedAt      time.Time       `json:"requestedAt"`
	FirstTokenAt     *time.Time      `json:"firstTokenAt"`
	CompletedAt      *time.Time      `json:"completedAt"`
}

type Content struct {
	Request      json.RawMessage `json:"request"`
	ResponseText string          `json:"responseText"`
}

type Detail struct {
	Call
	ContentState string      `json:"contentState"`
	Content      *Content    `json:"content,omitempty"`
	Subjects     []uuid.UUID `json:"subjects"`
	Events       []Event     `json:"events"`
}

type Event struct {
	ID        uuid.UUID       `json:"id"`
	ActorID   string          `json:"actorId"`
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	CreatedAt time.Time       `json:"createdAt"`
}

type EventRequest struct {
	Type         string     `json:"type" binding:"required,oneof=shown accepted edited rejected"`
	EditDistance *int       `json:"editDistance" binding:"omitempty,min=0"`
	ActionItemID *uuid.UUID `json:"actionItemId"`
}

func (r EventRequest) data() any {
	return struct {
		EditDistance *int       `json:"editDistance,omitempty"`
		ActionItemID *uuid.UUID `json:"actionItemId,omitempty"`
	}{r.EditDistance, r.ActionItemID}
}

func callOf(row db.AiCall) Call {
	return Call{
		ID:               row.ID,
		ActorID:          row.ActorID,
		ActorRole:        row.ActorRole,
		Feature:          row.Feature,
		Template:         textOf(row.Template),
		TemplateVersion:  textOf(row.TemplateVersion),
		RequestedModel:   textOf(row.RequestedModel),
		ServedModel:      textOf(row.ServedModel),
		Provider:         row.Provider,
		Params:           row.Params,
		ContextHash:      textOf(row.ContextHash),
		ResponseHash:     textOf(row.ResponseHash),
		Outcome:          row.Outcome,
		HTTPStatus:       intOf(row.HttpStatus),
		FinishReason:     textOf(row.FinishReason),
		ErrorCode:        textOf(row.ErrorCode),
		PromptTokens:     intOf(row.PromptTokens),
		CompletionTokens: intOf(row.CompletionTokens),
		Streamed:         row.Streamed,
		ServerVersion:    row.ServerVersion,
		RequestedAt:      row.RequestedAt.Time,
		FirstTokenAt:     timeOf(row.FirstTokenAt),
		CompletedAt:      timeOf(row.CompletedAt),
	}
}

func eventOf(row db.AiCallEvent) Event {
	return Event{ID: row.ID, ActorID: row.ActorID, Type: row.Type, Data: row.Data, CreatedAt: row.CreatedAt.Time}
}

func textOf(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func intOf(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}
	return &value.Int32
}

func timeOf(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}
