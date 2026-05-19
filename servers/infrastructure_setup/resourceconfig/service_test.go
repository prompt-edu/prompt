package resourceconfig

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

func setupResourceConfigTestDB(t *testing.T) (*sdkTestUtils.TestDB[*db.Queries], func()) {
	t.Helper()

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(context.Background(), "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries {
		return db.New(conn)
	})
	if err != nil {
		t.Fatalf("setup test db: %v", err)
	}
	return testDB, cleanup
}

func createProviderForResourceConfigTest(t *testing.T, queries *db.Queries, coursePhaseID uuid.UUID) {
	t.Helper()
	if _, err := queries.UpsertProviderConfig(context.Background(), db.UpsertProviderConfigParams{
		CoursePhaseID: coursePhaseID,
		ProviderType:  db.ProviderTypeGitlab,
		Credentials:   []byte("encrypted"),
	}); err != nil {
		t.Fatalf("upsert provider config: %v", err)
	}
}

func TestResourceConfigCRUD(t *testing.T) {
	testDB, cleanup := setupResourceConfigTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	createProviderForResourceConfigTest(t, testDB.Queries, coursePhaseID)
	service := NewService(testDB.Conn)

	created, err := service.CreateResourceConfig(context.Background(), coursePhaseID, CreateRequest{
		ProviderType:        "gitlab",
		ResourceType:        "group",
		Scope:               "per_team",
		NameTemplate:        "{{teamName}}",
		PermissionMapping:   map[string]string{"student": "developer"},
		ResourceExtraConfig: map[string]interface{}{"visibility": "private"},
	})
	if err != nil {
		t.Fatalf("CreateResourceConfig returned error: %v", err)
	}
	if created.ProviderType != "gitlab" || created.Scope != "per_team" {
		t.Fatalf("created config = %+v, want gitlab per_team", created)
	}

	list, err := service.ListResourceConfigs(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("ListResourceConfigs returned error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("resource configs = %d, want 1", len(list))
	}

	updated, err := service.UpdateResourceConfig(context.Background(), coursePhaseID, created.ID, UpdateRequest{
		ResourceType:        "group",
		Scope:               "per_student",
		NameTemplate:        "{{studentLogin}}",
		PermissionMapping:   map[string]string{"student": "maintainer"},
		ResourceExtraConfig: map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("UpdateResourceConfig returned error: %v", err)
	}
	if updated.Scope != "per_student" || updated.NameTemplate != "{{studentLogin}}" {
		t.Fatalf("updated config = %+v, want per_student student login template", updated)
	}

	got, err := service.GetResourceConfig(context.Background(), coursePhaseID, created.ID)
	if err != nil {
		t.Fatalf("GetResourceConfig returned error: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("got ID = %s, want %s", got.ID, created.ID)
	}

	if err := service.DeleteResourceConfig(context.Background(), coursePhaseID, created.ID); err != nil {
		t.Fatalf("DeleteResourceConfig returned error: %v", err)
	}
	list, err = service.ListResourceConfigs(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("ListResourceConfigs after delete returned error: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("resource configs after delete = %d, want 0", len(list))
	}
}

func TestCreateResourceConfigRequiresExistingProvider(t *testing.T) {
	testDB, cleanup := setupResourceConfigTestDB(t)
	defer cleanup()

	service := NewService(testDB.Conn)
	_, err := service.CreateResourceConfig(context.Background(), uuid.New(), CreateRequest{
		ProviderType: "gitlab",
		ResourceType: "group",
		Scope:        "per_team",
		NameTemplate: "{{teamName}}",
	})
	if err == nil {
		t.Fatal("CreateResourceConfig returned nil error without provider config")
	}
}
