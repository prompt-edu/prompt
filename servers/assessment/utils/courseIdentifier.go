package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	log "github.com/sirupsen/logrus"
)

// GetCourseIdentifierFromPhase retrieves the course identifier (name + semester_tag) from a course phase
func GetCourseIdentifierFromPhase(ctx context.Context, coursePhaseID uuid.UUID) (string, error) {
	coreURL := sdkUtils.GetCoreUrl()
	url := fmt.Sprintf("%s/api/course_phases/%s", coreURL, coursePhaseID.String())

	// Extract auth token from gin context, use empty string for test contexts
	var authToken string
	if ginCtx, ok := ctx.(*gin.Context); ok {
		authToken = ginCtx.GetHeader("Authorization")
	}
	// Allow empty auth token for testing scenarios (mock core service doesn't require auth)

	data, err := promptSDK.FetchJSON(url, authToken)
	if err != nil {
		log.WithError(err).Errorf("Failed to fetch course phase from core service: %s", url)
		return "", fmt.Errorf("failed to fetch course phase: %w", err)
	}

	var coursePhase struct {
		CourseID uuid.UUID `json:"courseID"`
	}
	if err := json.Unmarshal(data, &coursePhase); err != nil {
		log.WithError(err).Error("Failed to unmarshal course phase response")
		return "", fmt.Errorf("failed to unmarshal course phase: %w", err)
	}

	// Now fetch the course information
	courseURL := fmt.Sprintf("%s/api/courses/%s", coreURL, coursePhase.CourseID.String())
	courseData, err := promptSDK.FetchJSON(courseURL, authToken)
	if err != nil {
		log.WithError(err).Errorf("Failed to fetch course from core service: %s", courseURL)
		return "", fmt.Errorf("failed to fetch course: %w", err)
	}

	var course struct {
		Name        string `json:"name"`
		SemesterTag string `json:"semesterTag"`
	}
	if err := json.Unmarshal(courseData, &course); err != nil {
		log.WithError(err).Error("Failed to unmarshal course response")
		return "", fmt.Errorf("failed to unmarshal course: %w", err)
	}

	return fmt.Sprintf("%s%s", course.Name, course.SemesterTag), nil
}
