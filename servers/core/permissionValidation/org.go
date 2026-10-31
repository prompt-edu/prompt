package permissionValidation

import (
	"errors"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *ValidationService) CheckOrgPermission(c *gin.Context, orgID uuid.UUID, allowedRoles ...string) (bool, error) {
	rolesVal, exists := c.Get("userRoles")
	if !exists {
		return false, errors.New("user roles not found in context")
	}
	userRoles, ok := rolesVal.(map[string]bool)
	if !ok {
		return false, errors.New("invalid roles format in context")
	}

	if slices.Contains(allowedRoles, PromptAdmin) && userRoles[PromptAdmin] {
		return true, nil
	}

	slug, err := s.queries.GetOrgSlugByID(c, orgID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	for _, role := range allowedRoles {
		if (role == OrgAdmin || role == OrgMember) && userRoles[OrgRoleName(slug, role)] {
			return true, nil
		}
	}
	return false, nil
}
