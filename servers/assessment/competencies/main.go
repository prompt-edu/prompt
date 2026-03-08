package competencies

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/competencies/competencyMap"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

func InitCompetencyModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupCompetencyRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	CompetencyServiceSingleton = &CompetencyService{
		queries: queries,
		conn:    conn,
	}

	// Initialize competency map sub-module
	competencyMap.InitCompetencyMapModule(routerGroup, queries, conn)
}
