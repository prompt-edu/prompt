package evaluationCompletion

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

func InitEvaluationCompletionModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupEvaluationCompletionRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	EvaluationCompletionServiceSingleton = &EvaluationCompletionService{
		queries: queries,
		conn:    conn,
	}
}
