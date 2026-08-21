package execution

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

// ErrExecutionInProgress is returned when a run is already queued or running for the phase.
var ErrExecutionInProgress = errors.New("an execution is already in progress for this course phase")

// ErrInstanceNotRetryable is returned when an instance exists but is not in a retryable state.
var ErrInstanceNotRetryable = errors.New("only failed or partial instances can be retried")

// ErrInstanceNotFound is returned when the instance does not exist in the course phase.
var ErrInstanceNotFound = errors.New("resource instance not found")

// ErrProviderNotConfigured is returned when a referenced provider has no credentials.
var ErrProviderNotConfigured = errors.New("provider credentials are missing")

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
//
// Targets are resolved first, outside the transaction, because resolution calls core
// over HTTP. The transaction then takes a per-phase advisory lock so the in-progress
// check and the inserts cannot interleave with a second trigger.
func (s *Service) TriggerExecution(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) error {
	configs, err := s.queries.ListResourceConfigs(ctx, coursePhaseID)
	if err != nil {
		return err
	}
	if len(configs) == 0 {
		return nil
	}

	if err := s.assertProvidersConfigured(ctx, coursePhaseID, configs); err != nil {
		return err
	}

	targetsByScope, err := s.resolveScopes(ctx, authHeader, coursePhaseID, configs)
	if err != nil {
		return err
	}

	if err := s.createInstances(ctx, coursePhaseID, configs, targetsByScope); err != nil {
		return err
	}

	s.worker.RunPendingInstances(authHeader, coursePhaseID)
	return nil
}

// assertProvidersConfigured rejects a run whose providers lost their credentials, which
// is the state a copied phase starts in. Without this the instances would all be created
// and then fail one by one on decryption.
func (s *Service) assertProvidersConfigured(ctx context.Context, coursePhaseID uuid.UUID, configs []db.ResourceConfig) error {
	configured, err := s.queries.ListConfiguredProviderConfigs(ctx, coursePhaseID)
	if err != nil {
		return err
	}
	available := make(map[db.ProviderType]struct{}, len(configured))
	for _, pc := range configured {
		available[pc.ProviderType] = struct{}{}
	}

	for _, cfg := range configs {
		if _, ok := available[cfg.ProviderType]; !ok {
			return fmt.Errorf("%w: %s", ErrProviderNotConfigured, cfg.ProviderType)
		}
	}
	return nil
}

func (s *Service) resolveScopes(ctx context.Context, authHeader string, coursePhaseID uuid.UUID, configs []db.ResourceConfig) (map[db.ResourceScope][]ProvisioningTarget, error) {
	targetsByScope := make(map[db.ResourceScope][]ProvisioningTarget)
	for _, cfg := range configs {
		if _, done := targetsByScope[cfg.Scope]; done {
			continue
		}
		targets, err := s.resolver.ResolveTargets(ctx, authHeader, coursePhaseID, cfg.Scope)
		if err != nil {
			return nil, err
		}
		targetsByScope[cfg.Scope] = targets
	}
	return targetsByScope, nil
}

func (s *Service) createInstances(ctx context.Context, coursePhaseID uuid.UUID, configs []db.ResourceConfig, targetsByScope map[db.ResourceScope][]ProvisioningTarget) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	locked, err := qtx.TryLockPhaseExecution(ctx, coursePhaseID.String())
	if err != nil {
		return err
	}
	if !locked {
		return ErrExecutionInProgress
	}

	nonTerminal, err := qtx.CountNonTerminalInstances(ctx, coursePhaseID)
	if err != nil {
		return err
	}
	if nonTerminal > 0 {
		return ErrExecutionInProgress
	}

	for _, cfg := range configs {
		for _, target := range targetsByScope[cfg.Scope] {
			if _, err := qtx.CreateResourceInstance(ctx, createResourceInstanceParams(cfg, coursePhaseID, target)); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					continue
				}
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func createResourceInstanceParams(cfg db.ResourceConfig, coursePhaseID uuid.UUID, target ProvisioningTarget) db.CreateResourceInstanceParams {
	return db.CreateResourceInstanceParams{
		ResourceConfigID:      cfg.ID,
		CoursePhaseID:         coursePhaseID,
		TeamID:                target.TeamID,
		CourseParticipationID: target.CourseParticipationID,
	}
}

// ListInstances returns all resource instances for a course phase. The slice is never
// nil, so the endpoint answers with [] rather than null when nothing is provisioned.
func (s *Service) ListInstances(ctx context.Context, coursePhaseID uuid.UUID) ([]db.ResourceInstance, error) {
	instances, err := s.queries.ListResourceInstances(ctx, coursePhaseID)
	if err != nil {
		return nil, err
	}
	if instances == nil {
		return []db.ResourceInstance{}, nil
	}
	return instances, nil
}

// RetryInstance resets a failed or partial instance back to pending and starts the worker.
// A worker is only started when a row was actually reset.
func (s *Service) RetryInstance(ctx context.Context, authHeader string, coursePhaseID, instanceID uuid.UUID) error {
	_, err := s.queries.ResetInstanceToPending(ctx, db.ResetInstanceToPendingParams{
		ID:            instanceID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		// No row was reset: either the instance does not exist, or it is in a state
		// that cannot be retried.
		if _, getErr := s.queries.GetResourceInstance(ctx, db.GetResourceInstanceParams{
			ID:            instanceID,
			CoursePhaseID: coursePhaseID,
		}); getErr != nil {
			if errors.Is(getErr, pgx.ErrNoRows) {
				return ErrInstanceNotFound
			}
			return getErr
		}
		return ErrInstanceNotRetryable
	}

	s.worker.RunPendingInstances(authHeader, coursePhaseID)
	return nil
}

// DeleteInstance removes a resource instance. The external resource is never touched:
// providers adopt resources by name, so PROMPT cannot know that it owns them.
func (s *Service) DeleteInstance(ctx context.Context, coursePhaseID, instanceID uuid.UUID) error {
	return s.queries.DeleteResourceInstance(ctx, db.DeleteResourceInstanceParams{
		ID:            instanceID,
		CoursePhaseID: coursePhaseID,
	})
}
