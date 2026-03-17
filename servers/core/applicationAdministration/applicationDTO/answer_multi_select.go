package applicationDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type AnswerMultiSelect struct {
	ID                    uuid.UUID `json:"id"`
	ApplicationQuestionID uuid.UUID `json:"applicationQuestionID"`
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	CoursePhaseID         uuid.UUID `json:"coursePhaseID"`
	Answer                []string  `json:"answer"`
}

func (a AnswerMultiSelect) GetDBModel() db.ApplicationAnswerMultiSelect {
	return db.ApplicationAnswerMultiSelect{
		ID:                    a.ID,
		ApplicationQuestionID: a.ApplicationQuestionID,
		CourseParticipationID: a.CourseParticipationID,
		Answer:                a.Answer,
	}
}

func GetAnswerMultiSelectDTOFromDBModel(answer db.ApplicationAnswerMultiSelect) AnswerMultiSelect {
	return AnswerMultiSelect{
		ID:                    answer.ID,
		ApplicationQuestionID: answer.ApplicationQuestionID,
		CourseParticipationID: answer.CourseParticipationID,
		Answer:                answer.Answer,
	}
}

func GetAnswersMultiSelectDTOFromDBModels(answers []db.ApplicationAnswerMultiSelect) []AnswerMultiSelect {
	answerTextDTO := make([]AnswerMultiSelect, 0, len(answers))
	for _, answer := range answers {
		answerTextDTO = append(answerTextDTO, GetAnswerMultiSelectDTOFromDBModel(answer))
	}

	return answerTextDTO
}
