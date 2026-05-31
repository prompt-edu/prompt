package copy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

func setupCopyTestDB(t *testing.T) (*sdkTestUtils.TestDB[*db.Queries], func()) {
	t.Helper()

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(context.Background(), "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries {
		return db.New(conn)
	})
	if err != nil {
		t.Fatalf("setup test db: %v", err)
	}
	CopyServiceSingleton = &CopyService{queries: testDB.Queries, conn: testDB.Conn}
	return testDB, func() {
		CopyServiceSingleton = nil
		cleanup()
	}
}

func TestHandlePhaseCopyCopiesProviderStubsAndResourceConfigs(t *testing.T) {
	testDB, cleanup := setupCopyTestDB(t)
	defer cleanup()

	sourceID := uuid.New()
	targetID := uuid.New()

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
		PermissionMapping:   []byte(`{"student":"developer"}`),
		ResourceExtraConfig: []byte(`{}`),
	}); err != nil {
		t.Fatalf("create source resource config: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/copy", nil)

	err := (&InfrastructureSetupCopyHandler{}).HandlePhaseCopy(c, promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: sourceID,
		TargetCoursePhaseID: targetID,
	})
	if err != nil {
		t.Fatalf("HandlePhaseCopy returned error: %v", err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
	}

	copiedProvider, err := testDB.Queries.GetProviderConfig(context.Background(), db.GetProviderConfigParams{
		CoursePhaseID: targetID,
		ProviderType:  db.ProviderTypeGitlab,
	})
	if err != nil {
		t.Fatalf("get copied provider config: %v", err)
	}
	if len(copiedProvider.Credentials) != 0 {
		t.Fatalf("copied provider credentials length = %d, want 0", len(copiedProvider.Credentials))
	}

	copiedResources, err := testDB.Queries.ListResourceConfigs(context.Background(), targetID)
	if err != nil {
		t.Fatalf("list copied resource configs: %v", err)
	}
	if len(copiedResources) != 1 {
		t.Fatalf("copied resource configs = %d, want 1", len(copiedResources))
	}
	if copiedResources[0].NameTemplate != "{{teamName}}" {
		t.Fatalf("copied name template = %q, want {{teamName}}", copiedResources[0].NameTemplate)
	}
}

func TestConfigHandlerReportsConfigurationStatus(t *testing.T) {
	testDB, cleanup := setupCopyTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	sourceID := uuid.New()
	if _, err := testDB.Queries.UpsertCoursePhaseConfig(context.Background(), db.UpsertCoursePhaseConfigParams{
		CoursePhaseID:           coursePhaseID,
		TeamSourceCoursePhaseID: &sourceID,
		SemesterTag:             "ios26",
	}); err != nil {
		t.Fatalf("upsert course phase config: %v", err)
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/config", nil)
	c.Params = gin.Params{{Key: "coursePhaseID", Value: coursePhaseID.String()}}

	status, err := (&ConfigHandler{}).HandlePhaseConfig(c)
	if err != nil {
		t.Fatalf("HandlePhaseConfig returned error: %v", err)
	}
	if !status["sourcePhase"] || !status["semesterTag"] {
		t.Fatalf("status = %+v, want configured sourcePhase and semesterTag", status)
	}
}
