package execution

import (
	"context"
	"testing"

	"github.com/google/uuid"
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

func (f fakeTargetResolver) ResolveInstanceTarget(_ context.Context, _ string, instance db.ResourceInstance) (ProvisioningTarget, error) {
	for _, target := range f.targets {
		if instance.TeamID != nil && target.TeamID != nil && *instance.TeamID == *target.TeamID {
			return target, nil
		}
		if instance.CourseParticipationID != nil && target.CourseParticipationID != nil && *instance.CourseParticipationID == *target.CourseParticipationID {
			return target, nil
		}
	}
	return ProvisioningTarget{}, nil
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

func TestCreateInstancesForConfigCreatesOneInstancePerTeam(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	teamA := uuid.New()
	teamB := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{targets: []ProvisioningTarget{
		{Scope: db.ResourceScopePerTeam, TeamID: &teamA, TeamName: "Team A"},
		{Scope: db.ResourceScopePerTeam, TeamID: &teamB, TeamName: "Team B"},
	}})

	if err := service.createInstancesForConfig(context.Background(), "Bearer test", coursePhaseID, cfg); err != nil {
		t.Fatalf("create instances: %v", err)
	}
	if err := service.createInstancesForConfig(context.Background(), "Bearer test", coursePhaseID, cfg); err != nil {
		t.Fatalf("create duplicate instances: %v", err)
	}

	instances, err := testDB.Queries.ListResourceInstances(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("instances = %d, want 2", len(instances))
	}
}

func TestCreateInstancesForConfigCreatesOneInstancePerStudent(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	participationA := uuid.New()
	participationB := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerStudent)
	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{targets: []ProvisioningTarget{
		{Scope: db.ResourceScopePerStudent, CourseParticipationID: &participationA},
		{Scope: db.ResourceScopePerStudent, CourseParticipationID: &participationB},
	}})

	if err := service.createInstancesForConfig(context.Background(), "Bearer test", coursePhaseID, cfg); err != nil {
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

	if err := testDB.Queries.UpdateResourceInstanceStatus(context.Background(), db.UpdateResourceInstanceStatusParams{
		ID:     instance.ID,
		Status: db.ResourceStatusFailed,
	}); err != nil {
		t.Fatalf("mark failed: %v", err)
	}

	if err := testDB.Queries.ResetFailedInstanceToPending(context.Background(), instance.ID, otherCoursePhaseID); err != nil {
		t.Fatalf("retry wrong phase: %v", err)
	}
	got, err := testDB.Queries.GetResourceInstance(context.Background(), instance.ID, coursePhaseID)
	if err != nil {
		t.Fatalf("get instance after wrong retry: %v", err)
	}
	if got.Status != db.ResourceStatusFailed {
		t.Fatalf("status after wrong-phase retry = %s, want failed", got.Status)
	}

	if err := testDB.Queries.ResetFailedInstanceToPending(context.Background(), instance.ID, coursePhaseID); err != nil {
		t.Fatalf("retry correct phase: %v", err)
	}
	got, err = testDB.Queries.GetResourceInstance(context.Background(), instance.ID, coursePhaseID)
	if err != nil {
		t.Fatalf("get instance after correct retry: %v", err)
	}
	if got.Status != db.ResourceStatusPending {
		t.Fatalf("status after correct retry = %s, want pending", got.Status)
	}

	if err := testDB.Queries.DeleteResourceInstance(context.Background(), instance.ID, otherCoursePhaseID); err != nil {
		t.Fatalf("delete wrong phase: %v", err)
	}
	instances, err := testDB.Queries.ListResourceInstances(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("instances after wrong-phase delete = %d, want 1", len(instances))
	}

	if err := testDB.Queries.DeleteResourceInstance(context.Background(), instance.ID, coursePhaseID); err != nil {
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
