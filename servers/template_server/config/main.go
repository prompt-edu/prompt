package config

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/template_server/db/sqlc"
)

func InitConfigModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupConfigRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	ConfigServiceSingleton = &ConfigService{
		queries: queries,
		conn:    conn,
	}
}
