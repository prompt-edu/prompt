package evaluations

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/feedbackItem"
)

func InitEvaluationModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupEvaluationRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	EvaluationServiceSingleton = &EvaluationService{
		queries: queries,
		conn:    conn,
	}

	// Initialize evaluation sub-modules
	evaluationCompletion.InitEvaluationCompletionModule(routerGroup, queries, conn)
	feedbackItem.InitFeedbackItemModule(routerGroup, queries, conn)
}
