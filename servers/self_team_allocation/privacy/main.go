package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
)

func InitPrivacyModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	PrivacyServiceSingleton = &PrivacyService{
		Queries: queries,
	}
	promptTypes.RegisterPrivacyDataExportEndpoint(routerGroup, PrivacyDataExportHandler, []string{})
}
