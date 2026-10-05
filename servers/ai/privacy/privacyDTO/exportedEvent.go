package privacyDTO

import (
	"encoding/json"
	"time"
)

type ExportedEvent struct {
	Type          string          `json:"type"`
	Data          json.RawMessage `json:"data"`
	MadeBySubject bool            `json:"madeBySubject"`
	CreatedAt     time.Time       `json:"createdAt"`
}
