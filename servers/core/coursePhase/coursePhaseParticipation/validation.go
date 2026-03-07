package coursePhaseParticipation

import (
	"errors"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseParticipation/coursePhaseParticipationDTO"
	log "github.com/sirupsen/logrus"
)

func Validate(c coursePhaseParticipationDTO.CreateCoursePhaseParticipation) error {
	if c.CourseParticipationID == uuid.Nil {
		errorMessage := ("validation error: course participation id is required")
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}

	if c.CoursePhaseID == uuid.Nil {
		errorMessage := ("validation error: course phase id is required")
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}

	return nil
}
