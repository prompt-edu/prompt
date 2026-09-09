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

// queueInstances runs the transactional part of a trigger without spawning the
// background worker, which would need a real provider.
func queueInstances(t *testing.T, service *Service, coursePhaseID uuid.UUID, cfg db.ResourceConfig, targets []ProvisioningTarget) (TriggerSummary, error) {
	t.Helper()
	return service.queueInstances(
		context.Background(),
		coursePhaseID,
		[]db.ResourceConfig{cfg},
		map[db.ResourceScope][]ProvisioningTarget{cfg.Scope: targets},
	)
}

func createInstances(t *testing.T, service *Service, coursePhaseID uuid.UUID, cfg db.ResourceConfig, targets []ProvisioningTarget) error {
	t.Helper()
	_, err := queueInstances(t, service, coursePhaseID, cfg, targets)
	return err
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

// markInstance drives an instance into a terminal state the way the worker would.
func markInstance(t *testing.T, queries *db.Queries, instanceID uuid.UUID, status db.ResourceStatus) {
	t.Helper()
	externalID := "ext-1"
	externalURL := "https://example.test/ext-1"
	message := "one member could not be added"

	var err error
	switch status {
	case db.ResourceStatusFailed:
		err = queries.MarkInstanceFailed(context.Background(), db.MarkInstanceFailedParams{
			ID: instanceID, ErrorMessage: &message,
		})
	case db.ResourceStatusPartial:
		err = queries.MarkInstancePartial(context.Background(), db.MarkInstancePartialParams{
			ID: instanceID, ExternalID: &externalID, ExternalUrl: &externalURL, ErrorMessage: &message,
		})
	case db.ResourceStatusCreated:
		err = queries.MarkInstanceCreated(context.Background(), db.MarkInstanceCreatedParams{
			ID: instanceID, ExternalID: &externalID, ExternalUrl: &externalURL,
		})
	default:
		t.Fatalf("markInstance does not handle %s", status)
	}
	if err != nil {
		t.Fatalf("mark %s: %v", status, err)
	}
}

// The common first outcome of a run is a mix of successes and failures, so triggering
// again has to converge on the rows that are already there: requeue what did not
// finish, leave what did alone, and never put a second row beside an existing one.
func TestTriggerConvergesOnExistingInstances(t *testing.T) {
	for _, tc := range []struct {
		status   db.ResourceStatus
		requeued int
		upToDate int
		want     db.ResourceStatus
	}{
		{status: db.ResourceStatusFailed, requeued: 1, want: db.ResourceStatusPending},
		{status: db.ResourceStatusPartial, requeued: 1, want: db.ResourceStatusPending},
		{status: db.ResourceStatusCreated, upToDate: 1, want: db.ResourceStatusCreated},
	} {
		t.Run(string(tc.status), func(t *testing.T) {
			testDB, cleanup := setupExecutionTestDB(t)
			defer cleanup()

			coursePhaseID := uuid.New()
			teamID := uuid.New()
			cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
			targets := []ProvisioningTarget{{Scope: db.ResourceScopePerTeam, TeamID: &teamID, TeamName: "Team A"}}
			service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{targets: targets})

			first, err := queueInstances(t, service, coursePhaseID, cfg, targets)
			if err != nil {
				t.Fatalf("first trigger: %v", err)
			}
			if first.Queued != 1 {
				t.Fatalf("first trigger queued = %d, want 1", first.Queued)
			}

			instances, err := testDB.Queries.ListResourceInstances(context.Background(), coursePhaseID)
			if err != nil {
				t.Fatalf("list instances: %v", err)
			}
			markInstance(t, testDB.Queries, instances[0].ID, tc.status)

			second, err := queueInstances(t, service, coursePhaseID, cfg, targets)
			if err != nil {
				t.Fatalf("second trigger: %v", err)
			}
			if second.Queued != 0 {
				t.Fatalf("second trigger queued = %d, want 0 new instances", second.Queued)
			}
			if second.Requeued != tc.requeued || second.UpToDate != tc.upToDate {
				t.Fatalf("summary = %+v, want requeued %d and upToDate %d", second, tc.requeued, tc.upToDate)
			}

			after, err := testDB.Queries.ListResourceInstances(context.Background(), coursePhaseID)
			if err != nil {
				t.Fatalf("list instances: %v", err)
			}
			if len(after) != 1 {
				t.Fatalf("instances = %d, want the one row converged on", len(after))
			}
			if after[0].Status != tc.want {
				t.Fatalf("status = %s, want %s", after[0].Status, tc.want)
			}
		})
	}
}

