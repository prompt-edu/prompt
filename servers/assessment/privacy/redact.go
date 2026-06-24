package privacy

import (
	"context"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

func fetchAndApplyForEach[T any](ctx context.Context, ids []uuid.UUID, f func(context.Context, []uuid.UUID) ([]T, error), apply func(*T)) ([]T, error) {
	items, err := f(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		apply(&items[i])
	}
	return items, nil
}

func GetAllAssessmentsByCourseParticipationIDsNoAuthor(ctx context.Context, q db.Queries, ids []uuid.UUID) ([]db.Assessment, error) {
	return fetchAndApplyForEach(ctx, ids, q.GetAllAssessmentsByCourseParticipationIDs, func(assessment *db.Assessment) { assessment.Author = "" })
}

func GetAllAssessmentCompletionsByCourseParticipationIDsNoAuthor(ctx context.Context, q db.Queries, ids []uuid.UUID) ([]db.AssessmentCompletion, error) {
	return fetchAndApplyForEach(ctx, ids, q.GetAllAssessmentCompletionsByCourseParticipationIDs, func(assessment *db.AssessmentCompletion) { assessment.Author = "" })
}

func GetAllEvaluationsByCourseParticipationIDsNoAuthor(ctx context.Context, q db.Queries, ids []uuid.UUID) ([]db.Evaluation, error) {
	return fetchAndApplyForEach(ctx, ids, q.GetAllEvaluationsByCourseParticipationIDs, func(evaluation *db.Evaluation) { evaluation.AuthorCourseParticipationID = uuid.Nil })
}

func GetAllEvaluationCompletionsByCourseParticipationIDsNoAuthor(ctx context.Context, q db.Queries, ids []uuid.UUID) ([]db.EvaluationCompletion, error) {
	return fetchAndApplyForEach(ctx, ids, q.GetAllEvaluationCompletionsByCourseParticipationIDs, func(evalCompl *db.EvaluationCompletion) { evalCompl.AuthorCourseParticipationID = uuid.Nil })
}

func GetAllActionItemsByCourseParticipationIDsNoAuthor(ctx context.Context, q db.Queries, ids []uuid.UUID) ([]db.ActionItem, error) {
	return fetchAndApplyForEach(ctx, ids, q.GetAllActionItemsByCourseParticipationIDs, func(actionItem *db.ActionItem) { actionItem.Author = "" })
}

func GetAllFeedbackItemsByCourseParticipationIDsNoAuthor(ctx context.Context, q db.Queries, ids []uuid.UUID) ([]db.FeedbackItem, error) {
	return fetchAndApplyForEach(ctx, ids, q.GetAllFeedbackItemsByCourseParticipationIDs, func(feedbackItem *db.FeedbackItem) { feedbackItem.AuthorCourseParticipationID = uuid.Nil })
}
