package execution

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

func newExecutionTestRouter(svc *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router.Group("/api/course_phase/:coursePhaseID"), svc)
	return router
}

func TestListInstancesRouteReturnsInstances(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	router := newExecutionTestRouter(NewServiceWithResolver(testDB.Conn, fakeTargetResolver{}))

	req := httptest.NewRequest(http.MethodGet, "/api/course_phase/"+coursePhaseID.String()+"/instances", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", resp.Code, http.StatusOK, resp.Body.String())
	}
	var got []ResourceInstanceResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode instances response: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("instances = %d, want 0", len(got))
	}
}

func TestExecutionRoutesRejectInvalidIDs(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	router := newExecutionTestRouter(NewServiceWithResolver(testDB.Conn, fakeTargetResolver{}))
	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/course_phase/not-a-uuid/instances"},
		{method: http.MethodPost, path: "/api/course_phase/not-a-uuid/execute"},
		{method: http.MethodGet, path: "/api/course_phase/not-a-uuid/execute/preview"},
		{method: http.MethodPost, path: "/api/course_phase/" + uuid.New().String() + "/instances/not-a-uuid/retry"},
		{method: http.MethodDelete, path: "/api/course_phase/" + uuid.New().String() + "/instances/not-a-uuid"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			if resp.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d: %s", resp.Code, http.StatusBadRequest, resp.Body.String())
			}
		})
	}
}

// A phase no teams reach is a setup mistake the lecturer can fix, so both the trigger
// and the preview answer 400 with the reason instead of a 500.
func TestTriggerAndPreviewReportUnwiredTeamsAsBadRequest(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	router := newExecutionTestRouter(NewServiceWithResolver(testDB.Conn, failingTargetResolver{err: ErrTeamsNotWired}))

	for _, tt := range []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/api/course_phase/" + coursePhaseID.String() + "/execute"},
		{method: http.MethodGet, path: "/api/course_phase/" + coursePhaseID.String() + "/execute/preview"},
	} {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, httptest.NewRequest(tt.method, tt.path, nil))
			if resp.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d: %s", resp.Code, http.StatusBadRequest, resp.Body.String())
			}
			if !strings.Contains(resp.Body.String(), "course configurator") {
				t.Fatalf("body = %s, want the reason and where to fix it", resp.Body.String())
			}
		})
	}
}

// The participants page matches students to resources through the members the list
// carries.
func TestListInstancesRouteCarriesMembers(t *testing.T) {
	testDB, cleanup := setupExecutionTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	teamID := uuid.New()
	member := uuid.New()
	cfg := createResourceConfig(t, testDB.Queries, coursePhaseID, db.ResourceScopePerTeam)
	service := NewServiceWithResolver(testDB.Conn, fakeTargetResolver{})
	if err := createInstances(t, service, coursePhaseID, cfg, []ProvisioningTarget{
		{Scope: db.ResourceScopePerTeam, TeamID: &teamID, TeamName: "Team A"},
	}); err != nil {
		t.Fatalf("create instances: %v", err)
	}
	recordMembers(t, testDB.Queries, instanceFor(t, testDB.Queries, coursePhaseID, cfg.ID, teamID).ID, map[uuid.UUID]bool{member: true})

	resp := httptest.NewRecorder()
	newExecutionTestRouter(service).ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/api/course_phase/"+coursePhaseID.String()+"/instances", nil))
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", resp.Code, resp.Body.String())
	}
	var got []ResourceInstanceResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode instances response: %v", err)
	}
	if len(got) != 1 || len(got[0].Members) != 1 || got[0].Members[0].CourseParticipationID != member || !got[0].Members[0].Granted {
		t.Fatalf("instances = %+v, want the one instance carrying its granted member", got)
	}
}
