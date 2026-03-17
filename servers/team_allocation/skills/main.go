package skills

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

func InitSkillModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupSkillRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	SkillsServiceSingleton = &SkillsService{
		queries: queries,
		conn:    conn,
	}
}
