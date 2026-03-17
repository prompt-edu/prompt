package course

import (
	"context"
	"errors"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/course/courseDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase"
	log "github.com/sirupsen/logrus"
)

const (
	maxShortDescriptionLength = 255
	maxLongDescriptionLength  = 5000
)

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

func validateUpdateCourseOrder(ctx context.Context, courseID uuid.UUID, c []courseDTO.CoursePhaseGraph) error {
	// for each course phase check if the course id is the same
	for _, graphItem := range c {
		coursePhase, err := coursePhase.GetCoursePhaseByID(ctx, graphItem.ToCoursePhaseID)
		if err != nil {
			return err
		}
		if courseID != coursePhase.CourseID {
			errorMessage := "course id must be the same for all course phases"
			log.Error(errorMessage)
			return errors.New(errorMessage)
		}
	}

	return nil
}

func validateMetaDataGraph(ctx context.Context, courseID uuid.UUID, newGraph []courseDTO.MetaDataGraphItem) error {
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
	valid, err := coursePhase.CheckCoursePhasesBelongToCourse(ctx, courseID, uniqueCoursePhaseIDs)
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
