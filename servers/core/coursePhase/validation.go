package coursePhase

import (
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseDTO"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func validateCreateCoursePhase(coursePhase coursePhaseDTO.CreateCoursePhase) error {
	if coursePhase.Name == "" {
		errorMessage := "course phase name is required"
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}

	if coursePhase.CourseID == uuid.Nil {
		errorMessage := "course id is required"
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}
	return nil
}

func validateUpdateCoursePhase(coursePhase coursePhaseDTO.UpdateCoursePhase) error {
	if coursePhase.Name.Valid && coursePhase.Name.String == "" {
		errorMessage := "course phase name is required"
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}

	return nil
}
