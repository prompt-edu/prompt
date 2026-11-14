package permissionValidation

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ValidationService) CheckCoursePhasePermission(c *gin.Context, coursePhaseID uuid.UUID, allowedUsers ...string) (bool, error) {
	return checkCourseRole(c, s.queries.GetPermissionStringByCoursePhaseID, coursePhaseID, allowedUsers...)
}
