package mailingDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type SendStatusMail struct {
	StatusMailToBeSend db.PassStatus `json:"statusMailToBeSend"`
	// Optional: when set, the status mail is only sent to these participants instead of all
	// participants with the given pass status.
	RecipientCourseParticipationIDs []uuid.UUID `json:"recipientCourseParticipationIDs"`
}
