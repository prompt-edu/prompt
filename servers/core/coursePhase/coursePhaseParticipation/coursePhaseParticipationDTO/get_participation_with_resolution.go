package coursePhaseParticipationDTO

import "github.com/prompt-edu/prompt/servers/core/coursePhase/resolution/resolutionDTO"

type CoursePhaseParticipationWithResolution struct {
	Participation GetAllCPPsForCoursePhase   `json:"participation"`
	Resolutions   []resolutionDTO.Resolution `json:"resolutions"`
}
