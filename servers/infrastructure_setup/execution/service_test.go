package execution

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

type fakeTargetResolver struct {
	targets []ProvisioningTarget
}

func (f fakeTargetResolver) ResolveTargets(_ context.Context, _ string, _ uuid.UUID, _ db.ResourceScope) ([]ProvisioningTarget, error) {
	return f.targets, nil
}

func setupExecutionTestDB(t *testing.T) (*sdkTestUtils.TestDB[*db.Queries], func()) {
	t.Helper()

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(context.Background(), "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries {
		return db.New(conn)
	})
	if err != nil {
		t.Fatalf("setup test db: %v", err)
	}
	return testDB, cleanup
}

func createResourceConfig(t *testing.T, queries *db.Queries, coursePhaseID uuid.UUID, scope db.ResourceScope) db.ResourceConfig {
	t.Helper()

	_, err := queries.UpsertProviderConfig(context.Background(), db.UpsertProviderConfigParams{
		CoursePhaseID: coursePhaseID,
		ProviderType:  db.ProviderTypeGitlab,
		Credentials:   []byte("encrypted"),
	})
	if err != nil {
		t.Fatalf("upsert provider config: %v", err)
	}

	cfg, err := queries.CreateResourceConfig(context.Background(), db.CreateResourceConfigParams{
		CoursePhaseID:       coursePhaseID,
		ProviderType:        db.ProviderTypeGitlab,
		ResourceType:        "group",
		Scope:               scope,
		NameTemplate:        "{{teamName}}{{studentLogin}}",
		PermissionMapping:   []byte(`{"student":"developer"}`),
		ResourceExtraConfig: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("create resource config: %v", err)
	}
	return cfg
}

// createInstances runs the transactional part of a trigger without spawning the
// background worker, which would need a real provider.
func createInstances(t *testing.T, service *Service, coursePhaseID uuid.UUID, cfg db.ResourceConfig, targets []ProvisioningTarget) error {
	t.Helper()
	return service.createInstances(
		context.Background(),
		coursePhaseID,
		[]db.ResourceConfig{cfg},
		map[db.ResourceScope][]ProvisioningTarget{cfg.Scope: targets},
	)
}

func TestTriggerCreatesOneInstancePerTeam(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	teamA := uuid.New()
	teamB := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	targets := []ProvisioningTarget{
		{Scope: db.ResourceScopePerTeam, TeamID: &teamA, TeamName: "Team A"},
		{Scope: db.ResourceScopePerTeam, TeamID: &teamB, TeamName: "Team B"},
	}
	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{targets: targets})

	if err := createInstances(t, service, coursePhaseID, cfg, targets); err != nil {
		t.Fatalf("create instances: %v", err)
	}

	instances, err := testDB.Queries.ListResourceInstances(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("instances = %d, want 2", len(instances))
	}
}

func TestTriggerCreatesOneInstancePerStudent(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	participationA := uuid.New()
	participationB := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerStudent)
	targets := []ProvisioningTarget{
		{Scope: db.ResourceScopePerStudent, CourseParticipationID: &participationA},
		{Scope: db.ResourceScopePerStudent, CourseParticipationID: &participationB},
	}
	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{targets: targets})

	if err := createInstances(t, service, coursePhaseID, cfg, targets); err != nil {
		t.Fatalf("create instances: %v", err)
	}

	instances, err := testDB.Queries.ListResourceInstances(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("instances = %d, want 2", len(instances))
	}
}

func TestTriggerRejectsSecondRunWhileWorkIsPending(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	teamID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	targets := []ProvisioningTarget{{Scope: db.ResourceScopePerTeam, TeamID: &teamID, TeamName: "Team A"}}
	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{targets: targets})

	if err := createInstances(t, service, coursePhaseID, cfg, targets); err != nil {
		t.Fatalf("first trigger: %v", err)
	}

	err := createInstances(t, service, coursePhaseID, cfg, targets)
	if !errors.Is(err, ErrExecutionInProgress) {
		t.Fatalf("second trigger error = %v, want ErrExecutionInProgress", err)
	}
}

