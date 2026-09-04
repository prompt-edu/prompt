package config

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

func setupConfigTestDB(t *testing.T) (*sdkTestUtils.TestDB[*db.Queries], func()) {
	t.Helper()

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(context.Background(), "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries {
		return db.New(conn)
	})
	if err != nil {
		t.Fatalf("setup test db: %v", err)
	}
	return testDB, cleanup
}

func TestStatusReportsConfiguredPhase(t *testing.T) {
	testDB, cleanup := setupConfigTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	if _, err := testDB.Queries.UpsertCoursePhaseConfig(context.Background(), db.UpsertCoursePhaseConfigParams{
		CoursePhaseID: coursePhaseID,
		SemesterTag:   "ios26",
	}); err != nil {
		t.Fatalf("upsert course phase config: %v", err)
	}
	if _, err := testDB.Queries.UpsertProviderConfig(context.Background(), db.UpsertProviderConfigParams{
		CoursePhaseID: coursePhaseID,
		ProviderType:  db.ProviderTypeGitlab,
		Credentials:   []byte("encrypted-credentials"),
	}); err != nil {
		t.Fatalf("upsert provider config: %v", err)
	}
	if _, err := testDB.Queries.CreateResourceConfig(context.Background(), db.CreateResourceConfigParams{
		CoursePhaseID:       coursePhaseID,
		ProviderType:        db.ProviderTypeGitlab,
		ResourceType:        "group",
		Scope:               db.ResourceScopePerTeam,
		NameTemplate:        "{{teamName}}",
		PermissionMapping:   []byte(`{"student":"developer"}`),
		ResourceExtraConfig: []byte(`{}`),
	}); err != nil {
		t.Fatalf("create resource config: %v", err)
	}

	status, err := NewService(testDB.Conn).Status(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if !status["semesterTag"] || !status["providerConfig"] || !status["resourceConfig"] {
		t.Fatalf("status = %+v, want all flags true", status)
	}
}

// A phase whose setup page was never saved has no config row at all. That is
// "nothing configured yet", not an error.
func TestStatusReportsMissingConfigRowAsUnconfigured(t *testing.T) {
	testDB, cleanup := setupConfigTestDB(t)
	defer cleanup()

	status, err := NewService(testDB.Conn).Status(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	for key, configured := range status {
		if configured {
			t.Fatalf("%s = true for a phase with no configuration, want false", key)
		}
	}
}

// A copied phase must not look ready: the provider row survives the copy but its
// credentials do not, so nothing can actually be provisioned yet.
func TestStatusReportsCopiedPhaseAsUnconfigured(t *testing.T) {
	testDB, cleanup := setupConfigTestDB(t)
	defer cleanup()

	sourceID := uuid.New()
	targetID := uuid.New()

	if _, err := testDB.Queries.UpsertCoursePhaseConfig(context.Background(), db.UpsertCoursePhaseConfigParams{
		CoursePhaseID: sourceID,
		SemesterTag:   "ios26",
	}); err != nil {
		t.Fatalf("upsert source course phase config: %v", err)
	}
	if _, err := testDB.Queries.UpsertProviderConfig(context.Background(), db.UpsertProviderConfigParams{
		CoursePhaseID: sourceID,
		ProviderType:  db.ProviderTypeGitlab,
		Credentials:   []byte("encrypted-source-credentials"),
	}); err != nil {
		t.Fatalf("upsert source provider: %v", err)
	}
	if _, err := testDB.Queries.CreateResourceConfig(context.Background(), db.CreateResourceConfigParams{
		CoursePhaseID:       sourceID,
		ProviderType:        db.ProviderTypeGitlab,
		ResourceType:        "group",
		Scope:               db.ResourceScopePerTeam,
		NameTemplate:        "{{teamName}}",
		PermissionMapping:   []byte(`{}`),
		ResourceExtraConfig: []byte(`{}`),
	}); err != nil {
		t.Fatalf("create source resource config: %v", err)
	}

	if err := testDB.Queries.CopyCoursePhaseConfig(context.Background(), db.CopyCoursePhaseConfigParams{
		SourceCoursePhaseID: sourceID,
		TargetCoursePhaseID: targetID,
	}); err != nil {
		t.Fatalf("copy course phase config: %v", err)
	}
	if err := testDB.Queries.CopyProviderConfigsWithEmptyCredentials(context.Background(), db.CopyProviderConfigsWithEmptyCredentialsParams{
		SourceCoursePhaseID: sourceID,
		TargetCoursePhaseID: targetID,
	}); err != nil {
		t.Fatalf("copy provider configs: %v", err)
	}
	if err := testDB.Queries.CopyResourceConfigs(context.Background(), db.CopyResourceConfigsParams{
		SourceCoursePhaseID: sourceID,
		TargetCoursePhaseID: targetID,
	}); err != nil {
		t.Fatalf("copy resource configs: %v", err)
	}

	status, err := NewService(testDB.Conn).Status(context.Background(), targetID)
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if status["providerConfig"] {
		t.Fatalf("providerConfig = true for a copied phase with empty credentials, want false")
	}
	// The semester tag and resource configs do carry over, so the lecturer only has to
	// re-enter the secrets.
	if !status["semesterTag"] {
		t.Fatalf("semesterTag = false, want the copied tag to be reported")
	}
	if !status["resourceConfig"] {
		t.Fatalf("resourceConfig = false, want the copied resource configs to be reported")
	}
}
