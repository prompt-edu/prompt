package resourceconfig

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func newResourceConfigTestRouter(svc *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router.Group("/api/course_phase/:coursePhaseID"), svc)
	return router
}

func TestResourceConfigRoutes(t *testing.T) {
	testDB, cleanup := setupResourceConfigTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	createProviderForResourceConfigTest(t, testDB.Queries, coursePhaseID)
	router := newResourceConfigTestRouter(NewService(testDB.Conn))

	body := []byte(`{"providerType":"gitlab","resourceType":"group","scope":"per_team","nameTemplate":"{{teamName}}","permissionMapping":{"student":"developer"},"resourceExtraConfig":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/course_phase/"+coursePhaseID.String()+"/resource-configs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("POST status = %d, want %d: %s", resp.Code, http.StatusCreated, resp.Body.String())
	}

	var created ResourceConfigResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created config response: %v", err)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/course_phase/"+coursePhaseID.String()+"/resource-configs/"+created.ID.String(), nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d: %s", resp.Code, http.StatusOK, resp.Body.String())
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/course_phase/"+coursePhaseID.String()+"/resource-configs/"+created.ID.String(), nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, want %d: %s", resp.Code, http.StatusNoContent, resp.Body.String())
	}
}

func TestResourceConfigRouteRejectsInvalidIDs(t *testing.T) {
	testDB, cleanup := setupResourceConfigTestDB(t)
	defer cleanup()

	router := newResourceConfigTestRouter(NewService(testDB.Conn))
	req := httptest.NewRequest(http.MethodGet, "/api/course_phase/not-a-uuid/resource-configs", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("course phase status = %d, want %d", resp.Code, http.StatusBadRequest)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/course_phase/"+uuid.New().String()+"/resource-configs/not-a-uuid", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("resource config status = %d, want %d", resp.Code, http.StatusBadRequest)
	}
}
