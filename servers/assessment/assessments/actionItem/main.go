package actionItem

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

func InitActionItemModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupActionItemRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	ActionItemServiceSingleton = &ActionItemService{
		queries: queries,
		conn:    conn,
	}
}
