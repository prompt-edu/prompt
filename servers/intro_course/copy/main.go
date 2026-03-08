package copy

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
)

func InitCopyModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupCopyRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	CopyServiceSingleton = &CopyService{
		queries: queries,
		conn:    conn,
	}
}
