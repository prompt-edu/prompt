package interviewReview

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
)

var errScoreOutOfRange = errors.New("score must be between 1 and 5")

func InitInterviewReviewModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupInterviewReviewRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	InterviewReviewServiceSingleton = &InterviewReviewService{
		queries: queries,
		conn:    conn,
	}
}
