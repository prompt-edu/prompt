package course

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/course/courseDTO"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

const (
	maxShortDescriptionLength = 255
	maxLongDescriptionLength  = 5000
)

// ErrInvalidCourseIdentifier is returned when a new course's name or semester tag is empty or
// contains one of the characters "-", "/", "%", "?" or "#" or a control character.
var ErrInvalidCourseIdentifier = errors.New("invalid course name or semester tag")

// ValidateCourseIdentifier checks the name and semester tag of a new course, which join to the
// identifier "<semester>-<name>" that names the course's Keycloak groups and roles:
//   - A "-" would let two courses share an identifier, such as "ws24" with "ios-lab" and "ws24-ios"
//     with "lab". Without one, the identifier splits back into its semester tag and name in exactly
//     one way, and the database keeps (name, semester_tag) unique. Courses created before this rule
//     can contain a "-", but their identifiers then hold more than one and differ from a new one.
//   - Keycloak must be able to address the groups and roles of the identifier (see
//     permissionValidation.IsAddressableInKeycloak): "ws24" with "ios/Lecturer" would resolve to the
//     Lecturer group of the course "ws24-ios", and "lab#1" would look up "lab".
func ValidateCourseIdentifier(name, semesterTag string) error {
	for _, part := range []struct{ label, value string }{{"name", name}, {"semester tag", semesterTag}} {
		if part.value == "" {
			return fmt.Errorf("%w: the %s is required", ErrInvalidCourseIdentifier, part.label)
		}
		if strings.Contains(part.value, "-") || !permissionValidation.IsAddressableInKeycloak(part.value) {
			return fmt.Errorf(`%w: the %s must not contain "-", "/", "%%", "?", "#" or control characters`, ErrInvalidCourseIdentifier, part.label)
		}
	}
	return nil
}

func validateCreateCourse(c courseDTO.CreateCourse) error {
	if c.Name == "" {
		errorMessage := "course name is required"
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}
	if !c.StartDate.Valid {
		errorMessage := "start date is required"
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}
	if !c.EndDate.Valid {
		errorMessage := "end date is required"
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}
	if !c.Template {
		if c.StartDate.Time.After(c.EndDate.Time) {
			errorMessage := "start date must be before end date"
			log.Error(errorMessage)
			return errors.New(errorMessage)
		}
	}
	if !c.SemesterTag.Valid || c.SemesterTag.String == "" {
		errorMessage := "semester tag is required"
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}
	if err := ValidateCourseIdentifier(c.Name, c.SemesterTag.String); err != nil {
		log.Error(err)
		return err
	}
	if c.CourseType == "" {
		errorMessage := "course type is required"
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}

	if c.ShortDescription.Valid && utf8.RuneCountInString(c.ShortDescription.String) > maxShortDescriptionLength {
		errorMessage := "short description must be 255 characters or fewer"
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}

	if c.LongDescription.Valid && utf8.RuneCountInString(c.LongDescription.String) > maxLongDescriptionLength {
		errorMessage := "long description must be 5000 characters or fewer"
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}

	return nil
}

func (s *CourseService) validateMetaDataGraph(ctx context.Context, courseID uuid.UUID, newGraph []courseDTO.MetaDataGraphItem) error {
	// for each check if the course phase really belongs to this course
	uniqueCoursePhaseIDs := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]bool)

	// Loop through each graph item.
	for _, graphItem := range newGraph {
		// Check and add the FromCoursePhaseID if it's not already added.
		if !seen[graphItem.FromCoursePhaseID] {
			uniqueCoursePhaseIDs = append(uniqueCoursePhaseIDs, graphItem.FromCoursePhaseID)
			seen[graphItem.FromCoursePhaseID] = true
		}

		// Check and add the ToCoursePhaseID if it's not already added.
		if !seen[graphItem.ToCoursePhaseID] {
			uniqueCoursePhaseIDs = append(uniqueCoursePhaseIDs, graphItem.ToCoursePhaseID)
			seen[graphItem.ToCoursePhaseID] = true
		}
	}

	// Check if they all belong to this course and exist
	valid, err := s.coursePhases.CheckCoursePhasesBelongToCourse(ctx, courseID, uniqueCoursePhaseIDs)
	if err != nil {
		return err
	}

	if !valid {
		errorMessage := "not all course phases belong to this course"
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}

	return nil
}

func validateUpdateCourseData(c courseDTO.UpdateCourseData) error {
	if c.StartDate.Valid || c.EndDate.Valid {
		// either update both or none
		if !c.StartDate.Valid || !c.EndDate.Valid {
			errorMessage := "both start and end date must be updated"
			log.Error(errorMessage)
			return errors.New(errorMessage)
		}

		if c.StartDate.Time.After(c.EndDate.Time) {
			errorMessage := "start date must be before end date"
			log.Error(errorMessage)
			return errors.New(errorMessage)
		}
	}

	if c.ShortDescription.Valid && utf8.RuneCountInString(c.ShortDescription.String) > maxShortDescriptionLength {
		errorMessage := "short description must be 255 characters or fewer"
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}

	if c.LongDescription.Valid && utf8.RuneCountInString(c.LongDescription.String) > maxLongDescriptionLength {
		errorMessage := "long description must be 5000 characters or fewer"
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}
	return nil
}
