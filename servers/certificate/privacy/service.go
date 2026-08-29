package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
)

type PrivacyService struct {
	queries db.Queries
}

func NewPrivacyService(queries db.Queries) *PrivacyService {
	return &PrivacyService{
		queries: queries,
	}
}

func (s *PrivacyService) DataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	if subject.StudentID == uuid.Nil {
		return nil
	}

	exp.AddJSON("Certificate Downloads", "certificate_downloads.json", func() (any, error) {
		return s.queries.GetCertificateDownloadsByStudentID(c, subject.StudentID)
	})

	return nil
}

func (s *PrivacyService) DataDeletionHandler(c *gin.Context, subject sdkAuth.SubjectIdentifiers) error {
	if subject.StudentID == uuid.Nil {
		return nil
	}

	return s.queries.DeleteCertificateDownloadsByStudentID(c, subject.StudentID)
}
