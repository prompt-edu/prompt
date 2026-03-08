package assessments

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/actionItem"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

func InitAssessmentModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupAssessmentRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	AssessmentServiceSingleton = &AssessmentService{
		queries: queries,
		conn:    conn,
	}

	// Initialize assessment sub-modules
	assessmentCompletion.InitAssessmentCompletionModule(routerGroup, queries, conn)
	actionItem.InitActionItemModule(routerGroup, queries, conn)
	scoreLevel.InitScoreLevelModule(routerGroup, queries, conn)
}
