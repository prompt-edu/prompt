package permissionValidation

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

var errNotFound = errors.New("resource not found")

// CtxCourseOrgAdmin is true when the caller holds the Admin role of the checked course's org.
const CtxCourseOrgAdmin = "courseOrgAdmin"

// CtxCourseAccessViaOrg is true when the caller passed the course check only through a role in the
// course's org, not through a global or course role.
const CtxCourseAccessViaOrg = "courseAccessViaOrg"

// courseAccess is what a course check needs to know about a course. The GetPermissionStringBy*
// queries return exactly these columns, so their rows convert to it.
type courseAccess db.GetPermissionStringByCourseIDRow

type courseAccessLookup func(ctx context.Context, id uuid.UUID) (courseAccess, error)

func checkCourseRole(c *gin.Context, lookup courseAccessLookup, id uuid.UUID, allowedUsers ...string) (bool, error) {
	course, err := lookup(c, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, errNotFound
	}
	if err != nil {
		return false, err
	}
	return checkUserRole(c, course, allowedUsers...)
}

func userRolesFromContext(c *gin.Context) (map[string]bool, error) {
	rolesVal, exists := c.Get("userRoles")
	if !exists {
		return nil, errors.New("user roles not found in context")
	}

	userRoles, ok := rolesVal.(map[string]bool)
	if !ok {
		return nil, errors.New("invalid roles format in context")
	}
	return userRoles, nil
}

// checkUserRole grants access when the caller holds one of the allowed roles for the course. Two
// roles of the course's org count as well: its Admins pass wherever CourseLecturer is allowed, since
// they are Lecturer-equivalent on all of the org's courses, and its Members pass where OrgMember is
// allowed, but only for a template.
func checkUserRole(c *gin.Context, course courseAccess, allowedUsers ...string) (bool, error) {
	// Inject the course identifier for later use
	c.Set("courseTokenIdentifier", course.CourseIdentifier)

	userRoles, err := userRolesFromContext(c)
	if err != nil {
		return false, err
	}

	orgAdmin, orgMember, err := courseOrgRoles(c, course, userRoles)
	if err != nil {
		return false, err
	}
	c.Set(CtxCourseOrgAdmin, orgAdmin)

	for _, role := range allowedUsers {
		switch role {
		case PromptAdmin, PromptLecturer:
			if userRoles[role] {
				return true, nil
			}
		case OrgMember:
			if orgMember && course.Template {
				c.Set(CtxCourseAccessViaOrg, true)
				return true, nil
			}
		default:
			if userRoles[CourseRoleName(course.CourseIdentifier, role)] {
				return true, nil
			}
			if role == CourseLecturer && orgAdmin {
				c.Set(CtxCourseAccessViaOrg, true)
				return true, nil
			}
		}
	}

	return false, nil // No matching role found
}

// courseOrgRoles reports whether the caller holds the Admin and the Member role of the course's own
// org; nesting grants nothing, so a parent org's roles never count. Like CheckOrgPermission, it
// ignores a token issued before the org was created, which holds the roles of a deleted org with the
// same slug.
func courseOrgRoles(c *gin.Context, course courseAccess, userRoles map[string]bool) (admin, member bool, err error) {
	if !course.OrgSlug.Valid {
		return false, false, nil
	}
	admin = userRoles[OrgRoleName(course.OrgSlug.String, OrgAdmin)]
	member = userRoles[OrgRoleName(course.OrgSlug.String, OrgMember)]
	if !admin && !member {
		return false, false, nil
	}

	issuedAt, err := TokenIssuedAt(c)
	if err != nil {
		return false, false, err
	}
	if !TokenPostdatesOrg(issuedAt, course.OrgCreatedAt.Time) {
		return false, false, nil
	}
	return admin, member, nil
}
