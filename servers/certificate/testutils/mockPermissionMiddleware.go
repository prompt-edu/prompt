package testutils

import "github.com/gin-gonic/gin"

func MockPermissionMiddleware(authRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
