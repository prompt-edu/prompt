package execution

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
