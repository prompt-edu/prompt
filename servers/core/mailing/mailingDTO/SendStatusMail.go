package mailingDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type SendStatusMail struct {
	StatusMailToBeSend db.PassStatus `json:"statusMailToBeSend"`
	// Optional: when omitted, every participant with the given pass status is mailed. When set, the
	// mail only goes to those of these participants that carry the given pass status.
	RecipientCourseParticipationIDs []uuid.UUID `json:"recipientCourseParticipationIDs,omitempty"`
}
