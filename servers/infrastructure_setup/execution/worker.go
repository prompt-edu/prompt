package execution

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
	log "github.com/sirupsen/logrus"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxWorkers  = 5
	maxRetries  = 3
	baseBackoff = time.Second
)

// Registry maps provider type strings to provider factory functions.
// Populated by main.go during startup.
var Registry = map[string]func(credentials []byte) (provider.Provider, error){}

// Worker processes pending resource instances.
// It is safe to call concurrently: a semaphore limits parallelism to maxWorkers.
type Worker struct {
	pool     *pgxpool.Pool
	queries  *db.Queries
	resolver TargetResolver
}

// NewWorker creates a Worker.
func NewWorker(pool *pgxpool.Pool) *Worker {
	queries := db.New(pool)
	return &Worker{pool: pool, queries: queries, resolver: NewCoreTargetResolver(queries)}
}

// NewWorkerWithResolver creates a Worker with an injected resolver for tests.
func NewWorkerWithResolver(pool *pgxpool.Pool, resolver TargetResolver) *Worker {
	return &Worker{pool: pool, queries: db.New(pool), resolver: resolver}
}

// RunPendingInstances processes all pending instances for the given course phase.
// Spawning is done in a goroutine so the HTTP handler returns immediately.
func (w *Worker) RunPendingInstances(authHeader string, coursePhaseID uuid.UUID) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		if err := w.processPhase(ctx, authHeader, coursePhaseID); err != nil {
			log.WithError(err).WithField("coursePhaseID", coursePhaseID).
				Error("execution worker: processPhase failed")
		}
	}()
}

func (w *Worker) processPhase(ctx context.Context, authHeader string, coursePhaseID uuid.UUID) error {
	instances, err := w.queries.ClaimPendingInstances(ctx, coursePhaseID)
	if err != nil {
		return fmt.Errorf("claim pending instances: %w", err)
	}
	if len(instances) == 0 {
		return nil
	}

	configs, err := w.resourceConfigsByID(ctx, coursePhaseID)
	if err != nil {
		return err
	}

	targets, err := w.targetsByScope(ctx, authHeader, coursePhaseID, instances, configs)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxWorkers)

	for _, inst := range instances {
		wg.Add(1)
		sem <- struct{}{}
		go func(inst db.ResourceInstance) {
			defer func() {
				<-sem
				wg.Done()
			}()
			if err := w.processInstance(ctx, inst, configs, targets); err != nil {
				log.WithError(err).WithField("instanceID", inst.ID).
					Error("execution worker: processInstance failed")
			}
		}(inst)
	}

	wg.Wait()
	return nil
}

func (w *Worker) resourceConfigsByID(ctx context.Context, coursePhaseID uuid.UUID) (map[uuid.UUID]db.ResourceConfig, error) {
	list, err := w.queries.ListResourceConfigs(ctx, coursePhaseID)
	if err != nil {
		return nil, fmt.Errorf("list resource configs: %w", err)
	}
	configs := make(map[uuid.UUID]db.ResourceConfig, len(list))
	for _, cfg := range list {
		configs[cfg.ID] = cfg
	}
	return configs, nil
}

// targetsByScope resolves each scope present in the claimed instances exactly once.
// Resolution hits core over HTTP, so doing it per instance would issue two requests
// per resource rather than two per scope.
func (w *Worker) targetsByScope(
	ctx context.Context,
	authHeader string,
	coursePhaseID uuid.UUID,
	instances []db.ResourceInstance,
	configs map[uuid.UUID]db.ResourceConfig,
) (map[db.ResourceScope]targetIndex, error) {
	scopes := make(map[db.ResourceScope]struct{})
	for _, inst := range instances {
		if cfg, ok := configs[inst.ResourceConfigID]; ok {
			scopes[cfg.Scope] = struct{}{}
		}
	}

	resolved := make(map[db.ResourceScope]targetIndex, len(scopes))
	for scope := range scopes {
		targets, err := w.resolver.ResolveTargets(ctx, authHeader, coursePhaseID, scope)
		if err != nil {
			return nil, fmt.Errorf("resolve targets for scope %s: %w", scope, err)
		}
		resolved[scope] = newTargetIndex(targets)
	}
	return resolved, nil
}

// targetIndex looks up a provisioning target by the identifier stored on the instance.
type targetIndex struct {
	byTeam    map[uuid.UUID]ProvisioningTarget
	byStudent map[uuid.UUID]ProvisioningTarget
}

