package privacy

import (
	"github.com/gin-gonic/gin"
)

func setupPrivacyRouter(router *gin.RouterGroup, authMiddleware func() gin.HandlerFunc, permissionRoleMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	privacyRouter := router.Group("/privacy", authMiddleware())

	setupPrivacyExportRouter(privacyRouter, authMiddleware, permissionRoleMiddleware)
	setupPrivacyDeletionRouter(privacyRouter, authMiddleware, permissionRoleMiddleware)
}
