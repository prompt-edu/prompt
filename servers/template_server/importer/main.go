package importer

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/template_server/db/sqlc"
)

func InitImporterModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupImporterRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	ImporterServiceSingleton = &ImporterService{
		queries: queries,
		conn:    conn,
	}
}
