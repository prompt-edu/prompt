package permissionValidation

import (
	"errors"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
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

	org, err := s.queries.GetOrgByID(c, orgID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	issuedAt, err := TokenIssuedAt(c)
	if err != nil {
		return false, err
	}
	if !TokenPostdatesOrg(issuedAt, org.CreatedAt.Time) {
		return false, nil
	}

	for _, role := range allowedRoles {
		if (role == OrgAdmin || role == OrgMember) && userRoles[OrgRoleName(org.Slug, role)] {
			return true, nil
		}
	}
	return false, nil
}

// TokenIssuedAt returns the issue time of the caller's token, which the auth middleware stores.
func TokenIssuedAt(c *gin.Context) (time.Time, error) {
	issuedAtVal, exists := c.Get(keycloakTokenVerifier.CtxTokenIssuedAt)
	if !exists {
		return time.Time{}, errors.New("token issue time not found in context")
	}
	issuedAt, ok := issuedAtVal.(time.Time)
	if !ok {
		return time.Time{}, errors.New("invalid token issue time format in context")
	}
	return issuedAt, nil
}

// TokenPostdatesOrg reports whether a token's org roles may count for the org. A token issued
// before the org existed holds the roles of a deleted org with the same slug. The creation time is
// taken once Keycloak is provisioned, after such an org's roles are gone. The issue time has whole
// seconds, so a token from the second the org was created is rejected too.
func TokenPostdatesOrg(issuedAt, orgCreatedAt time.Time) bool {
	return issuedAt.After(orgCreatedAt)
}
