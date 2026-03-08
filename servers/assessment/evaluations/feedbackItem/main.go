package feedbackItem

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

func InitFeedbackItemModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupFeedbackItemRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	FeedbackItemServiceSingleton = &FeedbackItemService{
		queries: queries,
		conn:    conn,
	}
}
