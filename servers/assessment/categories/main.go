package categories

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

func InitCategoryModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupCategoryRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	CategoryServiceSingleton = &CategoryService{
		queries: queries,
		conn:    conn,
	}
}
