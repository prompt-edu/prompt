package resourceconfig

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/resourceconfig/resourceconfigDTO"
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

	created, err := service.CreateResourceConfig(context.Background(), coursePhaseID, resourceconfigDTO.CreateRequest{
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

	updated, err := service.UpdateResourceConfig(context.Background(), coursePhaseID, created.ID, resourceconfigDTO.UpdateRequest{
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

	if err := service.DeleteResourceConfig(context.Background(), coursePhaseID, created.ID, false); err != nil {
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
	_, err := service.CreateResourceConfig(context.Background(), uuid.New(), resourceconfigDTO.CreateRequest{
		ProviderType: "gitlab",
		ResourceType: "group",
		Scope:        "per_team",
		NameTemplate: "{{teamName}}",
	})
	if err == nil {
		t.Fatal("CreateResourceConfig returned nil error without provider config")
	}
}

// A placeholder the scope never fills resolves to an empty string at run time, so every
// team would converge on the same name and share one external resource.
func TestCreateResourceConfigRejectsAPlaceholderTheScopeCannotFill(t *testing.T) {
	testDB, cleanup := setupResourceConfigTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	createProviderForResourceConfigTest(t, testDB.Queries, coursePhaseID)
	service := NewService(testDB.Conn)

	_, err := service.CreateResourceConfig(context.Background(), coursePhaseID, resourceconfigDTO.CreateRequest{
		ProviderType:        "gitlab",
		ResourceType:        "group",
		Scope:               "per_team",
		NameTemplate:        "team-{{studentLogin}}",
		PermissionMapping:   map[string]string{"student": "developer"},
		ResourceExtraConfig: map[string]interface{}{},
	})
	if err == nil {
		t.Fatal("CreateResourceConfig accepted a per_team template using a student placeholder")
	}
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("error = %v, want it to wrap ErrValidation", err)
	}
}

// A GitLab project is created inside the team subgroup named by parent_group_template,
// so a config without it would fail once per instance during a run.
func TestCreateResourceConfigRequiresProviderExtraConfig(t *testing.T) {
	testDB, cleanup := setupResourceConfigTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	createProviderForResourceConfigTest(t, testDB.Queries, coursePhaseID)
	service := NewService(testDB.Conn)

	_, err := service.CreateResourceConfig(context.Background(), coursePhaseID, resourceconfigDTO.CreateRequest{
		ProviderType:        "gitlab",
		ResourceType:        "project",
		Scope:               "per_team",
		NameTemplate:        "app",
		PermissionMapping:   map[string]string{"student": "developer"},
		ResourceExtraConfig: map[string]interface{}{},
	})
	if err == nil {
		t.Fatal("CreateResourceConfig accepted a gitlab project without parent_group_template")
	}
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("error = %v, want it to wrap ErrValidation", err)
	}

	if _, err := service.CreateResourceConfig(context.Background(), coursePhaseID, resourceconfigDTO.CreateRequest{
		ProviderType:        "gitlab",
		ResourceType:        "project",
		Scope:               "per_team",
		NameTemplate:        "app",
		PermissionMapping:   map[string]string{"student": "developer"},
		ResourceExtraConfig: map[string]interface{}{"parent_group_template": "{{semesterTag}}-{{teamName}}"},
	}); err != nil {
		t.Fatalf("CreateResourceConfig with parent_group_template returned error: %v", err)
	}
}

// The identity is unique per phase (migration 0003), so a duplicate is a bad request
// rather than a 500 out of the database.
func TestCreateResourceConfigRejectsADuplicate(t *testing.T) {
	testDB, cleanup := setupResourceConfigTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	createProviderForResourceConfigTest(t, testDB.Queries, coursePhaseID)
	service := NewService(testDB.Conn)

	request := resourceconfigDTO.CreateRequest{
		ProviderType:        "gitlab",
		ResourceType:        "group",
		Scope:               "per_team",
		NameTemplate:        "{{teamName}}",
		PermissionMapping:   map[string]string{"student": "developer"},
		ResourceExtraConfig: map[string]interface{}{},
	}
	if _, err := service.CreateResourceConfig(context.Background(), coursePhaseID, request); err != nil {
		t.Fatalf("CreateResourceConfig returned error: %v", err)
	}

	_, err := service.CreateResourceConfig(context.Background(), coursePhaseID, request)
	if err == nil {
		t.Fatal("CreateResourceConfig accepted a duplicate configuration")
	}
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("error = %v, want it to wrap ErrValidation", err)
	}
}

// The instances are PROMPT's only record of the external resources, which are never
// deleted, so dropping them takes an explicit confirmation.
func TestDeleteResourceConfigWithInstancesNeedsConfirmation(t *testing.T) {
	testDB, cleanup := setupResourceConfigTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	teamID := uuid.New()
	createProviderForResourceConfigTest(t, testDB.Queries, coursePhaseID)
	service := NewService(testDB.Conn)

	created, err := service.CreateResourceConfig(context.Background(), coursePhaseID, resourceconfigDTO.CreateRequest{
		ProviderType:        "gitlab",
		ResourceType:        "group",
		Scope:               "per_team",
		NameTemplate:        "{{teamName}}",
		PermissionMapping:   map[string]string{"student": "developer"},
		ResourceExtraConfig: map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("CreateResourceConfig returned error: %v", err)
	}
	if _, err := testDB.Queries.CreateResourceInstance(context.Background(), db.CreateResourceInstanceParams{
		ResourceConfigID: created.ID,
		CoursePhaseID:    coursePhaseID,
		TeamID:           &teamID,
	}); err != nil {
		t.Fatalf("create instance: %v", err)
	}

	err = service.DeleteResourceConfig(context.Background(), coursePhaseID, created.ID, false)
	if err == nil {
		t.Fatal("DeleteResourceConfig dropped provisioned instances without confirmation")
	}
	if !errors.Is(err, ErrConfirmationRequired) {
		t.Fatalf("error = %v, want it to wrap ErrConfirmationRequired", err)
	}

	if err := service.DeleteResourceConfig(context.Background(), coursePhaseID, created.ID, true); err != nil {
		t.Fatalf("confirmed DeleteResourceConfig returned error: %v", err)
	}
	list, err := service.ListResourceConfigs(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("ListResourceConfigs returned error: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("resource configs after a confirmed delete = %d, want 0", len(list))
	}
}
