// Package config implements the SDK PhaseConfigHandler for the infrastructure setup phase.
package config

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

// Service provides phase-level configuration status.
//
// Upstream wiring (teams, team allocation) is owned by the phase configurator and
// surfaced by core, so this only reports phase-local readiness.
//
// It implements promptTypes.PhaseConfigHandler itself, so RegisterRoutes passes it
// straight to the SDK and there is no separate handler type to keep in sync.
type Service struct {
	queries *db.Queries
}

// NewService creates a Service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{queries: db.New(pool)}
}

// HandlePhaseConfig implements promptTypes.PhaseConfigHandler.
func (s *Service) HandlePhaseConfig(c *gin.Context) (map[string]bool, error) {
	coursePhaseID, err := uuid.Parse(c.Param("coursePhaseID"))
	if err != nil {
		return nil, err
	}
	return s.Status(c.Request.Context(), coursePhaseID)
}

// Status reports which of the phase's own configuration steps are done. A phase with
// no config row yet is unconfigured rather than an error; anything else the database
// reports is propagated, so a broken query surfaces as a 500 instead of a page that
// claims the lecturer never configured anything.
func (s *Service) Status(ctx context.Context, coursePhaseID uuid.UUID) (map[string]bool, error) {
	semesterTag := false
	cfg, err := s.queries.GetCoursePhaseConfig(ctx, coursePhaseID)
	switch {
	case err == nil:
		semesterTag = cfg.SemesterTag != ""
	case errors.Is(err, pgx.ErrNoRows):
	default:
		return nil, err
	}

	// Only providers holding credentials count: a copied phase carries provider rows
	// with empty credentials, which cannot provision anything.
	providers, err := s.queries.ListConfiguredProviderConfigs(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}
	resources, err := s.queries.ListResourceConfigs(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}

	return map[string]bool{
		"semesterTag":    semesterTag,
		"providerConfig": len(providers) > 0,
		"resourceConfig": len(resources) > 0,
	}, nil
}
