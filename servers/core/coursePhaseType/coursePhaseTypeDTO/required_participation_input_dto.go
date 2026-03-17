package coursePhaseTypeDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
)

type ParticipationInputDTO struct {
	ID                uuid.UUID     `json:"id"`
	CoursePhaseTypeID uuid.UUID     `json:"coursePhaseTypeID"`
	DtoName           string        `json:"dtoName"`
	Specification     meta.MetaData `json:"specification"` // the specification follows the same structure as the meta.MetaData
}

func GetParticipationInputDTOsFromDBModel(dbModel []db.CoursePhaseTypeParticipationRequiredInputDto) ([]ParticipationInputDTO, error) {
	var DTOs []ParticipationInputDTO

	for _, dbModel := range dbModel {
		dto, err := GetParticipationInputDTOFromDBModel(dbModel)
		if err != nil {
			return nil, err
		}
		DTOs = append(DTOs, dto)
	}

	return DTOs, nil
}

func GetParticipationInputDTOFromDBModel(dbModel db.CoursePhaseTypeParticipationRequiredInputDto) (ParticipationInputDTO, error) {
	specification, err := meta.GetMetaDataDTOFromDBModel(dbModel.Specification)
	if err != nil {
		return ParticipationInputDTO{}, err
	}

	return ParticipationInputDTO{
		ID:                dbModel.ID,
		CoursePhaseTypeID: dbModel.CoursePhaseTypeID,
		DtoName:           dbModel.DtoName,
		Specification:     specification,
	}, nil
}
