package feedbackItem

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/feedbackItem/feedbackItemDTO"
	log "github.com/sirupsen/logrus"
)

var ErrFeedbackItemNotFound = errors.New("feedback item not found")
var ErrNotFeedbackItemAuthor = errors.New("you are not the author of this feedback item")

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
	tx, err := FeedbackItemServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := FeedbackItemServiceSingleton.queries.WithTx(tx)

	err = evaluationCompletion.CheckEvaluationIsEditable(ctx, qtx, req.CourseParticipationID, req.CoursePhaseID, req.AuthorCourseParticipationID, req.Type)
	if err != nil {
		return err
	}

	err = qtx.CreateFeedbackItem(ctx, db.CreateFeedbackItemParams{
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

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func UpdateFeedbackItem(ctx context.Context, feedbackItemID, coursePhaseID, authorCourseParticipationID uuid.UUID, req feedbackItemDTO.UpdateFeedbackItemRequest) error {
	tx, err := FeedbackItemServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := FeedbackItemServiceSingleton.queries.WithTx(tx)

	existing, err := qtx.GetFeedbackItem(ctx, feedbackItemID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrFeedbackItemNotFound
		}
		log.Error("could not get feedback item: ", err)
		return errors.New("could not get feedback item")
	}

	if existing.CoursePhaseID != coursePhaseID {
		return ErrFeedbackItemNotFound
	}
	if existing.AuthorCourseParticipationID != authorCourseParticipationID {
		return ErrNotFeedbackItemAuthor
	}

	err = evaluationCompletion.CheckEvaluationIsEditable(ctx, qtx, existing.CourseParticipationID, existing.CoursePhaseID, existing.AuthorCourseParticipationID, assessmentType.MapDBAssessmentTypeToDTO(existing.Type))
	if err != nil {
		return err
	}

	rows, err := qtx.UpdateFeedbackItem(ctx, db.UpdateFeedbackItemParams{
		ID:                          feedbackItemID,
		CoursePhaseID:               coursePhaseID,
		AuthorCourseParticipationID: authorCourseParticipationID,
		FeedbackType:                req.FeedbackType,
		FeedbackText:                req.FeedbackText,
	})
	if err != nil {
		log.Error("could not update feedback item: ", err)
		return errors.New("could not update feedback item")
	}
	if rows == 0 {
		return ErrFeedbackItemNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
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
