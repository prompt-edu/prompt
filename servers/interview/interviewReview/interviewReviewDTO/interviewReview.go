package interviewReviewDTO

import (
	"encoding/json"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
	log "github.com/sirupsen/logrus"
)

type ScoreLevel string

const (
	ScoreLevelVeryBad  ScoreLevel = "veryBad"
	ScoreLevelBad      ScoreLevel = "bad"
	ScoreLevelOk       ScoreLevel = "ok"
	ScoreLevelGood     ScoreLevel = "good"
	ScoreLevelVeryGood ScoreLevel = "veryGood"
)

type InterviewAnswer struct {
	QuestionID int    `json:"questionID"`
	Answer     string `json:"answer"`
}

type InterviewReview struct {
	CourseParticipationID uuid.UUID         `json:"courseParticipationID"`
	Score                 *int32            `json:"score,omitempty"`
	ScoreLevel            *ScoreLevel       `json:"scoreLevel,omitempty"`
	Interviewer           string            `json:"interviewer"`
	InterviewAnswers      []InterviewAnswer `json:"interviewAnswers"`
}

type UpdateInterviewReviewRequest struct {
	Score            *int32            `json:"score"`
	Interviewer      string            `json:"interviewer"`
	InterviewAnswers []InterviewAnswer `json:"interviewAnswers"`
}

type ScoreWithParticipation struct {
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	Score                 int32     `json:"score"`
}

type ScoreLevelWithParticipation struct {
	CourseParticipationID uuid.UUID  `json:"courseParticipationID"`
	ScoreLevel            ScoreLevel `json:"scoreLevel"`
}

func DeriveScoreLevel(score int32) (ScoreLevel, bool) {
	switch {
	case score < 1 || score > 5:
		return "", false
	case score <= 1:
		return ScoreLevelVeryGood, true
	case score <= 2:
		return ScoreLevelGood, true
	case score <= 3:
		return ScoreLevelOk, true
	case score <= 4:
		return ScoreLevelBad, true
	default:
		return ScoreLevelVeryBad, true
	}
}

func GetInterviewReviewFromDB(review db.InterviewReview) InterviewReview {
	dto := InterviewReview{
		CourseParticipationID: review.CourseParticipationID,
		Interviewer:           review.Interviewer.String,
		InterviewAnswers:      decodeInterviewAnswers(review.InterviewAnswers),
	}

	if review.Score.Valid {
		score := review.Score.Int32
		dto.Score = &score
		if level, ok := DeriveScoreLevel(score); ok {
			dto.ScoreLevel = &level
		}
	}

	return dto
}

func GetInterviewReviewsFromDB(reviews []db.InterviewReview) []InterviewReview {
	result := make([]InterviewReview, 0, len(reviews))
	for _, review := range reviews {
		result = append(result, GetInterviewReviewFromDB(review))
	}
	return result
}

func decodeInterviewAnswers(raw []byte) []InterviewAnswer {
	answers := make([]InterviewAnswer, 0)
	if len(raw) == 0 {
		return answers
	}
	if err := json.Unmarshal(raw, &answers); err != nil {
		log.Warn("failed to unmarshal interview answers, returning empty slice: ", err)
		return make([]InterviewAnswer, 0)
	}
	return answers
}

func EncodeInterviewAnswers(answers []InterviewAnswer) ([]byte, error) {
	if answers == nil {
		answers = make([]InterviewAnswer, 0)
	}
	return json.Marshal(answers)
}
