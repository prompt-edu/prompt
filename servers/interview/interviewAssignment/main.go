package interviewAssignment

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
)

func InitInterviewAssignmentModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupInterviewAssignmentRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	InterviewAssignmentServiceSingleton = &InterviewAssignmentService{
		queries: queries,
		conn:    conn,
	}
}
