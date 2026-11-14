package permissionValidation

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ValidationService) CheckCourseParticipationPermission(c *gin.Context, courseParticipationID uuid.UUID, allowedUsers ...string) (bool, error) {
	return checkCourseRole(c, s.queries.GetPermissionStringByCourseParticipationID, courseParticipationID, allowedUsers...)
}
