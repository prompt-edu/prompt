package privacy

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/privacy/service"
)

func InitPrivacyModule(api *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {

	setupPrivacyRouter(api, keycloakTokenVerifier.KeycloakMiddleware, permissionValidation.CheckAccessControlByRole)

	service.InitPrivacyService(queries, conn)

	service.StartExportDeletionRoutine(context.Background())

}
