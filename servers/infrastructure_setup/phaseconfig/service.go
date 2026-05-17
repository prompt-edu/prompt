package phaseconfig

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

// Service handles infrastructure setup phase configuration.
type Service struct {
	queries *db.Queries
}

// NewService creates a Service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{queries: db.New(pool)}
}

// Get returns phase configuration or an empty default when it has not been saved.
func (s *Service) Get(ctx context.Context, coursePhaseID uuid.UUID) (Response, error) {
	cfg, err := s.queries.GetCoursePhaseConfig(ctx, coursePhaseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Response{CoursePhaseID: coursePhaseID}, nil
	}
	if err != nil {
		return Response{}, err
	}
	return toResponse(cfg), nil
}

// Upsert creates or updates phase configuration.
func (s *Service) Upsert(ctx context.Context, coursePhaseID uuid.UUID, req UpsertRequest) (Response, error) {
	cfg, err := s.queries.UpsertCoursePhaseConfig(ctx, db.UpsertCoursePhaseConfigParams{
		CoursePhaseID:              coursePhaseID,
		TeamSourceCoursePhaseID:    req.TeamSourceCoursePhaseID,
		StudentSourceCoursePhaseID: req.StudentSourceCoursePhaseID,
		SemesterTag:                req.SemesterTag,
	})
	if err != nil {
		return Response{}, err
	}
	return toResponse(cfg), nil
}

func toResponse(cfg db.CoursePhaseConfig) Response {
	return Response{
		CoursePhaseID:              cfg.CoursePhaseID,
		TeamSourceCoursePhaseID:    cfg.TeamSourceCoursePhaseID,
		StudentSourceCoursePhaseID: cfg.StudentSourceCoursePhaseID,
		SemesterTag:                cfg.SemesterTag,
	}
}
