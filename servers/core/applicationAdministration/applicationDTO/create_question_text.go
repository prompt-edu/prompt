package applicationDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type CreateQuestionText struct {
	CoursePhaseID   uuid.UUID `json:"coursePhaseID"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Placeholder     string    `json:"placeholder"`
	ValidationRegex string    `json:"validationRegex"`
	ErrorMessage    string    `json:"errorMessage"`
	IsRequired      bool      `json:"isRequired"`
	AllowedLength   int       `json:"allowedLength"`
	OrderNum        int       `json:"orderNum"`
	// using pgtype as this allows for optional values
	AccessibleForOtherPhases pgtype.Bool `json:"accessibleForOtherPhases" swaggertype:"boolean"`
	AccessKey                pgtype.Text `json:"accessKey" swaggertype:"string"`
}

func (a CreateQuestionText) GetDBModel() db.CreateApplicationQuestionTextParams {
	return db.CreateApplicationQuestionTextParams{
		CoursePhaseID:            a.CoursePhaseID,
		Title:                    pgtype.Text{String: a.Title, Valid: true},
		Description:              pgtype.Text{String: a.Description, Valid: true},
		Placeholder:              pgtype.Text{String: a.Placeholder, Valid: true},
		ValidationRegex:          pgtype.Text{String: a.ValidationRegex, Valid: true},
		ErrorMessage:             pgtype.Text{String: a.ErrorMessage, Valid: true},
		IsRequired:               pgtype.Bool{Bool: a.IsRequired, Valid: true},
		AllowedLength:            pgtype.Int4{Int32: int32(a.AllowedLength), Valid: true},
		OrderNum:                 pgtype.Int4{Int32: int32(a.OrderNum), Valid: true},
		AccessibleForOtherPhases: a.AccessibleForOtherPhases,
		AccessKey:                a.AccessKey,
	}
}