// Two triggers racing on separate connections must not both create instances. The
// advisory lock is transaction-scoped, so this exercises the real serialisation path.
func TestConcurrentTriggersCreateInstancesOnce(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	teamA := uuid.New()
	teamB := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	targets := []ProvisioningTarget{
		{Scope: db.ResourceScopePerTeam, TeamID: &teamA, TeamName: "Team A"},
		{Scope: db.ResourceScopePerTeam, TeamID: &teamB, TeamName: "Team B"},
	}
	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{targets: targets})

	var wg sync.WaitGroup
	results := make([]error, 2)
	start := make(chan struct{})
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			results[i] = createInstances(t, service, coursePhaseID, cfg, targets)
		}(i)
	}
	close(start)
	wg.Wait()

	succeeded := 0
	for _, err := range results {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, ErrExecutionInProgress):
		default:
			t.Fatalf("unexpected trigger error: %v", err)
		}
	}
	if succeeded != 1 {
		t.Fatalf("successful triggers = %d, want 1", succeeded)
	}

	instances, err := testDB.Queries.ListResourceInstances(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("instances = %d, want 2", len(instances))
	}
}

func TestTriggerRejectsProviderWithoutCredentials(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	// A copied phase keeps the provider row but drops the credentials.
	if _, err := testDB.Queries.UpsertProviderConfig(context.Background(), db.UpsertProviderConfigParams{
		CoursePhaseID: coursePhaseID,
		ProviderType:  db.ProviderTypeGitlab,
		Credentials:   []byte{},
	}); err != nil {
		t.Fatalf("clear credentials: %v", err)
	}

	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{})
	err := service.assertProvidersConfigured(context.Background(), coursePhaseID, []db.ResourceConfig{cfg})
	if !errors.Is(err, ErrProviderNotConfigured) {
		t.Fatalf("error = %v, want ErrProviderNotConfigured", err)
	}
}

func TestRetryDistinguishesMissingAndNonRetryableInstances(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	teamID := uuid.New()
	instance, err := testDB.Queries.CreateResourceInstance(context.Background(), db.CreateResourceInstanceParams{
		ResourceConfigID: cfg.ID,
		CoursePhaseID:    coursePhaseID,
		TeamID:           &teamID,
	})
	if err != nil {
		t.Fatalf("create instance: %v", err)
	}
	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{})

	if err := service.RetryInstance(context.Background(), "Bearer test", coursePhaseID, uuid.New()); !errors.Is(err, ErrInstanceNotFound) {
		t.Fatalf("retry of unknown instance = %v, want ErrInstanceNotFound", err)
	}

	// A pending instance is already queued and must not be retried.
	if err := service.RetryInstance(context.Background(), "Bearer test", coursePhaseID, instance.ID); !errors.Is(err, ErrInstanceNotRetryable) {
		t.Fatalf("retry of pending instance = %v, want ErrInstanceNotRetryable", err)
	}
}

