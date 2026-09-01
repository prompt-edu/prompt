package interviewReview

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
	"github.com/prompt-edu/prompt/servers/interview/interviewReview/interviewReviewDTO"
	log "github.com/sirupsen/logrus"
)

type InterviewReviewService struct {
	queries db.Queries
}

var errScoreOutOfRange = errors.New("score must be between 1 and 5")

func NewInterviewReviewService(queries db.Queries) *InterviewReviewService {
	return &InterviewReviewService{queries: queries}
}

func (s *InterviewReviewService) GetInterviewReviews(ctx context.Context, coursePhaseID uuid.UUID) ([]interviewReviewDTO.InterviewReview, error) {
	reviews, err := s.queries.GetInterviewReviewsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("failed to fetch interview reviews: ", err)
		return nil, fmt.Errorf("failed to fetch interview reviews: %w", err)
	}
	return interviewReviewDTO.GetInterviewReviewsFromDB(reviews), nil
}

func (s *InterviewReviewService) UpsertInterviewReview(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID, req interviewReviewDTO.UpdateInterviewReviewRequest) (interviewReviewDTO.InterviewReview, error) {
	answers, err := interviewReviewDTO.EncodeInterviewAnswers(req.InterviewAnswers)
	if err != nil {
		log.Error("failed to encode interview answers: ", err)
		return interviewReviewDTO.InterviewReview{}, fmt.Errorf("failed to encode interview answers: %w", err)
	}

	score := pgtype.Int4{}
	if req.Score != nil {
		score = pgtype.Int4{Int32: *req.Score, Valid: true}
	}

	review, err := s.queries.UpsertInterviewReview(ctx, db.UpsertInterviewReviewParams{
		CoursePhaseID:         coursePhaseID,
		CourseParticipationID: courseParticipationID,
		Score:                 score,
		Interviewer:           pgtype.Text{String: req.Interviewer, Valid: true},
		InterviewAnswers:      answers,
	})
	if err != nil {
		log.Error("failed to upsert interview review: ", err)
		return interviewReviewDTO.InterviewReview{}, fmt.Errorf("failed to upsert interview review: %w", err)
	}

	return interviewReviewDTO.GetInterviewReviewFromDB(review), nil
}

func (s *InterviewReviewService) GetScores(ctx context.Context, coursePhaseID uuid.UUID) ([]interviewReviewDTO.ScoreWithParticipation, error) {
	reviews, err := s.queries.GetScoredInterviewReviewsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("failed to fetch interview scores: ", err)
		return nil, fmt.Errorf("failed to fetch interview scores: %w", err)
	}

	scores := make([]interviewReviewDTO.ScoreWithParticipation, 0, len(reviews))
	for _, review := range reviews {
		if !review.Score.Valid {
			continue
		}
		scores = append(scores, interviewReviewDTO.ScoreWithParticipation{
			CourseParticipationID: review.CourseParticipationID,
			Score:                 review.Score.Int32,
		})
	}
	return scores, nil
}

func (s *InterviewReviewService) GetScoreLevels(ctx context.Context, coursePhaseID uuid.UUID) ([]interviewReviewDTO.ScoreLevelWithParticipation, error) {
	reviews, err := s.queries.GetScoredInterviewReviewsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("failed to fetch interview score levels: ", err)
		return nil, fmt.Errorf("failed to fetch interview score levels: %w", err)
	}

	scoreLevels := make([]interviewReviewDTO.ScoreLevelWithParticipation, 0, len(reviews))
	for _, review := range reviews {
		if !review.Score.Valid {
			continue
		}
		level, ok := interviewReviewDTO.DeriveScoreLevel(review.Score.Int32)
		if !ok {
			continue
		}
		scoreLevels = append(scoreLevels, interviewReviewDTO.ScoreLevelWithParticipation{
			CourseParticipationID: review.CourseParticipationID,
			ScoreLevel:            level,
		})
	}
	return scoreLevels, nil
}
