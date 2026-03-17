package coursePhaseTypeDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
	log "github.com/sirupsen/logrus"
)

type ParticipationOutputDTO struct {
	ID                uuid.UUID     `json:"id"`
	CoursePhaseTypeID uuid.UUID     `json:"coursePhaseTypeID"`
	DtoName           string        `json:"dtoName"`
	Specification     meta.MetaData `json:"specification"` // the specification follows the same structure as the meta.MetaData
	VersionNumber     int32         `json:"versionNumber"`
	EndpointPath      string        `json:"endpointPath"`
}

func GetParticipationOutputDTOsFromDBModel(dbModel []db.CoursePhaseTypeParticipationProvidedOutputDto) ([]ParticipationOutputDTO, error) {
	var DTOs []ParticipationOutputDTO

	for _, dbModel := range dbModel {
		dto, err := GetParticipationOutputDTOFromDBModel(dbModel)
		if err != nil {
			log.Error("Failed to get ProvidedOutputDTO from DB model")
			return nil, err
		}
		DTOs = append(DTOs, dto)
	}

	return DTOs, nil
}

func GetParticipationOutputDTOFromDBModel(dbModel db.CoursePhaseTypeParticipationProvidedOutputDto) (ParticipationOutputDTO, error) {
	specification, err := meta.GetMetaDataDTOFromDBModel(dbModel.Specification)
	if err != nil {
		return ParticipationOutputDTO{}, err
	}

	return ParticipationOutputDTO{
		ID:                dbModel.ID,
		CoursePhaseTypeID: dbModel.CoursePhaseTypeID,
		DtoName:           dbModel.DtoName,
		Specification:     specification,
		VersionNumber:     dbModel.VersionNumber,
		EndpointPath:      dbModel.EndpointPath,
	}, nil
}
