package interviewSlot

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
)

func InitInterviewSlotModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupInterviewSlotRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	InterviewSlotServiceSingleton = &InterviewSlotService{
		queries: queries,
		conn:    conn,
	}
}
