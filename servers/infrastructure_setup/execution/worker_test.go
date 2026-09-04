package execution

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
)

// failingTargetResolver stands in for core being unreachable.
type failingTargetResolver struct {
	err error
}

func (f failingTargetResolver) ResolveTargets(_ context.Context, _ string, _ uuid.UUID, _ db.ResourceScope) ([]ProvisioningTarget, error) {
	return nil, f.err
}

// fakeProvider records how often it was asked to create a resource and returns a
// scripted outcome.
type fakeProvider struct {
	calls    atomic.Int32
	failures int32
	warnings []string
	err      error
}

func (f *fakeProvider) GetType() string                           { return "gitlab" }
func (f *fakeProvider) GetAuthFields() []provider.AuthField       { return nil }
func (f *fakeProvider) SupportedResourceTypes() []string          { return []string{"group"} }
func (f *fakeProvider) ValidateCredentials(context.Context) error { return nil }
func (f *fakeProvider) TemplatedExtraConfigKeys() []string        { return nil }
func (f *fakeProvider) RequiredExtraConfigKeys(string) []string   { return nil }

func (f *fakeProvider) CreateResource(_ context.Context, input provider.CreateResourceInput) (*provider.Resource, error) {
	call := f.calls.Add(1)
	if f.err != nil {
		return nil, f.err
	}
	if call <= f.failures {
		return nil, fmt.Errorf("transient failure %d", call)
	}
	return &provider.Resource{
		ExternalID:  "ext-" + input.Name,
		ExternalURL: "https://example.test/" + input.Name,
		Warnings:    f.warnings,
	}, nil
}

// registerFakeProvider installs the fake in the package-level registry for one test.
func registerFakeProvider(t *testing.T, fake *fakeProvider) {
	t.Helper()
	previous := Registry["gitlab"]
	Registry["gitlab"] = func([]byte) (provider.Provider, error) { return fake, nil }
	t.Cleanup(func() {
		if previous == nil {
			delete(Registry, "gitlab")
			return
		}
		Registry["gitlab"] = previous
	})
}

func seedPendingInstance(t *testing.T, queries *db.Queries, cfg db.ResourceConfig, coursePhaseID, teamID uuid.UUID) db.ResourceInstance {
	t.Helper()
	instance, err := queries.CreateResourceInstance(context.Background(), db.CreateResourceInstanceParams{
		ResourceConfigID: cfg.ID,
		CoursePhaseID:    coursePhaseID,
		TeamID:           &teamID,
	})
	if err != nil {
		t.Fatalf("create instance: %v", err)
	}
	return instance
}

func TestWorkerMarksInstanceCreated(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	fake := &fakeProvider{}
	registerFakeProvider(t, fake)

	coursePhaseID := uuid.New()
	teamID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	instance := seedPendingInstance(t, testDB.Queries, cfg, coursePhaseID, teamID)

	worker := NewWorkerWithResolver(testDB.Conn, fakeTargetResolver{targets: []ProvisioningTarget{
		{Scope: db.ResourceScopePerTeam, TeamID: &teamID, TeamName: "Team A", TemplateData: TemplateData{TeamName: "Team A"}},
	}})
	if err := worker.processPhase(context.Background(), "Bearer test", coursePhaseID); err != nil {
		t.Fatalf("processPhase: %v", err)
	}

	got := getInstance(t, testDB.Queries, coursePhaseID, instance.ID)
	if got.Status != db.ResourceStatusCreated {
		t.Fatalf("status = %s, want created (error: %v)", got.Status, got.ErrorMessage)
	}
	if got.ExternalID == nil || *got.ExternalID != "ext-team-a" {
		t.Fatalf("externalID = %v, want ext-team-a", got.ExternalID)
	}
}

// A member the provider could not add leaves the resource in place, so the instance is
// partial and keeps its external ID.
func TestWorkerMarksInstancePartialOnMemberWarning(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	registerFakeProvider(t, &fakeProvider{warnings: []string{"ghost@example.com: not found"}})

	coursePhaseID := uuid.New()
	teamID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	instance := seedPendingInstance(t, testDB.Queries, cfg, coursePhaseID, teamID)

	worker := NewWorkerWithResolver(testDB.Conn, fakeTargetResolver{targets: []ProvisioningTarget{
		{Scope: db.ResourceScopePerTeam, TeamID: &teamID, TemplateData: TemplateData{TeamName: "Team A"}},
	}})
	if err := worker.processPhase(context.Background(), "Bearer test", coursePhaseID); err != nil {
		t.Fatalf("processPhase: %v", err)
	}

	got := getInstance(t, testDB.Queries, coursePhaseID, instance.ID)
	if got.Status != db.ResourceStatusPartial {
		t.Fatalf("status = %s, want partial", got.Status)
	}
	if got.ExternalID == nil {
		t.Fatal("externalID was cleared, but the resource exists")
	}
	if got.ErrorMessage == nil || *got.ErrorMessage == "" {
		t.Fatal("error message is empty, want the member warning")
	}
}

