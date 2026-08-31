package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

func RegisterRoutes(routerGroup *gin.RouterGroup, service *TeamsPrivacyService) {
	promptTypes.RegisterPrivacyDataExportEndpoint(routerGroup, service.DataExportHandler, []string{})
	promptTypes.RegisterPrivacyDataDeletionEndpoint(routerGroup, service.DataDeletionHandler)
}
