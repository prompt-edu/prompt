package applicationDTO

import (
	"github.com/google/uuid"
)

type CreateAnswerFileUpload struct {
	ApplicationQuestionID uuid.UUID `json:"applicationQuestionID"`
	FileID                uuid.UUID `json:"fileID"`
}
