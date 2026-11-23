package callDTO

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/ai/db/sqlc"
)

const (
	// EventContentViewed is recorded by this service when an admin opens a call's content. The
	// events a phase reports are in prompt-sdk's ai package.
	EventContentViewed = "content_viewed"
)

type Event struct {
	ID        uuid.UUID       `json:"id"`
	ActorID   string          `json:"actorId"`
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	CreatedAt time.Time       `json:"createdAt"`
}

func GetEventDTOFromDBModel(row db.AiCallEvent) Event {
	return Event{ID: row.ID, ActorID: row.ActorID, Type: row.Type, Data: row.Data, CreatedAt: row.CreatedAt.Time}
}
