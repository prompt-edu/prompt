package permissionValidation

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var errNotFound = errors.New("resource not found")

type courseIdentifierLookup func(ctx context.Context, id uuid.UUID) (interface{}, error)

func checkCourseRole(c *gin.Context, lookup courseIdentifierLookup, id uuid.UUID, allowedUsers ...string) (bool, error) {
	identifier, err := lookup(c, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, errNotFound
	}
	if err != nil {
		return false, err
	}
	courseIdentifier, ok := identifier.(string)
	if !ok {
		return false, fmt.Errorf("expected string but got %T", identifier)
	}

	return checkUserRole(c, courseIdentifier, allowedUsers...)
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

func checkUserRole(c *gin.Context, courseIdentifier string, allowedUsers ...string) (bool, error) {
	// Inject the course identifier for later use
	c.Set("courseTokenIdentifier", courseIdentifier)

	// Extract user roles from context
	userRoles, err := userRolesFromContext(c)
	if err != nil {
		return false, err
	}

	// Generate the desired role keys based on input
	for _, role := range allowedUsers {
		var desiredRole string
		switch role {
		case PromptAdmin:
			desiredRole = PromptAdmin
		case PromptLecturer:
			desiredRole = PromptLecturer
		default:
			desiredRole = CourseRoleName(courseIdentifier, role)
		}

		if userRoles[desiredRole] {
			return true, nil // Found at least one matching role
		}
	}

	return false, nil // No matching role found
}
