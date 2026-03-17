package allocation

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

func InitAllocationModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupAllocationRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	AllocationServiceSingleton = &AllocationService{
		queries: queries,
		conn:    conn,
	}
}
