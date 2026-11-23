package permissionValidation

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ValidationService) CheckCoursePhasePermission(c *gin.Context, coursePhaseID uuid.UUID, allowedUsers ...string) (bool, error) {
	return checkCourseRole(c, func(ctx context.Context, id uuid.UUID) (courseAccess, error) {
		row, err := s.queries.GetPermissionStringByCoursePhaseID(ctx, id)
		return courseAccess(row), err
	}, coursePhaseID, allowedUsers...)
}
