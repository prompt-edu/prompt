package permissionValidation

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ValidationService) CheckCoursePermission(c *gin.Context, courseID uuid.UUID, allowedUsers ...string) (bool, error) {
	return checkCourseRole(c, func(ctx context.Context, id uuid.UUID) (courseAccess, error) {
		row, err := s.queries.GetPermissionStringByCourseID(ctx, id)
		return courseAccess(row), err
	}, courseID, allowedUsers...)
}
