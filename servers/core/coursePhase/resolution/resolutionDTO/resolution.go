package resolutionDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type Resolution struct {
	DtoName       string    `json:"dtoName"`
	BaseURL       string    `json:"baseURL"`
	EndpointPath  string    `json:"endpointPath"`
	CoursePhaseID uuid.UUID `json:"coursePhaseID"`
}

// for the phase level resolution
func GetPhaseResolutionDTOFromDBModel(model db.GetPrevCoursePhaseDataResolutionRow) Resolution {
	return Resolution{
		DtoName:       model.DtoName,
		BaseURL:       model.BaseUrl,
		EndpointPath:  model.EndpointPath,
		CoursePhaseID: model.FromCoursePhaseID,
	}
}

// for the participation level resolution
func GetParticipationResolutionDTOFromDBModel(model db.GetResolutionsForCoursePhaseRow) Resolution {
	return Resolution{
		DtoName:       model.DtoName,
		BaseURL:       model.BaseUrl,
		EndpointPath:  model.EndpointPath,
		CoursePhaseID: model.FromCoursePhaseID,
	}
}

func GetParticipationResolutionsDTOFromDBModels(models []db.GetResolutionsForCoursePhaseRow) []Resolution {
	resolutionDTOs := make([]Resolution, 0, len(models))

	for _, resolution := range models {
		dto := GetParticipationResolutionDTOFromDBModel(resolution)
		resolutionDTOs = append(resolutionDTOs, dto)
	}
	return resolutionDTOs
}
