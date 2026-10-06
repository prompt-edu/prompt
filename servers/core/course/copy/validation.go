package copy

import (
	"errors"

	"github.com/prompt-edu/prompt/servers/core/course/copy/courseCopyDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	log "github.com/sirupsen/logrus"
)

func validateCopyCourseRequest(c courseCopyDTO.CopyCourseRequest) error {
	switch c.CourseType {
	case "", db.CourseTypeLecture, db.CourseTypeSeminar, db.CourseTypePracticalcourse:
		return nil
	default:
		errorMessage := "invalid course type"
		log.Error(errorMessage)
		return errors.New(errorMessage)
	}
}
