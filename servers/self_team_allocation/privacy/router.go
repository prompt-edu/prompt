package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
)

func RegisterRoutes(routerGroup *gin.RouterGroup, service *PrivacyService) {
	promptTypes.RegisterPrivacyModule(routerGroup, service.DataExportHandler, service.DataDeletionHandler, []string{})
}
