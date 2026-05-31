package privacy

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
)

type PrivacyService struct {
	Queries db.Queries
	Conn    *pgxpool.Pool
}

var PrivacyServiceSingleton *PrivacyService

func PrivacyDataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	exp.AddJSON("Self Team Allocation", "self_team_allocation.json", func() (any, error) {
		return PrivacyServiceSingleton.Queries.GetAssignmentsByParticipationIDs(c, subject.CourseParticipationIDs)
	})
	return nil
}
