package coursePhaseDeletion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

func setupDeletionTestDB(t *testing.T) (*sdkTestUtils.TestDB[*db.Queries], func()) {
	t.Helper()

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(context.Background(), "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries {
		return db.New(conn)
	})
	if err != nil {
		t.Fatalf("setup test db: %v", err)
	}
	return testDB, cleanup
}

// deletionRequest builds the context the SDK endpoint hands the handler.
func deletionRequest() *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/delete", nil)
	return c
}

// seedPhase fills every table the service owns for one phase.
func seedPhase(t *testing.T, queries *db.Queries, coursePhaseID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	teamID := uuid.New()

	if _, err := queries.UpsertCoursePhaseConfig(ctx, db.UpsertCoursePhaseConfigParams{
		CoursePhaseID: coursePhaseID,
		SemesterTag:   "ios26",
	}); err != nil {
		t.Fatalf("upsert phase config: %v", err)
	}
	if _, err := queries.UpsertProviderConfig(ctx, db.UpsertProviderConfigParams{
		CoursePhaseID: coursePhaseID,
		ProviderType:  db.ProviderTypeGitlab,
		Credentials:   []byte("encrypted-credentials"),
	}); err != nil {
		t.Fatalf("upsert provider config: %v", err)
	}
	config, err := queries.CreateResourceConfig(ctx, db.CreateResourceConfigParams{
		CoursePhaseID:       coursePhaseID,
		ProviderType:        db.ProviderTypeGitlab,
		ResourceType:        "group",
		Scope:               db.ResourceScopePerTeam,
		NameTemplate:        "{{teamName}}",
		PermissionMapping:   []byte(`{"student":"developer"}`),
		ResourceExtraConfig: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("create resource config: %v", err)
	}
	if _, err := queries.CreateResourceInstance(ctx, db.CreateResourceInstanceParams{
		ResourceConfigID: config.ID,
		CoursePhaseID:    coursePhaseID,
		TeamID:           &teamID,
	}); err != nil {
		t.Fatalf("create resource instance: %v", err)
	}
}

// countRows counts the raw tables rather than going through the generated queries: a
// query that joins its parent could not tell a cascaded delete from an orphaned row.
func countRows(t *testing.T, pool *pgxpool.Pool, table string, coursePhaseID uuid.UUID) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(context.Background(),
		"SELECT count(*) FROM "+table+" WHERE course_phase_id = $1", coursePhaseID,
	).Scan(&count); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return count
}

func TestDeletionRemovesEveryTableOfThePhase(t *testing.T) {
	testDB, cleanup := setupDeletionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	otherPhaseID := uuid.New()
	seedPhase(t, testDB.Queries, coursePhaseID)
	seedPhase(t, testDB.Queries, otherPhaseID)

	service := NewCoursePhaseDeletionService(testDB.Conn)
	if err := service.HandleCoursePhaseDeletion(deletionRequest(), coursePhaseID); err != nil {
		t.Fatalf("HandleCoursePhaseDeletion: %v", err)
	}

	for _, table := range []string{"resource_instance", "resource_config", "provider_config", "course_phase_config"} {
		if got := countRows(t, testDB.Conn, table, coursePhaseID); got != 0 {
			t.Fatalf("%s rows after deletion = %d, want 0", table, got)
		}
		// Another phase's data, in particular its stored credentials, must survive.
		if got := countRows(t, testDB.Conn, table, otherPhaseID); got != 1 {
			t.Fatalf("%s rows of the untouched phase = %d, want 1", table, got)
		}
	}
}

// Core calls this endpoint whether or not the phase was ever configured, and may retry
// it, so the handler has to be idempotent.
func TestDeletionIsIdempotent(t *testing.T) {
	testDB, cleanup := setupDeletionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	seedPhase(t, testDB.Queries, coursePhaseID)

	service := NewCoursePhaseDeletionService(testDB.Conn)
	if err := service.HandleCoursePhaseDeletion(deletionRequest(), coursePhaseID); err != nil {
		t.Fatalf("first deletion: %v", err)
	}
	if err := service.HandleCoursePhaseDeletion(deletionRequest(), coursePhaseID); err != nil {
		t.Fatalf("second deletion: %v", err)
	}
	if err := service.HandleCoursePhaseDeletion(deletionRequest(), uuid.New()); err != nil {
		t.Fatalf("deletion of a phase without data: %v", err)
	}
}
