package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
)

func InitPrivacyModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	promptTypes.RegisterPrivacyDataExportEndpoint(routerGroup, PrivacyDataExportHandler, []string{})
}

func PrivacyDataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	return nil
}
