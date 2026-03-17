package coursePhaseTypeDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type CoursePhaseType struct {
	ID                              uuid.UUID                `json:"id"`
	Name                            string                   `json:"name"`
	BaseUrl                         string                   `json:"baseUrl"`
	InitialPhase                    bool                     `json:"initialPhase"`
  Description                     pgtype.Text              `json:"description"`
	RequiredParticipationInputDTOs  []ParticipationInputDTO  `json:"requiredParticipationInputDTOs"`
	ProvidedParticipationOutputDTOs []ParticipationOutputDTO `json:"providedParticipationOutputDTOs"`
	RequiredPhaseInputDTOs          []PhaseInputDTO          `json:"requiredPhaseInputDTOs"`
	ProvidedPhaseOutputDTOs         []PhaseOutputDTO         `json:"providedPhaseOutputDTOs"`
}

func GetCoursePhaseTypeDTOFromDBModel(model db.CoursePhaseType, requiredParticipationInputs []ParticipationInputDTO, providedParticipationOutputs []ParticipationOutputDTO, requiredPhaseInputs []PhaseInputDTO, providedPhaseOutputs []PhaseOutputDTO) (CoursePhaseType, error) {
	return CoursePhaseType{
		ID:                              model.ID,
		Name:                            model.Name,
		BaseUrl:                         model.BaseUrl,
		InitialPhase:                    model.InitialPhase,
    Description:                     model.Description,
		RequiredParticipationInputDTOs:  requiredParticipationInputs,
		ProvidedParticipationOutputDTOs: providedParticipationOutputs,
		RequiredPhaseInputDTOs:          requiredPhaseInputs,
		ProvidedPhaseOutputDTOs:         providedPhaseOutputs,
	}, nil
}
