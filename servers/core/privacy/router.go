package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt/servers/core/privacy/service"
)

type privacyHandler struct {
	service *service.PrivacyService
}

func RegisterRoutes(router *gin.RouterGroup, privacyService *service.PrivacyService, authMiddleware func() gin.HandlerFunc, permissionRoleMiddleware func(allowedRoles ...string) gin.HandlerFunc) {
	handler := &privacyHandler{service: privacyService}

	privacyRouter := router.Group("/privacy", authMiddleware())

	handler.registerExportRoutes(privacyRouter, permissionRoleMiddleware)
	handler.registerDeletionRoutes(privacyRouter, permissionRoleMiddleware)
}
