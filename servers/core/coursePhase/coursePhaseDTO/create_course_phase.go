package coursePhaseDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
)

type CreateCoursePhase struct {
	CourseID            uuid.UUID     `json:"courseID"`
	Name                string        `json:"name"`
	IsInitialPhase      bool          `json:"isInitialPhase"`
	RestrictedData      meta.MetaData `json:"restrictedData"`
	StudentReadableData meta.MetaData `json:"studentReadableData"`
	CoursePhaseTypeID   uuid.UUID     `json:"coursePhaseTypeID"`
}

func (cp CreateCoursePhase) GetDBModel() (db.CreateCoursePhaseParams, error) {
	restrictedData, err := cp.RestrictedData.GetDBModel()
	if err != nil {
		return db.CreateCoursePhaseParams{}, err
	}

	studentReadableData, err := cp.StudentReadableData.GetDBModel()
	if err != nil {
		return db.CreateCoursePhaseParams{}, err
	}

	return db.CreateCoursePhaseParams{
		CourseID:            cp.CourseID,
		Name:                pgtype.Text{String: cp.Name, Valid: true},
		IsInitialPhase:      cp.IsInitialPhase,
		RestrictedData:      restrictedData,
		StudentReadableData: studentReadableData,
		CoursePhaseTypeID:   cp.CoursePhaseTypeID,
	}, nil
}
