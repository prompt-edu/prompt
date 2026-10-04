package permissionValidation

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// handleError is assumed to write a JSON error response and abort the context.
// Adjust or replace with your actual error handling function.
func handleAuthError(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied"})
}

type PermissionCheck func(c *gin.Context, id uuid.UUID, allowedRoles ...string) (bool, error)

// Reads an id from the URL and checks based on the URL the permission of the user
func CheckAccessControlByID(checkPermission PermissionCheck, idParamName string, allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract and parse the course UUID from route parameters.
		id, err := uuid.Parse(c.Param(idParamName))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check the user's permission for the given course with the provided roles.
		hasAccess, err := checkPermission(c, id, allowedRoles...)
		if err != nil {
			log.Error("Permission validation failed: ", err)
			handleAuthError(c)
			return
		}

		// If the user does not have access, abort the request.
		if !hasAccess {
			log.Debug("User does not have access. Required roles: ", allowedRoles, " provided roles: ", c.GetString("roles"))
			handleAuthError(c)
			return
		}

		// Otherwise, allow the request to proceed.
		c.Next()
	}
}

// Check for Endpoints without an ID (i.e. create sth)
func CheckAccessControlByRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rolesVal, exists := c.Get("userRoles")
		if !exists {
			err := errors.New("user roles not found in context")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		userRoles, ok := rolesVal.(map[string]bool)
		if !ok {
			err := errors.New("invalid roles format in context")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Attempt to find a match for any of the allowed roles
		roleAllowed := false
		for _, allowedRole := range allowedRoles {
			switch allowedRole {
			case PromptAdmin:
				if userRoles[PromptAdmin] {
					roleAllowed = true
				}
			case PromptLecturer:
				if userRoles[PromptLecturer] {
					roleAllowed = true
				}
			case CourseLecturer, CourseEditor, CourseStudent:
				// For these roles, we check if the user has any role that ends with the allowedRole value.
				for userRole := range userRoles {
					if strings.HasSuffix(userRole, allowedRole) {
						roleAllowed = true
						break
					}
				}
			default:
				log.Debug("Role unknown: ", allowedRole)

			}

			if roleAllowed {
				break
			}
		}

		if !roleAllowed {
			log.Debug("User does not have required roles. Required: ", allowedRoles, ", provided: ", userRoles)
			handleAuthError(c)
			return
		}

		// If we reach here, the user has the required role(s), so continue.
		c.Next()
	}
}
