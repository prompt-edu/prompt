package assessmentSchemas

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

func InitAssessmentSchemaModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	SetupAssessmentSchemaRouter(routerGroup, promptSDK.AuthenticationMiddleware)
	AssessmentSchemaServiceSingleton = &AssessmentSchemaService{
		queries: queries,
		conn:    conn,
	}
}
