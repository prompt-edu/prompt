package coursePhaseDTO

import (
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution/resolutionDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
)

type PrevCoursePhaseData struct {
	PrevData    meta.MetaData              `json:"prevData"`
	Resolutions []resolutionDTO.Resolution `json:"resolutions"`
}

func GetPrevCoursePhaseDataDTO(prevCoreData []byte, resolutions []db.GetPrevCoursePhaseDataResolutionRow) (PrevCoursePhaseData, error) {
	prevData, err := meta.GetMetaDataDTOFromDBModel(prevCoreData)
	if err != nil {
		return PrevCoursePhaseData{}, err
	}

	resolutionDTOs := make([]resolutionDTO.Resolution, 0, len(resolutions))
	for _, resolution := range resolutions {
		resolutionDTOs = append(resolutionDTOs, resolutionDTO.GetPhaseResolutionDTOFromDBModel(resolution))
	}

	return PrevCoursePhaseData{
		PrevData:    prevData,
		Resolutions: resolutionDTOs,
	}, nil
}
