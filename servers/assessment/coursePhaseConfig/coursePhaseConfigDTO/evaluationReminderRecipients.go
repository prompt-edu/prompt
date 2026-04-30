package coursePhaseConfigDTO

import (
	"time"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
)

type EvaluationReminderRecipients struct {
	EvaluationType                         assessmentType.AssessmentType `json:"evaluationType"`
	EvaluationTypeLabel                    string                        `json:"evaluationTypeLabel"`
	EvaluationEnabled                      bool                          `json:"evaluationEnabled"`
	Deadline                               *time.Time                    `json:"deadline"`
	EvaluationDeadlinePlaceholder          string                        `json:"evaluationDeadlinePlaceholder"`
	DeadlinePassed                         bool                          `json:"deadlinePassed"`
	IncompleteAuthorCourseParticipationIDs []uuid.UUID                   `json:"incompleteAuthorCourseParticipationIDs"`
	TotalAuthors                           int                           `json:"totalAuthors"`
	CompletedAuthors                       int                           `json:"completedAuthors"`
}
