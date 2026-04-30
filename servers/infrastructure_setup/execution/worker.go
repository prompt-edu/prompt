package execution

import (
	"context"
	"fmt"
	"math/rand"
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
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewWorker creates a Worker.
func NewWorker(pool *pgxpool.Pool) *Worker {
	return &Worker{pool: pool, queries: db.New(pool)}
}

// RunPendingInstances processes all pending instances for the given course phase.
// Spawning is done in a goroutine so the HTTP handler returns immediately.
func (w *Worker) RunPendingInstances(coursePhaseID uuid.UUID) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		if err := w.processPhase(ctx, coursePhaseID); err != nil {
			log.WithError(err).WithField("coursePhaseID", coursePhaseID).
				Error("execution worker: processPhase failed")
		}
	}()
}

func (w *Worker) processPhase(ctx context.Context, coursePhaseID uuid.UUID) error {
	// Retrieve pending instances from DB.
	instances, err := w.queries.ListPendingInstances(ctx, coursePhaseID)
	if err != nil {
		return fmt.Errorf("list pending instances: %w", err)
	}

	if len(instances) == 0 {
		return nil
	}

	// Semaphore for max concurrency.
	sem := make(chan struct{}, maxWorkers)
	done := make(chan struct{}, len(instances))

	for _, inst := range instances {
		inst := inst
		sem <- struct{}{}
		go func() {
			defer func() {
				<-sem
				done <- struct{}{}
			}()
			if err := w.processInstance(ctx, inst); err != nil {
				log.WithError(err).WithField("instanceID", inst.ID).
					Error("execution worker: processInstance failed")
			}
		}()
	}

	// Wait for all goroutines to finish.
	for range instances {
		<-done
	}
	return nil
}

// processInstance executes a single resource instance with retry/backoff.
func (w *Worker) processInstance(ctx context.Context, inst db.ResourceInstance) error {
	if err := w.queries.UpdateResourceInstanceStatus(ctx, db.UpdateResourceInstanceStatusParams{
		ID:     inst.ID,
		Status: db.ResourceStatusInProgress,
	}); err != nil {
		return err
	}

	// Load resource config.
	config, err := w.queries.GetResourceConfig(ctx, inst.ResourceConfigID, inst.CoursePhaseID)
	if err != nil {
		return w.failInstance(ctx, inst.ID, fmt.Sprintf("load resource config: %v", err))
	}

	// Load provider config.
	providerCfg, err := w.queries.GetProviderConfig(ctx, inst.CoursePhaseID, config.ProviderType)
	if err != nil {
		return w.failInstance(ctx, inst.ID, fmt.Sprintf("load provider config: %v", err))
	}

	// Build provider.
	factory, ok := Registry[string(config.ProviderType)]
	if !ok {
		return w.failInstance(ctx, inst.ID, fmt.Sprintf("unknown provider type: %s", config.ProviderType))
	}
	prov, err := factory(providerCfg.Credentials)
	if err != nil {
		return w.failInstance(ctx, inst.ID, fmt.Sprintf("build provider: %v", err))
	}

	// Build template data.
	tmplData := w.buildTemplateData(inst)

	// Resolve resource name.
	resolvedName, err := ResolveName(config.NameTemplate, tmplData)
	if err != nil {
		return w.failInstance(ctx, inst.ID, fmt.Sprintf("resolve name: %v", err))
	}

	// Parse config.
	permissionMap, err := ParsePermissionMapping(config.PermissionMapping)
	if err != nil {
		return w.failInstance(ctx, inst.ID, fmt.Sprintf("parse permission map: %v", err))
	}
	extraConfig, err := ParseExtraConfig(config.ResourceExtraConfig)
	if err != nil {
		return w.failInstance(ctx, inst.ID, fmt.Sprintf("parse extra config: %v", err))
	}

	// Build members list.
	members := w.buildMemberList(inst)

	input := provider.CreateResourceInput{
		Name:              resolvedName,
		ResourceType:      config.ResourceType,
		Members:           members,
		PermissionMapping: permissionMap,
		ExtraConfig:       extraConfig,
	}

	// Execute with retry.
	var resource *provider.Resource
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := baseBackoff * (1 << uint(attempt-1))
			// Add ±20% jitter.
			jitter := time.Duration(rand.Int63n(int64(backoff / 5)))
			time.Sleep(backoff + jitter)
		}

		resource, lastErr = prov.CreateResource(ctx, input)
		if lastErr == nil {
			break
		}
		log.WithFields(log.Fields{
			"instanceID": inst.ID,
			"attempt":    attempt + 1,
			"error":      lastErr,
		}).Warn("execution worker: retry")
	}

	if lastErr != nil {
		return w.failInstance(ctx, inst.ID, lastErr.Error())
	}

	return w.queries.UpdateResourceInstanceStatus(ctx, db.UpdateResourceInstanceStatusParams{
		ID:          inst.ID,
		Status:      db.ResourceStatusCreated,
		ExternalID:  &resource.ExternalID,
		ExternalURL: &resource.ExternalURL,
	})
}

func (w *Worker) failInstance(ctx context.Context, id uuid.UUID, msg string) error {
	return w.queries.UpdateResourceInstanceStatus(ctx, db.UpdateResourceInstanceStatusParams{
		ID:           id,
		Status:       db.ResourceStatusFailed,
		ErrorMessage: &msg,
	})
}

// buildTemplateData constructs a TemplateData from a resource instance.
// Team/student names are filled with placeholder values; extend this to load
// real names from the core API if needed.
func (w *Worker) buildTemplateData(inst db.ResourceInstance) TemplateData {
	data := TemplateData{}
	if inst.TeamID != nil {
		data.TeamName = inst.TeamID.String()[:8]
	}
	if inst.CourseParticipationID != nil {
		data.StudentLogin = inst.CourseParticipationID.String()[:8]
	}
	return data
}

// buildMemberList assembles the members list for resource creation.
// For per_student resources it returns the student; for per_team resources
// it returns an empty list (teams are handled by member lookup in the provider).
func (w *Worker) buildMemberList(inst db.ResourceInstance) []provider.Member {
	if inst.CourseParticipationID != nil {
		return []provider.Member{
			{Email: fmt.Sprintf("student+%s@placeholder.local", inst.CourseParticipationID), Role: "student"},
		}
	}
	return []provider.Member{}
}
