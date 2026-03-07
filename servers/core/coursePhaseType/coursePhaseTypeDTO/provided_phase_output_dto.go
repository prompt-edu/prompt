package coursePhaseTypeDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
	log "github.com/sirupsen/logrus"
)

type PhaseOutputDTO struct {
	ID                uuid.UUID     `json:"id"`
	CoursePhaseTypeID uuid.UUID     `json:"coursePhaseTypeID"`
	DtoName           string        `json:"dtoName"`
	Specification     meta.MetaData `json:"specification"` // the specification follows the same structure as the meta.MetaData
	VersionNumber     int32         `json:"versionNumber"`
	EndpointPath      string        `json:"endpointPath"`
}

func GetPhaseOutputDTOsFromDBModel(dbModel []db.CoursePhaseTypePhaseProvidedOutputDto) ([]PhaseOutputDTO, error) {
	var DTOs []PhaseOutputDTO

	for _, dbModel := range dbModel {
		dto, err := GetPhaseOutputDTOFromDBModel(dbModel)
		if err != nil {
			log.Error("Failed to get ProvidedOutputDTO from DB model")
			return nil, err
		}
		DTOs = append(DTOs, dto)
	}

	return DTOs, nil
}

func GetPhaseOutputDTOFromDBModel(dbModel db.CoursePhaseTypePhaseProvidedOutputDto) (PhaseOutputDTO, error) {
	specification, err := meta.GetMetaDataDTOFromDBModel(dbModel.Specification)
	if err != nil {
		return PhaseOutputDTO{}, err
	}

	return PhaseOutputDTO{
		ID:                dbModel.ID,
		CoursePhaseTypeID: dbModel.CoursePhaseTypeID,
		DtoName:           dbModel.DtoName,
		Specification:     specification,
		VersionNumber:     dbModel.VersionNumber,
		EndpointPath:      dbModel.EndpointPath,
	}, nil
}
