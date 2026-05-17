package execution

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

// Service handles resource instance lifecycle.
type Service struct {
	queries  *db.Queries
	pool     *pgxpool.Pool
	worker   *Worker
	resolver TargetResolver
}

// NewService creates a Service with a background worker.
func NewService(pool *pgxpool.Pool) *Service {
	queries := db.New(pool)
	return &Service{
		queries:  queries,
		pool:     pool,
		resolver: NewCoreTargetResolver(queries),
		worker:   NewWorker(pool),
	}
}

// NewServiceWithResolver creates a Service with an injected resolver for tests.
func NewServiceWithResolver(pool *pgxpool.Pool, resolver TargetResolver) *Service {
	queries := db.New(pool)
	return &Service{
		queries:  queries,
		pool:     pool,
		resolver: resolver,
		worker:   NewWorkerWithResolver(pool, resolver),
	}
}

// TriggerExecution creates pending resource instances for all resource configs in a
// course phase and then starts the async worker.
func (s *Service) TriggerExecution(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) error {
	configs, err := s.queries.ListResourceConfigs(ctx, coursePhaseID)
	if err != nil {
		return err
	}

	for _, cfg := range configs {
		if err := s.createInstancesForConfig(ctx, authHeader, coursePhaseID, cfg); err != nil {
			return err
		}
	}

	// Spawn the async worker AFTER all DB writes are committed.
	s.worker.RunPendingInstances(authHeader, coursePhaseID)
	return nil
}

func (s *Service) createInstancesForConfig(ctx context.Context, authHeader string, coursePhaseID uuid.UUID, cfg db.ResourceConfig) error {
	targets, err := s.resolver.ResolveTargets(ctx, authHeader, coursePhaseID, cfg.Scope)
	if err != nil {
		return err
	}

	for _, target := range targets {
		if _, err := s.queries.CreateResourceInstance(ctx, createResourceInstanceParams(cfg, coursePhaseID, target)); err != nil {
			return err
		}
	}

	return nil
}

func createResourceInstanceParams(cfg db.ResourceConfig, coursePhaseID uuid.UUID, target ProvisioningTarget) db.CreateResourceInstanceParams {
	return db.CreateResourceInstanceParams{
		ResourceConfigID:      cfg.ID,
		CoursePhaseID:         coursePhaseID,
		TeamID:                target.TeamID,
		CourseParticipationID: target.CourseParticipationID,
	}
}

// ListInstances returns all resource instances for a course phase.
func (s *Service) ListInstances(ctx context.Context, coursePhaseID uuid.UUID) ([]db.ResourceInstance, error) {
	return s.queries.ListResourceInstances(ctx, coursePhaseID)
}

// RetryFailedInstance resets a failed instance back to pending and triggers execution.
func (s *Service) RetryFailedInstance(ctx context.Context, coursePhaseID, instanceID uuid.UUID) error {
	return s.RetryFailedInstanceWithAuth(ctx, "", coursePhaseID, instanceID)
}

// RetryFailedInstanceWithAuth resets a failed instance back to pending and triggers execution.
func (s *Service) RetryFailedInstanceWithAuth(ctx context.Context, authHeader string, coursePhaseID, instanceID uuid.UUID) error {
	if err := s.queries.ResetFailedInstanceToPending(ctx, instanceID, coursePhaseID); err != nil {
		return err
	}
	s.worker.RunPendingInstances(authHeader, coursePhaseID)
	return nil
}

// DeleteInstance removes a resource instance.
func (s *Service) DeleteInstance(ctx context.Context, coursePhaseID, instanceID uuid.UUID) error {
	return s.queries.DeleteResourceInstance(ctx, instanceID, coursePhaseID)
}
