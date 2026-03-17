package applicationDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type AnswerText struct {
	ID                    uuid.UUID `json:"id"`
	ApplicationQuestionID uuid.UUID `json:"applicationQuestionID"`
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	Answer                string    `json:"answer"`
}

func (a AnswerText) GetDBModel() db.ApplicationAnswerText {
	return db.ApplicationAnswerText{
		ID:                    a.ID,
		ApplicationQuestionID: a.ApplicationQuestionID,
		CourseParticipationID: a.CourseParticipationID,
		Answer:                pgtype.Text{String: a.Answer, Valid: true},
	}
}

func GetAnswerTextDTOFromDBModel(answer db.ApplicationAnswerText) AnswerText {
	return AnswerText{
		ID:                    answer.ID,
		ApplicationQuestionID: answer.ApplicationQuestionID,
		CourseParticipationID: answer.CourseParticipationID,
		Answer:                answer.Answer.String,
	}
}

func GetAnswersTextDTOFromDBModels(answers []db.ApplicationAnswerText) []AnswerText {
	answerTextDTO := make([]AnswerText, 0, len(answers))
	for _, answer := range answers {
		answerTextDTO = append(answerTextDTO, GetAnswerTextDTOFromDBModel(answer))
	}

	return answerTextDTO
}
