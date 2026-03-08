package seatPlan

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
)

func InitSeatPlanModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	setupSeatPlanRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	SeatPlanServiceSingleton = &SeatPlanService{
		queries: queries,
		conn:    conn,
	}
}
