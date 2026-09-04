package privacy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkAuth "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

// deletionRequest builds what the SDK hands the handler: a request context and the
// subject core resolved.
func deletionRequest(participationIDs ...uuid.UUID) (*gin.Context, sdkAuth.SubjectIdentifiers) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/privacy/data-deletion", nil)
	return c, sdkAuth.SubjectIdentifiers{
		UserID:                 uuid.New(),
		StudentID:              uuid.New(),
		CourseParticipationIDs: participationIDs,
	}
}

func setupPrivacyTestDB(t *testing.T) (*sdkTestUtils.TestDB[*db.Queries], func()) {
	t.Helper()

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(context.Background(), "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries {
		return db.New(conn)
	})
	if err != nil {
		t.Fatalf("setup test db: %v", err)
	}
	return testDB, cleanup
}

// seedStudentAndTeamInstances creates one instance for the subject and one for a team,
// so a test can tell subject data from course data.
func seedStudentAndTeamInstances(t *testing.T, queries *db.Queries, coursePhaseID, participationID, teamID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	if _, err := queries.UpsertProviderConfig(ctx, db.UpsertProviderConfigParams{
		CoursePhaseID: coursePhaseID,
		ProviderType:  db.ProviderTypeGitlab,
		Credentials:   []byte("encrypted"),
	}); err != nil {
		t.Fatalf("upsert provider config: %v", err)
	}

	studentConfig, err := queries.CreateResourceConfig(ctx, db.CreateResourceConfigParams{
		CoursePhaseID:       coursePhaseID,
		ProviderType:        db.ProviderTypeGitlab,
		ResourceType:        "group",
		Scope:               db.ResourceScopePerStudent,
		NameTemplate:        "{{studentLogin}}",
		PermissionMapping:   []byte(`{"student":"developer"}`),
		ResourceExtraConfig: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("create per_student config: %v", err)
	}
	teamConfig, err := queries.CreateResourceConfig(ctx, db.CreateResourceConfigParams{
		CoursePhaseID:       coursePhaseID,
		ProviderType:        db.ProviderTypeGitlab,
		ResourceType:        "group",
		Scope:               db.ResourceScopePerTeam,
		NameTemplate:        "{{teamName}}",
		PermissionMapping:   []byte(`{"student":"developer"}`),
		ResourceExtraConfig: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("create per_team config: %v", err)
	}

	studentInstance, err := queries.CreateResourceInstance(ctx, db.CreateResourceInstanceParams{
		ResourceConfigID:      studentConfig.ID,
		CoursePhaseID:         coursePhaseID,
		CourseParticipationID: &participationID,
	})
	if err != nil {
		t.Fatalf("create student instance: %v", err)
	}
	externalID := "600"
	externalURL := "https://gitlab.test/groups/pv99tum"
	if err := queries.MarkInstanceCreated(ctx, db.MarkInstanceCreatedParams{
		ID:          studentInstance.ID,
		ExternalID:  &externalID,
		ExternalUrl: &externalURL,
	}); err != nil {
		t.Fatalf("mark student instance created: %v", err)
	}

	if _, err := queries.CreateResourceInstance(ctx, db.CreateResourceInstanceParams{
		ResourceConfigID: teamConfig.ID,
		CoursePhaseID:    coursePhaseID,
		TeamID:           &teamID,
	}); err != nil {
		t.Fatalf("create team instance: %v", err)
	}
}

func TestExportReturnsTheSubjectsProvisionedResources(t *testing.T) {
	testDB, cleanup := setupPrivacyTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	participationID := uuid.New()
	seedStudentAndTeamInstances(t, testDB.Queries, coursePhaseID, participationID, uuid.New())

	rows, err := testDB.Queries.GetResourceInstancesByCourseParticipationIDs(context.Background(), []uuid.UUID{participationID})
	if err != nil {
		t.Fatalf("export query: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("exported instances = %d, want only the subject's own", len(rows))
	}
	if rows[0].ExternalUrl == nil || *rows[0].ExternalUrl != "https://gitlab.test/groups/pv99tum" {
		t.Fatalf("externalUrl = %v, want the provisioned resource to be named", rows[0].ExternalUrl)
	}
	// The config is joined in so the export says what was created, not just an id.
	if rows[0].ProviderType != db.ProviderTypeGitlab || rows[0].ResourceType != "group" {
		t.Fatalf("row = %+v, want the provider and resource type resolved", rows[0])
	}
}

// A deletion removes what the phase stored about the subject and leaves the team's
// instance alone: that row belongs to the team, and it is the only record of an
// external resource nothing else will clean up.
func TestDeletionRemovesOnlyTheSubjectsInstances(t *testing.T) {
	testDB, cleanup := setupPrivacyTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	participationID := uuid.New()
	seedStudentAndTeamInstances(t, testDB.Queries, coursePhaseID, participationID, uuid.New())

	service := NewPrivacyService(testDB.Conn)
	c, subject := deletionRequest(participationID)
	if err := service.DataDeletionHandler(c, subject); err != nil {
		t.Fatalf("DataDeletionHandler: %v", err)
	}

	remaining, err := testDB.Queries.ListResourceInstances(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("remaining instances = %d, want the team's one to survive", len(remaining))
	}
	if remaining[0].CourseParticipationID != nil {
		t.Fatalf("remaining instance = %+v, want no per-student instance left", remaining[0])
	}
}

// Deleting a subject the phase never provisioned anything for is a success, not an
// error: core marks the whole request failed if any service reports one.
func TestDeletionOfAnUnknownSubjectSucceeds(t *testing.T) {
	testDB, cleanup := setupPrivacyTestDB(t)
	defer cleanup()

	service := NewPrivacyService(testDB.Conn)
	c, subject := deletionRequest(uuid.New())
	if err := service.DataDeletionHandler(c, subject); err != nil {
		t.Fatalf("DataDeletionHandler for an unknown subject: %v", err)
	}
}
