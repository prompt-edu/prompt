package copy

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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
	return testDB, cleanup
}

// The endpoint is exercised through the SDK registration rather than by calling the
// method: the SDK writes the response itself, so only the real route catches a handler
// that writes one of its own on top.
func newCopyTestRouter(svc *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router.Group("/api"), svc, func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockPermissionMiddleware(allowedRoles...)
	})
	return router
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

	copyPhase(t, newCopyTestRouter(NewService(testDB.Conn)), sourceID, targetID)

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

// Copying the same source twice must not duplicate the resource configs, which would
// provision every resource twice.
func TestHandlePhaseCopyIsIdempotent(t *testing.T) {
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
		PermissionMapping:   []byte(`{}`),
		ResourceExtraConfig: []byte(`{}`),
	}); err != nil {
		t.Fatalf("create source resource config: %v", err)
	}

	router := newCopyTestRouter(NewService(testDB.Conn))
	copyPhase(t, router, sourceID, targetID)
	copyPhase(t, router, sourceID, targetID)

	copied, err := testDB.Queries.ListResourceConfigs(context.Background(), targetID)
	if err != nil {
		t.Fatalf("list copied resource configs: %v", err)
	}
	if len(copied) != 1 {
		t.Fatalf("copied resource configs after two copies = %d, want 1", len(copied))
	}
}

// copyPhase drives POST /copy and asserts the response body is exactly one JSON
// object: the SDK already answers for both outcomes, so a handler that also writes
// leaves two concatenated objects that core cannot parse.
func copyPhase(t *testing.T, router *gin.Engine, sourceID, targetID uuid.UUID) {
	t.Helper()

	body, err := json.Marshal(map[string]string{
		"sourceCoursePhaseID": sourceID.String(),
		"targetCoursePhaseID": targetID.String(),
	})
	if err != nil {
		t.Fatalf("marshal copy request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/copy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("copy status = %d, want %d: %s", resp.Code, http.StatusOK, resp.Body.String())
	}

	decoder := json.NewDecoder(bytes.NewReader(resp.Body.Bytes()))
	var first map[string]any
	if err := decoder.Decode(&first); err != nil {
		t.Fatalf("decode copy response %q: %v", resp.Body.String(), err)
	}
	if decoder.More() {
		t.Fatalf("copy response holds more than one JSON object: %s", resp.Body.String())
	}
}
