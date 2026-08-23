package permissionValidation

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func checkUserRole(c *gin.Context, courseIdentifier string, allowedUsers ...string) (bool, error) {
	// Inject the course identifier for later use
	c.Set("courseTokenIdentifier", courseIdentifier)

	// Extract user roles from context
	rolesVal, exists := c.Get("userRoles")
	if !exists {
		err := errors.New("user roles not found in context")
		c.IndentedJSON(500, gin.H{"error": err.Error()})
		return false, err
	}

	userRoles, ok := rolesVal.(map[string]bool)
	if !ok {
		err := errors.New("invalid roles format in context")
		c.IndentedJSON(500, gin.H{"error": err.Error()})
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

	c.IndentedJSON(403, gin.H{"error": "no matching permission found"})

	return false, nil // No matching role found
}