func TestRetryResetsPartialInstanceAndKeepsExternalID(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	teamID := uuid.New()
	instance, err := testDB.Queries.CreateResourceInstance(context.Background(), db.CreateResourceInstanceParams{
		ResourceConfigID: cfg.ID,
		CoursePhaseID:    coursePhaseID,
		TeamID:           &teamID,
	})
	if err != nil {
		t.Fatalf("create instance: %v", err)
	}

	externalID := "42"
	externalURL := "https://gitlab.example.com/groups/team-a"
	warning := "student@example.com: not found"
	if err := testDB.Queries.MarkInstancePartial(context.Background(), db.MarkInstancePartialParams{
		ID:           instance.ID,
		ExternalID:   &externalID,
		ExternalUrl:  &externalURL,
		ErrorMessage: &warning,
	}); err != nil {
		t.Fatalf("mark partial: %v", err)
	}

	reset, err := testDB.Queries.ResetInstanceToPending(context.Background(), db.ResetInstanceToPendingParams{
		ID:            instance.ID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		t.Fatalf("reset partial instance: %v", err)
	}
	if reset.Status != db.ResourceStatusPending {
		t.Fatalf("status = %s, want pending", reset.Status)
	}
	if reset.ExternalID == nil || *reset.ExternalID != externalID {
		t.Fatalf("externalID = %v, want %s", reset.ExternalID, externalID)
	}
}

// A failure must never blank an external ID recorded by an earlier success.
func TestMarkInstanceFailedKeepsExternalID(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	teamID := uuid.New()
	instance, err := testDB.Queries.CreateResourceInstance(context.Background(), db.CreateResourceInstanceParams{
		ResourceConfigID: cfg.ID,
		CoursePhaseID:    coursePhaseID,
		TeamID:           &teamID,
	})
	if err != nil {
		t.Fatalf("create instance: %v", err)
	}

	externalID := "42"
	externalURL := "https://gitlab.example.com/groups/team-a"
	if err := testDB.Queries.MarkInstanceCreated(context.Background(), db.MarkInstanceCreatedParams{
		ID:          instance.ID,
		ExternalID:  &externalID,
		ExternalUrl: &externalURL,
	}); err != nil {
		t.Fatalf("mark created: %v", err)
	}

	message := "later failure"
	if err := testDB.Queries.MarkInstanceFailed(context.Background(), db.MarkInstanceFailedParams{
		ID:           instance.ID,
		ErrorMessage: &message,
	}); err != nil {
		t.Fatalf("mark failed: %v", err)
	}

	got, err := testDB.Queries.GetResourceInstance(context.Background(), db.GetResourceInstanceParams{
		ID:            instance.ID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		t.Fatalf("get instance: %v", err)
	}
	if got.ExternalID == nil || *got.ExternalID != externalID {
		t.Fatalf("externalID after failure = %v, want %s", got.ExternalID, externalID)
	}
}

func TestDeleteAndRetryAreScopedByCoursePhase(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	otherCoursePhaseID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	teamID := uuid.New()
	instance, err := testDB.Queries.CreateResourceInstance(context.Background(), db.CreateResourceInstanceParams{
		ResourceConfigID: cfg.ID,
		CoursePhaseID:    coursePhaseID,
		TeamID:           &teamID,
	})
	if err != nil {
		t.Fatalf("create instance: %v", err)
	}

	message := "boom"
	if err := testDB.Queries.MarkInstanceFailed(context.Background(), db.MarkInstanceFailedParams{
		ID:           instance.ID,
		ErrorMessage: &message,
	}); err != nil {
		t.Fatalf("mark failed: %v", err)
	}

	if _, err := testDB.Queries.ResetInstanceToPending(context.Background(), db.ResetInstanceToPendingParams{
		ID:            instance.ID,
		CoursePhaseID: otherCoursePhaseID,
	}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("retry from wrong phase = %v, want pgx.ErrNoRows", err)
	}

	reset, err := testDB.Queries.ResetInstanceToPending(context.Background(), db.ResetInstanceToPendingParams{
		ID:            instance.ID,
		CoursePhaseID: coursePhaseID,
	})
	if err != nil {
		t.Fatalf("retry correct phase: %v", err)
	}
	if reset.Status != db.ResourceStatusPending {
		t.Fatalf("status after correct retry = %s, want pending", reset.Status)
	}

	if err := testDB.Queries.DeleteResourceInstance(context.Background(), db.DeleteResourceInstanceParams{
		ID:            instance.ID,
		CoursePhaseID: otherCoursePhaseID,
	}); err != nil {
		t.Fatalf("delete wrong phase: %v", err)
	}
	instances, err := testDB.Queries.ListResourceInstances(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("instances after wrong-phase delete = %d, want 1", len(instances))
	}

	if err := testDB.Queries.DeleteResourceInstance(context.Background(), db.DeleteResourceInstanceParams{
		ID:            instance.ID,
		CoursePhaseID: coursePhaseID,
	}); err != nil {
		t.Fatalf("delete correct phase: %v", err)
	}
	instances, err = testDB.Queries.ListResourceInstances(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("list instances after delete: %v", err)
	}
	if len(instances) != 0 {
		t.Fatalf("instances after correct delete = %d, want 0", len(instances))
	}
}
