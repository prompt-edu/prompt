package feedbackItemDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type FeedbackItem struct {
	ID                          uuid.UUID                     `json:"id"`
	FeedbackType                db.FeedbackType               `json:"feedbackType"`
	FeedbackText                string                        `json:"feedbackText"`
	CourseParticipationID       uuid.UUID                     `json:"courseParticipationID"`
	CoursePhaseID               uuid.UUID                     `json:"coursePhaseID"`
	AuthorCourseParticipationID uuid.UUID                     `json:"authorCourseParticipationID"`
	CreatedAt                   pgtype.Timestamptz            `json:"createdAt"`
	Type                        assessmentType.AssessmentType `json:"type"`
}

// GetFeedbackItemDTOsFromDBModels converts a slice of db.FeedbackItem to DTOs.
func GetFeedbackItemDTOsFromDBModels(dbFeedbackItems []db.FeedbackItem) []FeedbackItem {
	feedbackItems := make([]FeedbackItem, 0, len(dbFeedbackItems))
	for _, f := range dbFeedbackItems {
		feedbackItems = append(feedbackItems, MapDBFeedbackItemToFeedbackItemDTO(f))
	}
	return feedbackItems
}

// MapDBFeedbackItemToFeedbackItemDTO converts a db.FeedbackItem to DTO.
func MapDBFeedbackItemToFeedbackItemDTO(dbFeedbackItem db.FeedbackItem) FeedbackItem {
	return FeedbackItem{
		ID:                          dbFeedbackItem.ID,
		FeedbackType:                dbFeedbackItem.FeedbackType,
		FeedbackText:                dbFeedbackItem.FeedbackText,
		CourseParticipationID:       dbFeedbackItem.CourseParticipationID,
		CoursePhaseID:               dbFeedbackItem.CoursePhaseID,
		AuthorCourseParticipationID: dbFeedbackItem.AuthorCourseParticipationID,
		CreatedAt:                   dbFeedbackItem.CreatedAt,
		Type:                        assessmentType.MapDBAssessmentTypeToDTO(dbFeedbackItem.Type),
	}
}
