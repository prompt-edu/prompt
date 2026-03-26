package applicationDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type QuestionFileUpload struct {
	ID                       uuid.UUID   `json:"id"`
	CoursePhaseID            uuid.UUID   `json:"coursePhaseID"`
	Title                    string      `json:"title"`
	Description              string      `json:"description"`
	IsRequired               bool        `json:"isRequired"`
	AllowedFileTypes         string      `json:"allowedFileTypes"`
	MaxFileSizeMB            int         `json:"maxFileSizeMB"`
	OrderNum                 int         `json:"orderNum"`
	AccessibleForOtherPhases pgtype.Bool `json:"accessibleForOtherPhases" swaggertype:"boolean"`
	AccessKey                pgtype.Text `json:"accessKey" swaggertype:"string"`
}

func (a QuestionFileUpload) GetDBModel() db.UpdateApplicationQuestionFileUploadParams {
	var description pgtype.Text
	if a.Description != "" {
		description = pgtype.Text{String: a.Description, Valid: true}
	}

	var allowedFileTypes pgtype.Text
	if a.AllowedFileTypes != "" {
		allowedFileTypes = pgtype.Text{String: a.AllowedFileTypes, Valid: true}
	}

	var maxFileSizeMB pgtype.Int4
	if a.MaxFileSizeMB > 0 {
		maxFileSizeMB = pgtype.Int4{Int32: int32(a.MaxFileSizeMB), Valid: true}
	}

	return db.UpdateApplicationQuestionFileUploadParams{
		ID:                       a.ID,
		Title:                    a.Title,
		Description:              description,
		IsRequired:               a.IsRequired,
		AllowedFileTypes:         allowedFileTypes,
		MaxFileSizeMb:            maxFileSizeMB,
		OrderNum:                 int32(a.OrderNum),
		AccessibleForOtherPhases: a.AccessibleForOtherPhases.Bool,
		AccessKey:                a.AccessKey,
	}
}

func GetQuestionFileUploadDTOFromDBModel(question db.ApplicationQuestionFileUpload) QuestionFileUpload {
	var description string
	if question.Description.Valid {
		description = question.Description.String
	}

	var allowedFileTypes string
	if question.AllowedFileTypes.Valid {
		allowedFileTypes = question.AllowedFileTypes.String
	}

	var maxFileSizeMB int
	if question.MaxFileSizeMb.Valid {
		maxFileSizeMB = int(question.MaxFileSizeMb.Int32)
	}

	return QuestionFileUpload{
		ID:                       question.ID,
		CoursePhaseID:            question.CoursePhaseID,
		Title:                    question.Title,
		Description:              description,
		IsRequired:               question.IsRequired,
		AllowedFileTypes:         allowedFileTypes,
		MaxFileSizeMB:            maxFileSizeMB,
		OrderNum:                 int(question.OrderNum),
		AccessibleForOtherPhases: pgtype.Bool{Bool: question.AccessibleForOtherPhases, Valid: true},
		AccessKey:                question.AccessKey,
	}
}
