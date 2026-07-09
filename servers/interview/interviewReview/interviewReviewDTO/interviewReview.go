package interviewReviewDTO

import (
	"encoding/json"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
)

// ScoreLevel is the categorized interview outcome derived from the numeric score.
type ScoreLevel string

const (
	ScoreLevelVeryBad  ScoreLevel = "veryBad"
	ScoreLevelBad      ScoreLevel = "bad"
	ScoreLevelOk       ScoreLevel = "ok"
	ScoreLevelGood     ScoreLevel = "good"
	ScoreLevelVeryGood ScoreLevel = "veryGood"
)

// InterviewAnswer stores a single answer to a configured interview question.
type InterviewAnswer struct {
	QuestionID int    `json:"questionID"`
	Answer     string `json:"answer"`
}

// InterviewReview is the full review record for a single course participation.
type InterviewReview struct {
	CourseParticipationID uuid.UUID         `json:"courseParticipationID"`
	Score                 *int32            `json:"score,omitempty"`
	ScoreLevel            *ScoreLevel       `json:"scoreLevel,omitempty"`
	Interviewer           string            `json:"interviewer"`
	InterviewAnswers      []InterviewAnswer `json:"interviewAnswers"`
}

// UpdateInterviewReviewRequest is the payload used to create or update a review.
type UpdateInterviewReviewRequest struct {
	Score            *int32            `json:"score"`
	Interviewer      string            `json:"interviewer"`
	InterviewAnswers []InterviewAnswer `json:"interviewAnswers"`
}

// ScoreWithParticipation is the data-graph output DTO for the "score" output.
type ScoreWithParticipation struct {
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	Score                 int32     `json:"score"`
}

// ScoreLevelWithParticipation is the data-graph output DTO for the "scoreLevel" output.
type ScoreLevelWithParticipation struct {
	CourseParticipationID uuid.UUID  `json:"courseParticipationID"`
	ScoreLevel            ScoreLevel `json:"scoreLevel"`
}

// DeriveScoreLevel maps a numeric interview score to its categorized level.
// Scores outside the valid 1-5 range have no level (ok == false), mirroring the
// previous core SQL derivation.
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

// GetInterviewReviewFromDB maps a database record to the API DTO, deriving the
// score level and decoding the stored interview answers.
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

// GetInterviewReviewsFromDB maps a list of database records to API DTOs.
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
		return make([]InterviewAnswer, 0)
	}
	return answers
}

// EncodeInterviewAnswers serializes interview answers for storage as jsonb.
func EncodeInterviewAnswers(answers []InterviewAnswer) ([]byte, error) {
	if answers == nil {
		answers = make([]InterviewAnswer, 0)
	}
	return json.Marshal(answers)
}
