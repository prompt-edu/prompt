package assessmentCompletionDTO

import "github.com/google/uuid"

type BatchCompletionRequest struct {
	CourseParticipationIDs []uuid.UUID `json:"courseParticipationIDs" binding:"required,min=1"`
}

type SkipReason string

const (
	SkipReasonNoCompletion         SkipReason = "no_completion"
	SkipReasonRemainingAssessments SkipReason = "remaining_assessments"
	SkipReasonAlreadyCompleted     SkipReason = "already_completed"
	SkipReasonNotCompleted         SkipReason = "not_completed"
)

type SkippedCompletion struct {
	CourseParticipationID uuid.UUID  `json:"courseParticipationID"`
	Reason                SkipReason `json:"reason"`
}

type BatchMarkResult struct {
	Marked  []uuid.UUID         `json:"marked"`
	Skipped []SkippedCompletion `json:"skipped"`
}

type BatchUnmarkResult struct {
	Unmarked []uuid.UUID         `json:"unmarked"`
	Skipped  []SkippedCompletion `json:"skipped"`
}
