package scoreLevel

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

func InitScoreLevelModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupScoreLevelRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	ScoreLevelServiceSingleton = &ScoreLevelService{
		queries: queries,
		conn:    conn,
	}
}
