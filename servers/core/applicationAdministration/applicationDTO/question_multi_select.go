package applicationDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type QuestionMultiSelect struct {
	ID            uuid.UUID `json:"id"`
	CoursePhaseID uuid.UUID `json:"coursePhaseID"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Placeholder   string    `json:"placeholder"`
	ErrorMessage  string    `json:"errorMessage"`
	IsRequired    bool      `json:"isRequired"`
	MinSelect     int       `json:"minSelect"`
	MaxSelect     int       `json:"maxSelect"`
	Options       []string  `json:"options"`
	OrderNum      int       `json:"orderNum"`
	// using pgtype as this allows for optional values
	AccessibleForOtherPhases pgtype.Bool `json:"accessibleForOtherPhases" swaggertype:"boolean"`
	AccessKey                pgtype.Text `json:"accessKey" swaggertype:"string"`
}

func (a QuestionMultiSelect) GetDBModel() db.UpdateApplicationQuestionMultiSelectParams {
	return db.UpdateApplicationQuestionMultiSelectParams{
		ID:                       a.ID,
		Title:                    pgtype.Text{String: a.Title, Valid: true},
		Description:              pgtype.Text{String: a.Description, Valid: true},
		Placeholder:              pgtype.Text{String: a.Placeholder, Valid: true},
		ErrorMessage:             pgtype.Text{String: a.ErrorMessage, Valid: true},
		IsRequired:               pgtype.Bool{Bool: a.IsRequired, Valid: true},
		MinSelect:                pgtype.Int4{Int32: int32(a.MinSelect), Valid: true},
		MaxSelect:                pgtype.Int4{Int32: int32(a.MaxSelect), Valid: true},
		Options:                  a.Options,
		OrderNum:                 pgtype.Int4{Int32: int32(a.OrderNum), Valid: true},
		AccessibleForOtherPhases: a.AccessibleForOtherPhases,
		AccessKey:                a.AccessKey,
	}

}

func GetQuestionMultiSelectDTOFromDBModel(applicationQuestionMultiSelect db.ApplicationQuestionMultiSelect) QuestionMultiSelect {
	return QuestionMultiSelect{
		ID:                       applicationQuestionMultiSelect.ID,
		CoursePhaseID:            applicationQuestionMultiSelect.CoursePhaseID,
		Title:                    applicationQuestionMultiSelect.Title.String,
		Description:              applicationQuestionMultiSelect.Description.String,
		Placeholder:              applicationQuestionMultiSelect.Placeholder.String,
		ErrorMessage:             applicationQuestionMultiSelect.ErrorMessage.String,
		IsRequired:               applicationQuestionMultiSelect.IsRequired.Bool,
		MinSelect:                int(applicationQuestionMultiSelect.MinSelect.Int32),
		MaxSelect:                int(applicationQuestionMultiSelect.MaxSelect.Int32),
		Options:                  applicationQuestionMultiSelect.Options,
		OrderNum:                 int(applicationQuestionMultiSelect.OrderNum.Int32),
		AccessibleForOtherPhases: applicationQuestionMultiSelect.AccessibleForOtherPhases,
		AccessKey:                applicationQuestionMultiSelect.AccessKey,
	}
}
