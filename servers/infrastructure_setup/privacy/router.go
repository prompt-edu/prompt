package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

// RegisterRoutes mounts the SDK privacy endpoints. Core addresses them as
// <base URL>/privacy/..., so the group must be the service's API root, and both
// endpoints bring their own middleware (an admin token for deletion).
func RegisterRoutes(rg *gin.RouterGroup, svc *PrivacyService) {
	promptTypes.RegisterPrivacyDataExportEndpoint(rg, svc.DataExportHandler, []string{})
	promptTypes.RegisterPrivacyDataDeletionEndpoint(rg, svc.DataDeletionHandler)
}
