package privacy

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

type TeamsPrivacyService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var TeamsPrivacyServiceSingleton *TeamsPrivacyService

func PrivacyDataExportHandler(c *gin.Context, exp *utils.Export, subject sdkAuth.SubjectIdentifiers) error {
	exp.AddJSON("Team Allocation", "team_allocation.json", func() (any, error) {
		return getTeamForCourseParticipationIDs(c, subject.CourseParticipationIDs)
	})
	return nil
}

func getTeamForCourseParticipationIDs(ctx context.Context, courseParticipationIDs []uuid.UUID) (any, error) {
	return TeamsPrivacyServiceSingleton.queries.GetAllocationByCourseParticipationID(ctx, courseParticipationIDs)
}
