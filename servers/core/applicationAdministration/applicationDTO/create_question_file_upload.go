package applicationDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type CreateQuestionFileUpload struct {
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

func (a CreateQuestionFileUpload) GetDBModel() db.CreateApplicationQuestionFileUploadParams {
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

	return db.CreateApplicationQuestionFileUploadParams{
		CoursePhaseID:            a.CoursePhaseID,
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
