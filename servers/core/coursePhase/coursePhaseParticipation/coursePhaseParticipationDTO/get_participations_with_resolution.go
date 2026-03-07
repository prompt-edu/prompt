package coursePhaseParticipationDTO

import "github.com/prompt-edu/prompt/servers/core/coursePhase/resolution/resolutionDTO"

type CoursePhaseParticipationsWithResolutions struct {
	Participations []GetAllCPPsForCoursePhase `json:"participations"`
	Resolutions    []resolutionDTO.Resolution `json:"resolutions"`
}
