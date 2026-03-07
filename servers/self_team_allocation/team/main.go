package teams

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
)

func InitTeamModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupTeamRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	TeamsServiceSingleton = &TeamsService{
		queries: queries,
		conn:    conn,
	}
	AssignmentServiceSingleton = &AssignmentService{
		queries: queries,
		conn:    conn,
	}
}
