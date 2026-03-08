package applicationDTO

import (
	"time"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type AnswerFileUpload struct {
	ID                    uuid.UUID `json:"id"`
	ApplicationQuestionID uuid.UUID `json:"applicationQuestionID"`
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	FileID                uuid.UUID `json:"fileID"`
	FileName              string    `json:"fileName"`
	FileSize              int64     `json:"fileSize"`
	UploadedAt            time.Time `json:"uploadedAt"`
	DownloadURL           string    `json:"downloadUrl"`
}

func (a AnswerFileUpload) GetDBModel() db.ApplicationAnswerFileUpload {
	return db.ApplicationAnswerFileUpload{
		ID:                    a.ID,
		ApplicationQuestionID: a.ApplicationQuestionID,
		CourseParticipationID: a.CourseParticipationID,
		FileID:                a.FileID,
	}
}

func GetAnswerFileUploadDTOFromDBModel(answer db.ApplicationAnswerFileUpload) AnswerFileUpload {
	return AnswerFileUpload{
		ID:                    answer.ID,
		ApplicationQuestionID: answer.ApplicationQuestionID,
		CourseParticipationID: answer.CourseParticipationID,
		FileID:                answer.FileID,
	}
}

func GetAnswersFileUploadDTOFromDBModels(answers []db.ApplicationAnswerFileUpload) []AnswerFileUpload {
	answerFileUploadDTO := make([]AnswerFileUpload, 0, len(answers))
	for _, answer := range answers {
		answerFileUploadDTO = append(answerFileUploadDTO, GetAnswerFileUploadDTOFromDBModel(answer))
	}

	return answerFileUploadDTO
}
