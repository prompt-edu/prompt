package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
)

type CertificatePrivacyService struct {
	queries db.Queries
}

var singleton *CertificatePrivacyService

func InitPrivacyModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	promptTypes.RegisterPrivacyDataExportEndpoint(routerGroup, PrivacyDataExportHandler, []string{})
	singleton = &CertificatePrivacyService{queries: queries}
}

func PrivacyDataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	if subject.StudentID == uuid.Nil {
		return nil
	}

	exp.AddJSON("Certificate Downloads", "certificate_downloads.json", func() (any, error) {
		return singleton.queries.GetCertificateDownloadsByStudentID(c, subject.StudentID)
	})

	return nil
}
