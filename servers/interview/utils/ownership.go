package utils

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ErrCourseParticipationIDNotFound = errors.New("course participation ID not found")
var ErrInvalidCourseParticipationIDFormat = errors.New("invalid course participation ID format")
var ErrInvalidCourseParticipationID = errors.New("invalid course participation ID")
var ErrOwnershipValidation = errors.New("you can only manage your own interview assignments")

func GetUserCourseParticipationIDErrorStatus(err error) int {
	switch err {
	case ErrCourseParticipationIDNotFound:
		return http.StatusUnauthorized
	case ErrInvalidCourseParticipationIDFormat, ErrInvalidCourseParticipationID:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func GetUserCourseParticipationID(c *gin.Context) (uuid.UUID, error) {
	userCourseParticipationID, exists := c.Get("courseParticipationID")
	if !exists {
		return uuid.UUID{}, ErrCourseParticipationIDNotFound
	}

	userCourseParticipationUUID, ok := userCourseParticipationID.(uuid.UUID)
	if !ok {
		userCourseParticipationStr, ok := userCourseParticipationID.(string)
		if !ok {
			return uuid.UUID{}, ErrInvalidCourseParticipationIDFormat
		}
		var err error
		userCourseParticipationUUID, err = uuid.Parse(userCourseParticipationStr)
		if err != nil {
			return uuid.UUID{}, ErrInvalidCourseParticipationID
		}
	}

	return userCourseParticipationUUID, nil
}

func ValidateStudentOwnership(c *gin.Context, authorCourseParticipationID uuid.UUID) (int, error) {
	userCourseParticipationUUID, err := GetUserCourseParticipationID(c)
	if err != nil {
		return GetUserCourseParticipationIDErrorStatus(err), err
	}

	if authorCourseParticipationID != userCourseParticipationUUID {
		return http.StatusForbidden, ErrOwnershipValidation
	}

	return http.StatusOK, nil
}
