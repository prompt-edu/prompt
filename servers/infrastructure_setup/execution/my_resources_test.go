package execution

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

func createConfig(t *testing.T, queries *db.Queries, coursePhaseID uuid.UUID, providerType db.ProviderType, scope db.ResourceScope, nameTemplate string) db.ResourceConfig {
	t.Helper()
	if _, err := queries.UpsertProviderConfig(context.Background(), db.UpsertProviderConfigParams{
		CoursePhaseID: coursePhaseID,
		ProviderType:  providerType,
		Credentials:   []byte("encrypted"),
	}); err != nil {
		t.Fatalf("upsert provider config: %v", err)
	}
	cfg, err := queries.CreateResourceConfig(context.Background(), db.CreateResourceConfigParams{
		CoursePhaseID:       coursePhaseID,
		ProviderType:        providerType,
		ResourceType:        "group",
		Scope:               scope,
		NameTemplate:        nameTemplate,
		PermissionMapping:   []byte(`{"student":"developer"}`),
		ResourceExtraConfig: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("create resource config: %v", err)
	}
	return cfg
}

func recordMembers(t *testing.T, queries *db.Queries, instanceID uuid.UUID, granted map[uuid.UUID]bool) {
	t.Helper()
	ids := make([]uuid.UUID, 0, len(granted))
	flags := make([]bool, 0, len(granted))
	for id, flag := range granted {
		ids = append(ids, id)
		flags = append(flags, flag)
	}
	if err := queries.InsertInstanceMembers(context.Background(), db.InsertInstanceMembersParams{
		ResourceInstanceID:     instanceID,
		CourseParticipationIds: ids,
		Granted:                flags,
	}); err != nil {
		t.Fatalf("record members: %v", err)
	}
}

func instanceFor(t *testing.T, queries *db.Queries, coursePhaseID, configID uuid.UUID, target uuid.UUID) db.ResourceInstance {
	t.Helper()
	instances, err := queries.ListResourceInstances(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("list instances: %v", err)
	}
	for _, instance := range instances {
		if instance.ResourceConfigID != configID {
			continue
		}
		if (instance.TeamID != nil && *instance.TeamID == target) ||
			(instance.CourseParticipationID != nil && *instance.CourseParticipationID == target) {
			return instance
		}
	}
	t.Fatalf("no instance of config %s for %s", configID, target)
	return db.ResourceInstance{}
}

// A student sees their own resource, their team's resource with whether they were let
// in, and a configured resource nothing was provisioned for yet. Another team's resource
// and the teammate's personal one stay out of it.
func TestListMyResourcesShowsTheStudentsOwnAndTeamResources(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()
	ctx := context.Background()

	coursePhaseID := uuid.New()
	me, teammate, stranger := uuid.New(), uuid.New(), uuid.New()
	myTeam, otherTeam := uuid.New(), uuid.New()

	personal := createConfig(t, testDB.Queries, coursePhaseID, db.ProviderTypeGitlab, db.ResourceScopePerStudent, "{{studentLogin}}")
	team := createConfig(t, testDB.Queries, coursePhaseID, db.ProviderTypeGitlab, db.ResourceScopePerTeam, "{{teamName}}")
	later := createConfig(t, testDB.Queries, coursePhaseID, db.ProviderTypeGitlab, db.ResourceScopePerTeam, "{{teamName}}-app")

	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{})
	if err := createInstances(t, service, coursePhaseID, personal, []ProvisioningTarget{
		{Scope: db.ResourceScopePerStudent, CourseParticipationID: &me},
		{Scope: db.ResourceScopePerStudent, CourseParticipationID: &teammate},
	}); err != nil {
		t.Fatalf("create personal instances: %v", err)
	}
	// A trigger refuses to queue while anything is pending, so the personal run finishes
	// first.
	mine := instanceFor(t, testDB.Queries, coursePhaseID, personal.ID, me)
	markInstance(t, testDB.Queries, mine.ID, db.ResourceStatusCreated)
	recordMembers(t, testDB.Queries, mine.ID, map[uuid.UUID]bool{me: true})
	markInstance(t, testDB.Queries, instanceFor(t, testDB.Queries, coursePhaseID, personal.ID, teammate).ID, db.ResourceStatusCreated)

	if err := createInstances(t, service, coursePhaseID, team, []ProvisioningTarget{
		{Scope: db.ResourceScopePerTeam, TeamID: &myTeam, TeamName: "Team A"},
		{Scope: db.ResourceScopePerTeam, TeamID: &otherTeam, TeamName: "Team B"},
	}); err != nil {
		t.Fatalf("create team instances: %v", err)
	}

	ours := instanceFor(t, testDB.Queries, coursePhaseID, team.ID, myTeam)
	markInstance(t, testDB.Queries, ours.ID, db.ResourceStatusPartial)
	recordMembers(t, testDB.Queries, ours.ID, map[uuid.UUID]bool{me: false, teammate: true})

	theirs := instanceFor(t, testDB.Queries, coursePhaseID, team.ID, otherTeam)
	markInstance(t, testDB.Queries, theirs.ID, db.ResourceStatusCreated)
	recordMembers(t, testDB.Queries, theirs.ID, map[uuid.UUID]bool{stranger: true})

	resources, err := service.ListMyResources(ctx, coursePhaseID, me)
	if err != nil {
		t.Fatalf("ListMyResources: %v", err)
	}
	if len(resources) != 3 {
		t.Fatalf("resources = %+v, want my own, my team's, and the unprovisioned one", resources)
	}

	byConfig := make(map[uuid.UUID]MyResourceResponse, len(resources))
	for _, resource := range resources {
		byConfig[resource.ResourceConfigID] = resource
	}

	own := byConfig[personal.ID]
	if own.Status == nil || *own.Status != db.ResourceStatusCreated || own.Granted == nil || !*own.Granted {
		t.Fatalf("own resource = %+v, want created and granted", own)
	}
	if own.URL == nil || *own.URL == "" {
		t.Fatalf("own resource url = %v, want the GitLab link", own.URL)
	}

	shared := byConfig[team.ID]
	if shared.Status == nil || *shared.Status != db.ResourceStatusPartial {
		t.Fatalf("team resource = %+v, want my team's partial instance, not the other team's", shared)
	}
	if shared.Granted == nil || *shared.Granted {
		t.Fatalf("team resource granted = %v, want false: I was the one left out", shared.Granted)
	}
	if shared.TeamName != "Team A" {
		t.Fatalf("team name = %q, want my team", shared.TeamName)
	}

	pending := byConfig[later.ID]
	if pending.Status != nil || pending.Granted != nil || pending.URL != nil {
		t.Fatalf("unprovisioned resource = %+v, want no status, access or link", pending)
	}
}

