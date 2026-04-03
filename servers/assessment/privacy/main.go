package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type PrivacyService struct {
	Queries db.Queries
	Conn    *pgxpool.Pool
}

var PrivacyServiceSingleton PrivacyService

func InitPrivacyModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
	promptTypes.RegisterPrivacyDataExportEndpoint(routerGroup, PrivacyDataExportHandler, []string{})
	PrivacyServiceSingleton = PrivacyService{
		Queries: queries,
		Conn:    conn,
	}
}

func PrivacyDataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {

	exp.AddJSON("Assessments", "student/assessment.json", func() (any, error) {
		return PrivacyServiceSingleton.Queries.GetAllAssessmentsByCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})

	return nil
}
