package copy

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt/servers/core/course/copy/courseCopyDTO"
	"github.com/prompt-edu/prompt/servers/core/course/courseDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type CourseCopyService struct {
	queries db.Queries
	conn    *pgxpool.Pool
	// use dependency injection for keycloak to allow mocking
	createCourseGroupsAndRoles func(ctx context.Context, courseName, iterationName, userID string) error
}

var CourseCopyServiceSingleton *CourseCopyService

func CheckAllCoursePhasesCopyable(c *gin.Context, sourceCourseID uuid.UUID) ([]string, error) {
	missing, err := checkAllCoursePhasesCopyable(c, sourceCourseID)
	if err != nil {
		return nil, fmt.Errorf("failed to check course phases copyable: %w", err)
	}
	return missing, nil
}

func CopyCourse(c *gin.Context, sourceCourseID uuid.UUID, courseVariables courseCopyDTO.CopyCourseRequest, requesterID string) (courseDTO.Course, error) {
	course, err := copyCourseInternal(c, sourceCourseID, courseVariables, requesterID)
	if err != nil {
		return courseDTO.Course{}, fmt.Errorf("course copy failed: %w", err)
	}
	return course, nil
}