func TestWorkerRetriesThenFails(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	fake := &fakeProvider{err: errors.New("provider down")}
	registerFakeProvider(t, fake)

	coursePhaseID := uuid.New()
	teamID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	instance := seedPendingInstance(t, testDB.Queries, cfg, coursePhaseID, teamID)

	worker := NewWorkerWithResolver(testDB.Conn, fakeTargetResolver{targets: []ProvisioningTarget{
		{Scope: db.ResourceScopePerTeam, TeamID: &teamID, TemplateData: TemplateData{TeamName: "Team A"}},
	}})
	if err := worker.processPhase(context.Background(), "Bearer test", coursePhaseID); err != nil {
		t.Fatalf("processPhase: %v", err)
	}

	if got := fake.calls.Load(); got != maxRetries {
		t.Fatalf("provider calls = %d, want %d", got, maxRetries)
	}
	got := getInstance(t, testDB.Queries, coursePhaseID, instance.ID)
	if got.Status != db.ResourceStatusFailed {
		t.Fatalf("status = %s, want failed", got.Status)
	}
}

// A transient failure must not exhaust the budget: the second attempt succeeds.
func TestWorkerSucceedsAfterTransientFailure(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	fake := &fakeProvider{failures: 1}
	registerFakeProvider(t, fake)

	coursePhaseID := uuid.New()
	teamID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	instance := seedPendingInstance(t, testDB.Queries, cfg, coursePhaseID, teamID)

	worker := NewWorkerWithResolver(testDB.Conn, fakeTargetResolver{targets: []ProvisioningTarget{
		{Scope: db.ResourceScopePerTeam, TeamID: &teamID, TemplateData: TemplateData{TeamName: "Team A"}},
	}})
	if err := worker.processPhase(context.Background(), "Bearer test", coursePhaseID); err != nil {
		t.Fatalf("processPhase: %v", err)
	}

	if got := getInstance(t, testDB.Queries, coursePhaseID, instance.ID); got.Status != db.ResourceStatusCreated {
		t.Fatalf("status = %s, want created", got.Status)
	}
}

// Two workers running the same phase concurrently must provision exactly once. Before
// the claim was atomic, both picked up the same pending row.
func TestConcurrentWorkersProvisionOnce(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	fake := &fakeProvider{}
	registerFakeProvider(t, fake)

	coursePhaseID := uuid.New()
	teamID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	seedPendingInstance(t, testDB.Queries, cfg, coursePhaseID, teamID)

	resolver := fakeTargetResolver{targets: []ProvisioningTarget{
		{Scope: db.ResourceScopePerTeam, TeamID: &teamID, TemplateData: TemplateData{TeamName: "Team A"}},
	}}

	var wg sync.WaitGroup
	start := make(chan struct{})
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			worker := NewWorkerWithResolver(testDB.Conn, resolver)
			if err := worker.processPhase(context.Background(), "Bearer test", coursePhaseID); err != nil {
				t.Errorf("processPhase: %v", err)
			}
		}()
	}
	close(start)
	wg.Wait()

	if got := fake.calls.Load(); got != 1 {
		t.Fatalf("provider called %d times, want exactly 1", got)
	}
}

// A target that disappeared upstream fails its instance instead of provisioning
// something unrelated.
func TestWorkerFailsInstanceWithoutTarget(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	fake := &fakeProvider{}
	registerFakeProvider(t, fake)

	coursePhaseID := uuid.New()
	teamID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	instance := seedPendingInstance(t, testDB.Queries, cfg, coursePhaseID, teamID)

	otherTeam := uuid.New()
	worker := NewWorkerWithResolver(testDB.Conn, fakeTargetResolver{targets: []ProvisioningTarget{
		{Scope: db.ResourceScopePerTeam, TeamID: &otherTeam},
	}})
	if err := worker.processPhase(context.Background(), "Bearer test", coursePhaseID); err != nil {
		t.Fatalf("processPhase: %v", err)
	}

	if fake.calls.Load() != 0 {
		t.Fatal("provider was called for an instance whose target no longer exists")
	}
	if got := getInstance(t, testDB.Queries, coursePhaseID, instance.ID); got.Status != db.ResourceStatusFailed {
		t.Fatalf("status = %s, want failed", got.Status)
	}
}

