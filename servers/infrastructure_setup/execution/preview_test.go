package execution

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

// The preview is what the provisioning page shows before anyone clicks, so it has to
// give the trigger's answer and must not act on it.
func TestPreviewReportsWhatATriggerWouldDo(t *testing.T) {
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

	if err := createInstances(t, service, coursePhaseID, cfg, targets[:1]); err != nil {
		t.Fatalf("create instances: %v", err)
	}
	existing, err := testDB.Queries.ListResourceInstances(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	markInstance(t, testDB.Queries, existing[0].ID, db.ResourceStatusFailed)

	preview, err := service.PreviewExecution(context.Background(), "Bearer test", coursePhaseID)
	if err != nil {
		t.Fatalf("PreviewExecution: %v", err)
	}
	if preview.Queued != 1 || preview.Requeued != 1 || preview.UpToDate != 0 || preview.Running != 0 {
		t.Fatalf("preview = %+v, want Team B queued and Team A's failure retried", preview)
	}
	if preview.Teams == nil || *preview.Teams != 2 {
		t.Fatalf("teams = %v, want both resolved teams counted", preview.Teams)
	}
	if preview.Students != nil {
		t.Fatalf("students = %v, want null: no config is per_student", *preview.Students)
	}

	after, err := testDB.Queries.ListResourceInstances(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	if len(after) != 1 || after[0].Status != db.ResourceStatusFailed {
		t.Fatalf("instances after preview = %+v, want the one failed instance untouched", after)
	}
}

// While a run is going a trigger is refused, which the page has to be able to say.
func TestPreviewCountsRunningInstances(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	teamID := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	targets := []ProvisioningTarget{{Scope: db.ResourceScopePerTeam, TeamID: &teamID, TeamName: "Team A"}}
	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{targets: targets})

	if err := createInstances(t, service, coursePhaseID, cfg, targets); err != nil {
		t.Fatalf("create instances: %v", err)
	}

	preview, err := service.PreviewExecution(context.Background(), "Bearer test", coursePhaseID)
	if err != nil {
		t.Fatalf("PreviewExecution: %v", err)
	}
	if preview.Running != 1 {
		t.Fatalf("running = %d, want the pending instance counted", preview.Running)
	}
}

func TestPreviewRefusesForTheReasonsATriggerDoes(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{})
	if _, err := service.PreviewExecution(context.Background(), "Bearer test", uuid.New()); !errors.Is(err, ErrNothingConfigured) {
		t.Fatalf("err = %v, want ErrNothingConfigured", err)
	}

	coursePhaseID := uuid.New()
	createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	unwired := NewServiceWithResolver(testDB.Conn, failingTargetResolver{err: ErrTeamsNotWired})
	if _, err := unwired.PreviewExecution(context.Background(), "Bearer test", coursePhaseID); !errors.Is(err, ErrTeamsNotWired) {
		t.Fatalf("err = %v, want ErrTeamsNotWired", err)
	}
}
