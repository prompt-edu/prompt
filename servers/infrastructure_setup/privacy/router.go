package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

// RegisterRoutes mounts the SDK privacy module. Core addresses its endpoints as
// <base URL>/privacy/..., so the group must be the service's API root, and both
// endpoints bring their own middleware (an admin token for deletion).
func RegisterRoutes(rg *gin.RouterGroup, svc *PrivacyService) {
	promptTypes.RegisterPrivacyModule(rg, svc.DataExportHandler, svc.DataDeletionHandler, []string{})
}
