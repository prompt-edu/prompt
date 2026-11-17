package permissionValidation

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ValidationService) CheckCoursePermission(c *gin.Context, courseID uuid.UUID, allowedUsers ...string) (bool, error) {
	return checkCourseRole(c, s.queries.GetPermissionStringByCourseID, courseID, allowedUsers...)
}
