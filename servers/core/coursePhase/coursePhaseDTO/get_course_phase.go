package coursePhaseDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
)

type CoursePhase struct {
	ID                  uuid.UUID     `json:"id"`
	CourseID            uuid.UUID     `json:"courseID"`
	Name                string        `json:"name"`
	IsInitialPhase      bool          `json:"isInitialPhase"`
	RestrictedData      meta.MetaData `json:"restrictedData"`
	StudentReadableData meta.MetaData `json:"studentReadableData"`
	CoursePhaseTypeID   uuid.UUID     `json:"coursePhaseTypeID"`
	CoursePhaseTypeName string        `json:"coursePhaseTypeName"`
}

func GetCoursePhaseDTOFromDBModel(model db.GetCoursePhaseRow) (CoursePhase, error) {
	restrictedData, err := meta.GetMetaDataDTOFromDBModel(model.RestrictedData)
	if err != nil {
		return CoursePhase{}, err
	}

	studentReadableData, err := meta.GetMetaDataDTOFromDBModel(model.StudentReadableData)
	if err != nil {
		return CoursePhase{}, err
	}

	return CoursePhase{
		ID:                  model.ID,
		CourseID:            model.CourseID,
		Name:                model.Name.String,
		IsInitialPhase:      model.IsInitialPhase,
		RestrictedData:      restrictedData,
		StudentReadableData: studentReadableData,
		CoursePhaseTypeID:   model.CoursePhaseTypeID,
		CoursePhaseTypeName: model.CoursePhaseTypeName,
	}, nil
}
