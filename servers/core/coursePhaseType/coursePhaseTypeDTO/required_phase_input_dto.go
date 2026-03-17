package coursePhaseTypeDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
)

type PhaseInputDTO struct {
	ID                uuid.UUID     `json:"id"`
	CoursePhaseTypeID uuid.UUID     `json:"coursePhaseTypeID"`
	DtoName           string        `json:"dtoName"`
	Specification     meta.MetaData `json:"specification"` // the specification follows the same structure as the meta.MetaData
}

func GetPhaseInputDTOsFromDBModel(dbModel []db.CoursePhaseTypePhaseRequiredInputDto) ([]PhaseInputDTO, error) {
	var DTOs []PhaseInputDTO

	for _, dbModel := range dbModel {
		dto, err := GetPhaseInputDTOFromDBModel(dbModel)
		if err != nil {
			return nil, err
		}
		DTOs = append(DTOs, dto)
	}

	return DTOs, nil
}

func GetPhaseInputDTOFromDBModel(dbModel db.CoursePhaseTypePhaseRequiredInputDto) (PhaseInputDTO, error) {
	specification, err := meta.GetMetaDataDTOFromDBModel(dbModel.Specification)
	if err != nil {
		return PhaseInputDTO{}, err
	}

	return PhaseInputDTO{
		ID:                dbModel.ID,
		CoursePhaseTypeID: dbModel.CoursePhaseTypeID,
		DtoName:           dbModel.DtoName,
		Specification:     specification,
	}, nil
}
