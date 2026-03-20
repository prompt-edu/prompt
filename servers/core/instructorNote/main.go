package instructorNote

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
)

func InitInstructorNoteModule(api *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {

	setupInstructorNoteRouter(api, keycloakTokenVerifier.KeycloakMiddleware, permissionValidation.CheckAccessControlByRole)
	InstructorNoteServiceSingleton = &InstructorNoteService{
		queries: queries,
		conn:    conn,
	}

	// possibly more setup tasks
}