// An unresolvable name template must not reach the provider.
func TestWorkerFailsInstanceOnBadNameTemplate(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	fake := &fakeProvider{}
	registerFakeProvider(t, fake)

	coursePhaseID := uuid.New()
	teamID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	if _, err := testDB.Queries.UpdateResourceConfig(context.Background(), db.UpdateResourceConfigParams{
		ID:                  cfg.ID,
		CoursePhaseID:       coursePhaseID,
		ResourceType:        cfg.ResourceType,
		Scope:               cfg.Scope,
		NameTemplate:        "{{unknownPlaceholder}}",
		PermissionMapping:   cfg.PermissionMapping,
		ResourceExtraConfig: cfg.ResourceExtraConfig,
	}); err != nil {
		t.Fatalf("update resource config: %v", err)
	}
	instance := seedPendingInstance(t, testDB.Queries, cfg, coursePhaseID, teamID)

	worker := NewWorkerWithResolver(testDB.Conn, fakeTargetResolver{targets: []ProvisioningTarget{
		{Scope: db.ResourceScopePerTeam, TeamID: &teamID},
	}})
	if err := worker.processPhase(context.Background(), "Bearer test", coursePhaseID); err != nil {
		t.Fatalf("processPhase: %v", err)
	}

	if fake.calls.Load() != 0 {
		t.Fatal("provider was called with an unresolved name template")
	}
	if got := getInstance(t, testDB.Queries, coursePhaseID, instance.ID); got.Status != db.ResourceStatusFailed {
		t.Fatalf("status = %s, want failed", got.Status)
	}
}

// A resolver failure after the instances were claimed must hand them back: an
// instance left in_progress makes every later trigger answer 409 and cannot be
// retried either, so the phase would be stuck until the row is deleted.
func TestWorkerFailsClaimedInstancesWhenTargetResolutionFails(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	fake := &fakeProvider{}
	registerFakeProvider(t, fake)

	coursePhaseID := uuid.New()
	teamID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	instance := seedPendingInstance(t, testDB.Queries, cfg, coursePhaseID, teamID)

	worker := NewWorkerWithResolver(testDB.Conn, failingTargetResolver{err: errors.New("core is unreachable")})
	if err := worker.processPhase(context.Background(), "Bearer test", coursePhaseID); err == nil {
		t.Fatal("processPhase = nil error, want the resolver failure")
	}

	got := getInstance(t, testDB.Queries, coursePhaseID, instance.ID)
	if got.Status != db.ResourceStatusFailed {
		t.Fatalf("status = %s, want failed so the phase can be triggered again", got.Status)
	}
	if got.ErrorMessage == nil || *got.ErrorMessage == "" {
		t.Fatal("error message is empty, want the resolver failure recorded on the instance")
	}
	if fake.calls.Load() != 0 {
		t.Fatal("provider was called although no target could be resolved")
	}
}

// Instances left claimed by a crashed process are marked failed, which is terminal
// (so the phase can run again) and retryable by the lecturer.
func TestFailStaleClaimsRecoversAbandonedInstances(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	teamID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	seedPendingInstance(t, testDB.Queries, cfg, coursePhaseID, teamID)

	claimed, err := testDB.Queries.ClaimPendingInstances(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if len(claimed) != 1 {
		t.Fatalf("claimed = %d, want 1", len(claimed))
	}

	worker := NewWorkerWithResolver(testDB.Conn, fakeTargetResolver{})

	// A claim younger than the cutoff may belong to a run that is still going.
	if recovered, err := worker.FailStaleClaims(context.Background()); err != nil {
		t.Fatalf("FailStaleClaims: %v", err)
	} else if recovered != 0 {
		t.Fatalf("recovered = %d for a fresh claim, want 0", recovered)
	}
	if got := getInstance(t, testDB.Queries, coursePhaseID, claimed[0].ID); got.Status != db.ResourceStatusInProgress {
		t.Fatalf("status = %s, want a fresh claim left in_progress", got.Status)
	}

	ageClaim(t, testDB.Conn, claimed[0].ID)

	if recovered, err := worker.FailStaleClaims(context.Background()); err != nil {
		t.Fatalf("FailStaleClaims: %v", err)
	} else if recovered != 1 {
		t.Fatalf("recovered = %d, want 1", recovered)
	}
	got := getInstance(t, testDB.Queries, coursePhaseID, claimed[0].ID)
	if got.Status != db.ResourceStatusFailed {
		t.Fatalf("status = %s, want failed", got.Status)
	}
	if got.ErrorMessage == nil || *got.ErrorMessage != staleClaimMessage {
		t.Fatalf("error message = %v, want %q", got.ErrorMessage, staleClaimMessage)
	}
}

// ageClaim backdates the claim so the sweep sees it as abandoned.
func ageClaim(t *testing.T, pool *pgxpool.Pool, instanceID uuid.UUID) {
	t.Helper()
	if _, err := pool.Exec(context.Background(),
		"UPDATE resource_instance SET updated_at = NOW() - $2::interval WHERE id = $1",
		instanceID, fmt.Sprintf("%d seconds", int((staleClaimAge+time.Minute).Seconds())),
	); err != nil {
		t.Fatalf("age claim: %v", err)
	}
}

func getInstance(t *testing.T, queries *db.Queries, coursePhaseID, instanceID uuid.UUID) db.ResourceInstance {
	t.Helper()
	instance, err := queries.GetResourceInstance(context.Background(), db.GetResourceInstanceParams{
		ID:            instanceID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		t.Fatalf("get instance: %v", err)
	}
	return instance
}
