package coursePhase

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseDTO"
	"github.com/prompt-edu/prompt/servers/core/meta"
	log "github.com/sirupsen/logrus"
)

// welcomeTextKey is served to unauthenticated applicants by the apply endpoint, so
// its size is capped here even though the rest of the metadata is schema-less.
const (
	welcomeTextKey      = "welcomeText"
	maxWelcomeTextBytes = 20000
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

	return validateWelcomeText(coursePhase.RestrictedData)
}

func validateWelcomeText(restrictedData meta.MetaData) error {
	value, exists := restrictedData[welcomeTextKey]
	if !exists || value == nil {
		return nil
	}

	text, isString := value.(string)
	if !isString {
		errorMessage := "welcomeText must be a string"
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}

	if len(text) > maxWelcomeTextBytes {
		errorMessage := fmt.Sprintf("welcomeText must not exceed %d bytes", maxWelcomeTextBytes)
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}

	return nil
}
