package privacy

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

type TeamsPrivacyService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var TeamsPrivacyServiceSingleton *TeamsPrivacyService

func GetTeamForCourseParticipationID(ctx context.Context, courseParticipationIDs []uuid.UUID) (any, error) {
	allocations, err := TeamsPrivacyServiceSingleton.queries.GetAllocationByCourseParticipationID(ctx, courseParticipationIDs)
  if err != nil {
    return nil, err
  }
  return allocations, nil
}

