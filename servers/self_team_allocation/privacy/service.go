package privacy

import (
	"github.com/gin-gonic/gin"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
)

type SelfTeamAllocationPrivacyService struct {
	queries db.Queries
}

var singleton *SelfTeamAllocationPrivacyService

func PrivacyDataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	exp.AddJSON("Self Team Allocation", "self_team_allocation.json", func() (any, error) {
		return singleton.queries.GetAssignmentsByParticipationIDs(c, subject.CourseParticipationIDs)
	})
	return nil
}
