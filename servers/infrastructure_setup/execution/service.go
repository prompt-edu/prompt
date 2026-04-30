package execution

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

// Service handles resource instance lifecycle.
type Service struct {
	queries *db.Queries
	pool    *pgxpool.Pool
	worker  *Worker
}

// NewService creates a Service with a background worker.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		queries: db.New(pool),
		pool:    pool,
		worker:  NewWorker(pool),
	}
}

// TriggerExecution creates pending resource instances for all resource configs in a
// course phase and then starts the async worker.
func (s *Service) TriggerExecution(ctx context.Context, coursePhaseID uuid.UUID) error {
	configs, err := s.queries.ListResourceConfigs(ctx, coursePhaseID)
	if err != nil {
		return err
	}

	for _, cfg := range configs {
		if err := s.createInstancesForConfig(ctx, coursePhaseID, cfg); err != nil {
			return err
		}
	}

	// Spawn the async worker AFTER all DB writes are committed.
	s.worker.RunPendingInstances(coursePhaseID)
	return nil
}

func (s *Service) createInstancesForConfig(ctx context.Context, coursePhaseID uuid.UUID, cfg db.ResourceConfig) error {
	switch cfg.Scope {
	case db.ResourceScopePerTeam:
		// TODO: fetch team IDs from core service and create one instance per team.
		// For now create a single placeholder instance.
		_, err := s.queries.CreateResourceInstance(ctx, db.CreateResourceInstanceParams{
			ResourceConfigID: cfg.ID,
			CoursePhaseID:    coursePhaseID,
		})
		return err
	case db.ResourceScopePerStudent:
		// TODO: fetch student IDs from core service and create one instance per student.
		_, err := s.queries.CreateResourceInstance(ctx, db.CreateResourceInstanceParams{
			ResourceConfigID: cfg.ID,
			CoursePhaseID:    coursePhaseID,
		})
		return err
	}
	return nil
}

// ListInstances returns all resource instances for a course phase.
func (s *Service) ListInstances(ctx context.Context, coursePhaseID uuid.UUID) ([]db.ResourceInstance, error) {
	return s.queries.ListResourceInstances(ctx, coursePhaseID)
}

// RetryFailedInstance resets a failed instance back to pending and triggers execution.
func (s *Service) RetryFailedInstance(ctx context.Context, coursePhaseID, instanceID uuid.UUID) error {
	if err := s.queries.ResetFailedInstanceToPending(ctx, instanceID); err != nil {
		return err
	}
	s.worker.RunPendingInstances(coursePhaseID)
	return nil
}

// DeleteInstance removes a resource instance.
func (s *Service) DeleteInstance(ctx context.Context, instanceID uuid.UUID) error {
	return s.queries.DeleteResourceInstance(ctx, instanceID)
}
