package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

func InitPrivacyModule(routerGroup *gin.RouterGroup, queries db.Queries, conn *pgxpool.Pool) {
  allowedHosts := []string{}
  promptTypes.RegisterPrivacyDataExportEndpoint(routerGroup, PrivacyDataExportHandler, allowedHosts)

	TeamsPrivacyServiceSingleton = &TeamsPrivacyService{
		queries: queries,
		conn:    conn,
	}
}

func PrivacyDataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {

  exp.AddJSON("Team Allocation", "team_allocation.json", func() (any, error) {
    return GetTeamForCourseParticipationID(c, subject.CourseParticipationIDs)
  })

  return nil
}
