package execution

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
)

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

// Instances left in_progress by a crash are returned to pending at startup.
func TestResetInProgressToPendingRecoversStuckInstances(t *testing.T) {
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

	if err := testDB.Queries.ResetInProgressToPending(context.Background()); err != nil {
		t.Fatalf("reset in progress: %v", err)
	}
	if got := getInstance(t, testDB.Queries, coursePhaseID, claimed[0].ID); got.Status != db.ResourceStatusPending {
		t.Fatalf("status = %s, want pending", got.Status)
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
