package providerconfig

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func newProviderConfigTestRouter(svc *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router.Group("/api/course_phase/:coursePhaseID"), svc)
	return router
}

func TestProviderConfigRoutes(t *testing.T) {
	setProviderConfigEncryptionKey(t)
	testDB, cleanup := setupProviderConfigTestDB(t)
	defer cleanup()

	router := newProviderConfigTestRouter(NewService(testDB.Conn))
	coursePhaseID := uuid.New()

	body := []byte(`{"providerType":"slack","credentials":{"bot_token":"secret"}}`)
	req := httptest.NewRequest(http.MethodPut, "/api/course_phase/"+coursePhaseID.String()+"/provider-configs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want %d: %s", resp.Code, http.StatusOK, resp.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/course_phase/"+coursePhaseID.String()+"/provider-configs", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d: %s", resp.Code, http.StatusOK, resp.Body.String())
	}
	var got []ProviderConfigResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode provider configs response: %v", err)
	}
	if len(got) != 1 || got[0].ProviderType != "slack" {
		t.Fatalf("provider configs = %+v, want one slack config", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/course_phase/"+coursePhaseID.String()+"/provider-configs/slack/fields", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("fields status = %d, want %d: %s", resp.Code, http.StatusOK, resp.Body.String())
	}
}

func TestProviderConfigRouteRejectsInvalidCoursePhaseID(t *testing.T) {
	testDB, cleanup := setupProviderConfigTestDB(t)
	defer cleanup()

	router := newProviderConfigTestRouter(NewService(testDB.Conn))
	req := httptest.NewRequest(http.MethodGet, "/api/course_phase/not-a-uuid/provider-configs", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusBadRequest)
	}
}
