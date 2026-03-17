package coursePhaseConfigDTO

import (
	"time"

	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
)

type SendEvaluationReminderRequest struct {
	EvaluationType assessmentType.AssessmentType `json:"evaluationType"`
}

type EvaluationReminderSendReport struct {
	SuccessfulEmails    []string                      `json:"successfulEmails"`
	FailedEmails        []string                      `json:"failedEmails"`
	RequestedRecipients int                           `json:"requestedRecipients"`
	EvaluationType      assessmentType.AssessmentType `json:"evaluationType"`
	Deadline            *time.Time                    `json:"deadline"`
	DeadlinePassed      bool                          `json:"deadlinePassed"`
	SentAt              time.Time                     `json:"sentAt"`
	PreviousSentAt      *time.Time                    `json:"previousSentAt"`
}
