package permissionValidation

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
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

// OwnOrg is an org whose role the caller's token holds, with that role.
type OwnOrg struct {
	db.Org
	Role string
}

// OwnOrgs returns the orgs whose roles the caller's token holds, each with the caller's role; an
// Admin role wins over a Member role of the same org. As in CheckOrgPermission, an org created after
// the token was issued is left out: the token's roles for its slug belong to a deleted org.
func OwnOrgs(ctx context.Context, queries db.Queries, userRoles map[string]bool, tokenIssuedAt time.Time) ([]OwnOrg, error) {
	roleBySlug := map[string]string{}
	for role, held := range userRoles {
		slug, orgRole, ok := ParseOrgRole(role)
		if held && ok && roleBySlug[slug] != OrgAdmin {
			roleBySlug[slug] = orgRole
		}
	}

	ownOrgs := []OwnOrg{}
	if len(roleBySlug) == 0 {
		return ownOrgs, nil
	}
	slugs := slices.Collect(maps.Keys(roleBySlug))
	orgs, err := queries.GetOrgsBySlugs(ctx, slugs)
	if err != nil {
		return nil, fmt.Errorf("failed to get the orgs with the slugs %v: %w", slugs, err)
	}
	for _, org := range orgs {
		if TokenPostdatesOrg(tokenIssuedAt, org.CreatedAt.Time) {
			ownOrgs = append(ownOrgs, OwnOrg{Org: org, Role: roleBySlug[org.Slug]})
		}
	}
	return ownOrgs, nil
}

// OwnOrgIDs returns the IDs of the caller's orgs in which they hold one of the given roles.
func OwnOrgIDs(ownOrgs []OwnOrg, roles ...string) []uuid.UUID {
	ids := []uuid.UUID{}
	for _, org := range ownOrgs {
		if slices.Contains(roles, org.Role) {
			ids = append(ids, org.ID)
		}
	}
	return ids
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