func newTargetIndex(targets []ProvisioningTarget) targetIndex {
	index := targetIndex{
		byTeam:    make(map[uuid.UUID]ProvisioningTarget),
		byStudent: make(map[uuid.UUID]ProvisioningTarget),
	}
	for _, target := range targets {
		if target.TeamID != nil {
			index.byTeam[*target.TeamID] = target
		}
		if target.CourseParticipationID != nil {
			index.byStudent[*target.CourseParticipationID] = target
		}
	}
	return index
}

func (i targetIndex) find(inst db.ResourceInstance) (ProvisioningTarget, bool) {
	if inst.TeamID != nil {
		target, ok := i.byTeam[*inst.TeamID]
		return target, ok
	}
	if inst.CourseParticipationID != nil {
		target, ok := i.byStudent[*inst.CourseParticipationID]
		return target, ok
	}
	return ProvisioningTarget{}, false
}

// processInstance executes a single claimed resource instance with retry/backoff.
func (w *Worker) processInstance(
	ctx context.Context,
	inst db.ResourceInstance,
	configs map[uuid.UUID]db.ResourceConfig,
	targets map[db.ResourceScope]targetIndex,
) error {
	config, ok := configs[inst.ResourceConfigID]
	if !ok {
		return w.failInstance(ctx, inst.ID, "resource config no longer exists")
	}

	target, ok := targets[config.Scope].find(inst)
	if !ok {
		return w.failInstance(ctx, inst.ID, "resource instance target no longer exists")
	}

	providerCfg, err := w.queries.GetProviderConfig(ctx, db.GetProviderConfigParams{
		CoursePhaseID: inst.CoursePhaseID,
		ProviderType:  config.ProviderType,
	})
	if err != nil {
		return w.failInstance(ctx, inst.ID, fmt.Sprintf("load provider config: %v", err))
	}

	factory, ok := Registry[string(config.ProviderType)]
	if !ok {
		return w.failInstance(ctx, inst.ID, fmt.Sprintf("unknown provider type: %s", config.ProviderType))
	}
	prov, err := factory(providerCfg.Credentials)
	if err != nil {
		return w.failInstance(ctx, inst.ID, fmt.Sprintf("build provider: %v", err))
	}

	resolvedName, err := ResolveName(config.NameTemplate, target.TemplateData)
	if err != nil {
		return w.failInstance(ctx, inst.ID, fmt.Sprintf("resolve name: %v", err))
	}

	permissionMap, err := ParsePermissionMapping(config.PermissionMapping)
	if err != nil {
		return w.failInstance(ctx, inst.ID, fmt.Sprintf("parse permission map: %v", err))
	}
	extraConfig, err := ParseExtraConfig(config.ResourceExtraConfig)
	if err != nil {
		return w.failInstance(ctx, inst.ID, fmt.Sprintf("parse extra config: %v", err))
	}

	input := provider.CreateResourceInput{
		Name:              resolvedName,
		ResourceType:      config.ResourceType,
		Members:           target.Members,
		PermissionMapping: permissionMap,
		ExtraConfig:       extraConfig,
	}

	resource, err := w.createWithRetry(ctx, prov, input, inst.ID)
	if err != nil {
		return w.failInstance(ctx, inst.ID, err.Error())
	}

	if len(resource.Warnings) > 0 {
		message := strings.Join(resource.Warnings, "; ")
		return w.queries.MarkInstancePartial(ctx, db.MarkInstancePartialParams{
			ID:           inst.ID,
			ExternalID:   &resource.ExternalID,
			ExternalUrl:  &resource.ExternalURL,
			ErrorMessage: &message,
		})
	}

	return w.queries.MarkInstanceCreated(ctx, db.MarkInstanceCreatedParams{
		ID:          inst.ID,
		ExternalID:  &resource.ExternalID,
		ExternalUrl: &resource.ExternalURL,
	})
}

func (w *Worker) createWithRetry(ctx context.Context, prov provider.Provider, input provider.CreateResourceInput, instanceID uuid.UUID) (*provider.Resource, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := baseBackoff * (1 << uint(attempt-1))
			// Add up to 20% jitter.
			jitter := time.Duration(rand.Int63n(int64(backoff / 5)))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff + jitter):
			}
		}

		resource, err := prov.CreateResource(ctx, input)
		if err == nil {
			return resource, nil
		}
		lastErr = err
		log.WithFields(log.Fields{
			"instanceID": instanceID,
			"attempt":    attempt + 1,
			"error":      err,
		}).Warn("execution worker: retry")
	}
	return nil, lastErr
}

func (w *Worker) failInstance(ctx context.Context, id uuid.UUID, msg string) error {
	return w.queries.MarkInstanceFailed(ctx, db.MarkInstanceFailedParams{
		ID:           id,
		ErrorMessage: &msg,
	})
}
