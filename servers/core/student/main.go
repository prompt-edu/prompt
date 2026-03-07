package student

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
)

func InitStudentModule(api *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {

	setupStudentRouter(api, keycloakTokenVerifier.KeycloakMiddleware, permissionValidation.CheckAccessControlByRole)
	StudentServiceSingleton = &StudentService{
		queries: queries,
		conn:    conn,
	}

	// possibly more setup tasks
}
