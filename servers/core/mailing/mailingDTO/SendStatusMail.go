package mailingDTO

import db "github.com/prompt-edu/prompt/servers/core/db/sqlc"

type SendStatusMail struct {
	StatusMailToBeSend db.PassStatus `json:"statusMailToBeSend"`
}
