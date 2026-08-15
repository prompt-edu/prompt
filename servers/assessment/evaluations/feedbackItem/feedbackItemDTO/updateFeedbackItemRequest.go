package feedbackItemDTO

import (
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type UpdateFeedbackItemRequest struct {
	FeedbackType db.FeedbackType `json:"feedbackType" binding:"required,oneof='positive' 'negative'"`
	FeedbackText string          `json:"feedbackText" binding:"required"`
}
