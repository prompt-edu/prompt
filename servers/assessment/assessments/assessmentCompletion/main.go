package assessmentCompletion

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

func InitAssessmentCompletionModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupAssessmentCompletionRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	AssessmentCompletionServiceSingleton = &AssessmentCompletionService{
		queries: queries,
		conn:    conn,
	}
}
