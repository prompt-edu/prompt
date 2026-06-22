package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

func InitPrivacyModule(routerGroup *gin.RouterGroup, queries db.Queries) {
	TeamsPrivacyServiceSingleton = &TeamsPrivacyService{
		queries: queries,
	}
	promptTypes.RegisterPrivacyDataExportEndpoint(routerGroup, PrivacyDataExportHandler, []string{})
}
