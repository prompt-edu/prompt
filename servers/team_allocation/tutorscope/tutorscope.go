package tutorscope

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

// Resolver adapts the service's sqlc queries to the SDK's TutorTeamResolver so
// promptSDK.TutorScopingMiddleware can scope tutor access to their assigned team.
type Resolver struct {
	q db.Queries
}

func NewResolver(q db.Queries) Resolver {
	return Resolver{q: q}
}

func (r Resolver) ResolveTutorTeam(ctx context.Context, coursePhaseID uuid.UUID, universityLogin string) (uuid.UUID, error) {
	return r.q.GetTutorTeamByUniversityLogin(ctx, db.GetTutorTeamByUniversityLoginParams{
		CoursePhaseID:   coursePhaseID,
		UniversityLogin: pgtype.Text{String: universityLogin, Valid: true},
	})
}
