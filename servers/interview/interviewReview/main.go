package interviewReview

import (
	"errors"

	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
)

var errScoreOutOfRange = errors.New("score must be between 1 and 5")

func InitInterviewReviewModule(routerGroup *gin.RouterGroup, queries db.Queries) {
	setupInterviewReviewRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	InterviewReviewServiceSingleton = &InterviewReviewService{
		queries: queries,
	}
}
