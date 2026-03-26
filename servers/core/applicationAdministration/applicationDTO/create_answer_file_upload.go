package applicationDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type CreateAnswerFileUpload struct {
	ApplicationQuestionID uuid.UUID `json:"applicationQuestionID"`
	FileID                uuid.UUID `json:"fileID"`
}

func (a CreateAnswerFileUpload) GetDBModel() db.CreateApplicationAnswerFileUploadParams {
	return db.CreateApplicationAnswerFileUploadParams{
		ApplicationQuestionID: a.ApplicationQuestionID,
		FileID:                a.FileID,
	}
}