// A Keycloak group's link is the admin console, which a student cannot open.
func TestListMyResourcesHidesTheKeycloakAdminLink(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	me := uuid.New()
	cfg := createConfig(t, testDB.Queries, coursePhaseID, db.ProviderTypeKeycloak, db.ResourceScopePerStudent, "{{studentLogin}}")
	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{})
	if err := createInstances(t, service, coursePhaseID, cfg, []ProvisioningTarget{
		{Scope: db.ResourceScopePerStudent, CourseParticipationID: &me},
	}); err != nil {
		t.Fatalf("create instances: %v", err)
	}
	markInstance(t, testDB.Queries, instanceFor(t, testDB.Queries, coursePhaseID, cfg.ID, me).ID, db.ResourceStatusCreated)

	resources, err := service.ListMyResources(context.Background(), coursePhaseID, me)
	if err != nil {
		t.Fatalf("ListMyResources: %v", err)
	}
	if len(resources) != 1 || resources[0].Status == nil {
		t.Fatalf("resources = %+v, want the one provisioned group", resources)
	}
	if resources[0].URL != nil {
		t.Fatalf("url = %q, want no link to the admin console", *resources[0].URL)
	}
}

// The route answers for the caller's own participation, which the SDK middleware puts on
// the context; without it the caller is not a student of the phase.
func TestMyResourcesRouteAnswersForTheCallersParticipation(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	me := uuid.New()
	cfg := createConfig(t, testDB.Queries, coursePhaseID, db.ProviderTypeGitlab, db.ResourceScopePerStudent, "{{studentLogin}}")
	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{})
	if err := createInstances(t, service, coursePhaseID, cfg, []ProvisioningTarget{
		{Scope: db.ResourceScopePerStudent, CourseParticipationID: &me},
	}); err != nil {
		t.Fatalf("create instances: %v", err)
	}
	markInstance(t, testDB.Queries, instanceFor(t, testDB.Queries, coursePhaseID, cfg.ID, me).ID, db.ResourceStatusPartial)

	gin.SetMode(gin.TestMode)
	path := "/api/course_phase/" + coursePhaseID.String() + "/my-resources"

	anonymous := gin.New()
	RegisterStudentRoutes(anonymous.Group("/api/course_phase/:coursePhaseID"), service)
	resp := httptest.NewRecorder()
	anonymous.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, path, nil))
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("status without a participation = %d, want 401: %s", resp.Code, resp.Body.String())
	}

	student := gin.New()
	group := student.Group("/api/course_phase/:coursePhaseID", func(c *gin.Context) {
		c.Set("courseParticipationID", me)
	})
	RegisterStudentRoutes(group, service)
	resp = httptest.NewRecorder()
	student.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, path, nil))
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", resp.Code, resp.Body.String())
	}
	// A partial run's message names the other members who were left out.
	if strings.Contains(resp.Body.String(), "could not be added") {
		t.Fatalf("response = %s, want no error text for a student", resp.Body.String())
	}
	var got []MyResourceResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 1 || got[0].Status == nil || *got[0].Status != db.ResourceStatusPartial {
		t.Fatalf("resources = %+v, want the caller's partial instance", got)
	}
}

// A team's instance is its members' from the moment it is queued. Without that, a team
// waiting for a free worker would read as never provisioned while a student with a
// personal resource in the same state already sees it being set up.
func TestQueuedTeamInstanceIsVisibleToItsMembers(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	me, teamID := uuid.New(), uuid.New()
	cfg := createConfig(t, testDB.Queries, coursePhaseID, db.ProviderTypeGitlab, db.ResourceScopePerTeam, "{{teamName}}")
	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{})

	if err := createInstances(t, service, coursePhaseID, cfg, []ProvisioningTarget{{
		Scope:    db.ResourceScopePerTeam,
		TeamID:   &teamID,
		TeamName: "Team A",
		People:   []TargetPerson{{CourseParticipationID: me, Email: "me@example.com"}},
	}}); err != nil {
		t.Fatalf("create instances: %v", err)
	}

	resources, err := service.ListMyResources(context.Background(), coursePhaseID, me)
	if err != nil {
		t.Fatalf("ListMyResources: %v", err)
	}
	if len(resources) != 1 || resources[0].Status == nil || *resources[0].Status != db.ResourceStatusPending {
		t.Fatalf("resources = %+v, want my team's instance, pending", resources)
	}
	if resources[0].Granted == nil || *resources[0].Granted {
		t.Fatalf("granted = %v, want false until the run has let anyone in", resources[0].Granted)
	}
	if resources[0].TeamName != "Team A" {
		t.Fatalf("team name = %q, want my team", resources[0].TeamName)
	}
}