// A trigger that queued nothing must say so. Reporting "execution started" over a no-op
// is what let a lecturer believe a phase full of partial instances was being retried.
func TestTriggerSummaryReportsWhetherWorkWasQueued(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	teamID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	targets := []ProvisioningTarget{{Scope: db.ResourceScopePerTeam, TeamID: &teamID, TeamName: "Team A"}}
	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{targets: targets})

	summary, err := queueInstances(t, service, coursePhaseID, cfg, targets)
	if err != nil {
		t.Fatalf("first trigger: %v", err)
	}
	if !summary.Started() {
		t.Fatalf("summary = %+v, want a started run", summary)
	}

	instances, err := testDB.Queries.ListResourceInstances(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	markInstance(t, testDB.Queries, instances[0].ID, db.ResourceStatusCreated)

	summary, err = queueInstances(t, service, coursePhaseID, cfg, targets)
	if err != nil {
		t.Fatalf("second trigger: %v", err)
	}
	if summary.Started() {
		t.Fatalf("summary = %+v, want nothing queued for an already provisioned phase", summary)
	}
}

// A target new to the phase joins the existing instances rather than waiting for the
// others to be deleted first.
func TestTriggerQueuesOnlyTheNewTarget(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	teamA := uuid.New()
	teamB := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	targets := []ProvisioningTarget{{Scope: db.ResourceScopePerTeam, TeamID: &teamA, TeamName: "Team A"}}
	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{targets: targets})

	if _, err := queueInstances(t, service, coursePhaseID, cfg, targets); err != nil {
		t.Fatalf("first trigger: %v", err)
	}
	instances, err := testDB.Queries.ListResourceInstances(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	markInstance(t, testDB.Queries, instances[0].ID, db.ResourceStatusCreated)

	targets = append(targets, ProvisioningTarget{Scope: db.ResourceScopePerTeam, TeamID: &teamB, TeamName: "Team B"})
	summary, err := queueInstances(t, service, coursePhaseID, cfg, targets)
	if err != nil {
		t.Fatalf("second trigger: %v", err)
	}
	if summary.Queued != 1 || summary.UpToDate != 1 {
		t.Fatalf("summary = %+v, want one queued and one up to date", summary)
	}
}

// The one-instance-per-target rule is the database's, not the trigger's: a second row
// for the same target is what made a stale failed row collide with its replacement the
// moment either was retried.
func TestDatabaseRefusesASecondInstanceForTheSameTarget(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	teamID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)

	first, err := testDB.Queries.CreateResourceInstance(context.Background(), db.CreateResourceInstanceParams{
		ResourceConfigID: cfg.ID,
		CoursePhaseID:    coursePhaseID,
		TeamID:           &teamID,
	})
	if err != nil {
		t.Fatalf("create instance: %v", err)
	}
	markInstance(t, testDB.Queries, first.ID, db.ResourceStatusFailed)

	// ON CONFLICT DO NOTHING makes the second insert a no-op rather than an error.
	if _, err := testDB.Queries.CreateResourceInstance(context.Background(), db.CreateResourceInstanceParams{
		ResourceConfigID: cfg.ID,
		CoursePhaseID:    coursePhaseID,
		TeamID:           &teamID,
	}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("second insert error = %v, want no row inserted", err)
	}

	instances, err := testDB.Queries.ListResourceInstances(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("instances = %d, want 1 even though the first one failed", len(instances))
	}
}

func TestTriggerRejectsAPhaseWithoutResourceConfigs(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{})
	if _, err := service.TriggerExecution(context.Background(), "Bearer test", uuid.New()); !errors.Is(err, ErrNothingConfigured) {
		t.Fatalf("error = %v, want ErrNothingConfigured", err)
	}
}
