package developerProfile

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
)

func InitDeveloperProfileModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupDeveloperProfileRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	DeveloperProfileServiceSingleton = &DeveloperProfileService{
		queries: queries,
		conn:    conn,
	}
}
