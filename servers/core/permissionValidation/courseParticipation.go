package permissionValidation

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ValidationService) CheckCourseParticipationPermission(c *gin.Context, courseParticipationID uuid.UUID, allowedUsers ...string) (bool, error) {
	return checkCourseRole(c, func(ctx context.Context, id uuid.UUID) (courseAccess, error) {
		row, err := s.queries.GetPermissionStringByCourseParticipationID(ctx, id)
		return courseAccess(row), err
	}, courseParticipationID, allowedUsers...)
}
