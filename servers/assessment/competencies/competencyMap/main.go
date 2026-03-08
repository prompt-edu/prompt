package competencyMap

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

func InitCompetencyMapModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupCompetencyMapRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	CompetencyMapServiceSingleton = &CompetencyMapService{
		queries: queries,
		conn:    conn,
	}
}
