package privacy

import (
	"context"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

func GetAllAssessmentsByCourseParticipationIDsNoAuthor(ctx context.Context, q db.Queries, ids []uuid.UUID) ([]db.Assessment, error) {
	items, err := q.GetAllAssessmentsByCourseParticipationIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Author = ""
	}
	return items, nil
}

func GetAllAssessmentCompletionsByCourseParticipationIDsNoAuthor(ctx context.Context, q db.Queries, ids []uuid.UUID) ([]db.AssessmentCompletion, error) {
	items, err := q.GetAllAssessmentCompletionsByCourseParticipationIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Author = ""
	}
	return items, nil
}

func GetAllEvaluationsByCourseParticipationIDsNoAuthor(ctx context.Context, q db.Queries, ids []uuid.UUID) ([]db.Evaluation, error) {
	items, err := q.GetAllEvaluationsByCourseParticipationIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].AuthorCourseParticipationID = uuid.Nil
	}
	return items, nil
}

func GetAllEvaluationCompletionsByCourseParticipationIDsNoAuthor(ctx context.Context, q db.Queries, ids []uuid.UUID) ([]db.EvaluationCompletion, error) {
	items, err := q.GetAllEvaluationCompletionsByCourseParticipationIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].AuthorCourseParticipationID = uuid.Nil
	}
	return items, nil
}

func GetAllActionItemsByCourseParticipationIDsNoAuthor(ctx context.Context, q db.Queries, ids []uuid.UUID) ([]db.ActionItem, error) {
	items, err := q.GetAllActionItemsByCourseParticipationIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Author = ""
	}
	return items, nil
}

func GetAllFeedbackItemsByCourseParticipationIDsNoAuthor(ctx context.Context, q db.Queries, ids []uuid.UUID) ([]db.FeedbackItem, error) {
	items, err := q.GetAllFeedbackItemsByCourseParticipationIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].AuthorCourseParticipationID = uuid.Nil
	}
	return items, nil
}
