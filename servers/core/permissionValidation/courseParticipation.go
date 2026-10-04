package permissionValidation

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ValidationService) CheckCourseParticipationPermission(c *gin.Context, coursePhaseID uuid.UUID, allowedUsers ...string) (bool, error) {
	courseIdentifier, err := s.courseIdentifierStringFromCourseParticipationID(c, coursePhaseID)
	if err != nil {
		c.IndentedJSON(500, gin.H{"error": err.Error()})
		return false, err
	}

	return checkUserRole(c, courseIdentifier, allowedUsers...)
}

func (s *ValidationService) courseIdentifierStringFromCourseParticipationID(ctx context.Context, uuid uuid.UUID) (string, error) {
	identifier, err := s.queries.GetPermissionStringByCourseParticipationID(ctx, uuid)
	if err != nil {
		return "", err
	}
	value, ok := identifier.(string)
	if !ok {
		return "", fmt.Errorf("expected string but got %T", identifier)
	}
	return value, nil

}
