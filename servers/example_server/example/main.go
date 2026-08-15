package example

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/example_server/db/sqlc"
)

func InitExampleModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupExampleRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	ExampleServiceSingleton = &ExampleService{
		queries: queries,
		conn:    conn,
	}
}
