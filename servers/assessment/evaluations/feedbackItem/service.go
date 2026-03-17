package feedbackItem

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/feedbackItem/feedbackItemDTO"
	log "github.com/sirupsen/logrus"
)

type FeedbackItemService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var FeedbackItemServiceSingleton *FeedbackItemService

func GetFeedbackItem(ctx context.Context, feedbackItemID uuid.UUID) (feedbackItemDTO.FeedbackItem, error) {
	feedbackItem, err := FeedbackItemServiceSingleton.queries.GetFeedbackItem(ctx, feedbackItemID)
	if err != nil {
		log.Error("could not get feedback item: ", err)
		return feedbackItemDTO.FeedbackItem{}, errors.New("could not get feedback item")
	}
	return feedbackItemDTO.MapDBFeedbackItemToFeedbackItemDTO(feedbackItem), nil
}

func ListFeedbackItemsForCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]feedbackItemDTO.FeedbackItem, error) {
	feedbackItems, err := FeedbackItemServiceSingleton.queries.ListFeedbackItemsForCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not list feedback items for course phase: ", err)
		return nil, errors.New("could not list feedback items for course phase")
	}
	return feedbackItemDTO.GetFeedbackItemDTOsFromDBModels(feedbackItems), nil
}

func ListFeedbackItemsForParticipantInPhase(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) ([]feedbackItemDTO.FeedbackItem, error) {
	feedbackItems, err := FeedbackItemServiceSingleton.queries.ListFeedbackItemsForParticipantInPhase(ctx, db.ListFeedbackItemsForParticipantInPhaseParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not list feedback items for participant in phase: ", err)
		return nil, errors.New("could not list feedback items for participant in phase")
	}
	return feedbackItemDTO.GetFeedbackItemDTOsFromDBModels(feedbackItems), nil
}

func ListFeedbackItemsForTutorInPhase(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) ([]feedbackItemDTO.FeedbackItem, error) {
	feedbackItems, err := FeedbackItemServiceSingleton.queries.ListFeedbackItemsForTutorInPhase(ctx, db.ListFeedbackItemsForTutorInPhaseParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("could not list feedback items for tutor in phase: ", err)
		return nil, errors.New("could not list feedback items for tutor in phase")
	}
	return feedbackItemDTO.GetFeedbackItemDTOsFromDBModels(feedbackItems), nil
}

func ListFeedbackItemsByAuthorInPhase(ctx context.Context, authorCourseParticipationID, coursePhaseID uuid.UUID) ([]feedbackItemDTO.FeedbackItem, error) {
	feedbackItems, err := FeedbackItemServiceSingleton.queries.ListFeedbackItemsByAuthorInPhase(ctx, db.ListFeedbackItemsByAuthorInPhaseParams{
		AuthorCourseParticipationID: authorCourseParticipationID,
		CoursePhaseID:               coursePhaseID,
	})
	if err != nil {
		log.Error("could not list feedback items by author in phase: ", err)
		return nil, errors.New("could not list feedback items by author in phase")
	}
	return feedbackItemDTO.GetFeedbackItemDTOsFromDBModels(feedbackItems), nil
}

func CreateFeedbackItem(ctx context.Context, req feedbackItemDTO.CreateFeedbackItemRequest) error {
	err := FeedbackItemServiceSingleton.queries.CreateFeedbackItem(ctx, db.CreateFeedbackItemParams{
		ID:                          uuid.New(),
		FeedbackType:                req.FeedbackType,
		FeedbackText:                req.FeedbackText,
		CourseParticipationID:       req.CourseParticipationID,
		CoursePhaseID:               req.CoursePhaseID,
		AuthorCourseParticipationID: req.AuthorCourseParticipationID,
		Type:                        assessmentType.MapDTOtoDBAssessmentType(req.Type),
	})
	if err != nil {
		log.Error("could not create feedback item: ", err)
		return errors.New("could not create feedback item")
	}
	return nil
}

func UpdateFeedbackItem(ctx context.Context, req feedbackItemDTO.UpdateFeedbackItemRequest) error {
	err := FeedbackItemServiceSingleton.queries.UpdateFeedbackItem(ctx, db.UpdateFeedbackItemParams{
		ID:                          req.ID,
		FeedbackType:                req.FeedbackType,
		FeedbackText:                req.FeedbackText,
		CourseParticipationID:       req.CourseParticipationID,
		CoursePhaseID:               req.CoursePhaseID,
		AuthorCourseParticipationID: req.AuthorCourseParticipationID,
		Type:                        assessmentType.MapDTOtoDBAssessmentType(req.Type),
	})
	if err != nil {
		log.Error("could not update feedback item: ", err)
		return errors.New("could not update feedback item")
	}
	return nil
}

func DeleteFeedbackItem(ctx context.Context, feedbackItemID uuid.UUID) error {
	err := FeedbackItemServiceSingleton.queries.DeleteFeedbackItem(ctx, feedbackItemID)
	if err != nil {
		log.Error("could not delete feedback item: ", err)
		return errors.New("could not delete feedback item")
	}
	return nil
}

func IsFeedbackItemAuthor(ctx context.Context, feedbackItemID, authorID uuid.UUID) bool {
	feedbackItem, err := FeedbackItemServiceSingleton.queries.GetFeedbackItem(ctx, feedbackItemID)
	if err != nil {
		log.Error("Error fetching feedback item: ", err)
		return false
	}

	return feedbackItem.AuthorCourseParticipationID == authorID
}
