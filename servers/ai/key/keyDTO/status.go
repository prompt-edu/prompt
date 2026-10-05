package keyDTO

import (
	"time"

	db "github.com/prompt-edu/prompt/servers/ai/db/sqlc"
)

// Status describes a phase's key without ever containing it.
type Status struct {
	Configured bool       `json:"configured"`
	Last4      string     `json:"last4,omitempty"`
	SetBy      string     `json:"setBy,omitempty"`
	SetAt      *time.Time `json:"setAt,omitempty"`
}

func GetStatusDTOFromDBModel(row db.AiPhaseKey) Status {
	setAt := row.SetAt.Time
	return Status{Configured: true, Last4: row.Last4, SetBy: row.SetBy, SetAt: &setAt}
}
