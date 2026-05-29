package mailingDTO

import (
	"time"

	"github.com/google/uuid"
)

type SendManualMailRequest struct {
	Subject                         string            `json:"subject"`
	Content                         string            `json:"content"`
	RecipientCourseParticipationIDs []uuid.UUID       `json:"recipientCourseParticipationIDs"`
	AdditionalPlaceholders          map[string]string `json:"additionalPlaceholders"`
}

type ManualMailReport struct {
	SuccessfulEmails    []string  `json:"successfulEmails"`
	FailedEmails        []string  `json:"failedEmails"`
	RequestedRecipients int       `json:"requestedRecipients"`
	SentAt              time.Time `json:"sentAt"`
}
