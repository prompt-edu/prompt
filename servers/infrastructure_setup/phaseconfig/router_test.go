package phaseconfig

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/phaseconfig/phaseconfigDTO"
)

func newPhaseConfigTestRouter(svc *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router.Group("/api/course_phase/:coursePhaseID"), svc)
	return router
}

func TestSetupConfigRoutes(t *testing.T) {
	testDB, cleanup := setupPhaseConfigTestDB(t)
	defer cleanup()

	router := newPhaseConfigTestRouter(NewService(testDB.Conn))
	coursePhaseID := uuid.New()

	body := []byte(`{"semesterTag":"ios26"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/course_phase/"+coursePhaseID.String()+"/setup-config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want %d: %s", resp.Code, http.StatusOK, resp.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/course_phase/"+coursePhaseID.String()+"/setup-config", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d: %s", resp.Code, http.StatusOK, resp.Body.String())
	}
	var got phaseconfigDTO.Response
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode setup config response: %v", err)
	}
	if got.SemesterTag != "ios26" {
		t.Fatalf("semester tag = %q, want ios26", got.SemesterTag)
	}
}

func TestSetupConfigRouteRejectsInvalidCoursePhaseID(t *testing.T) {
	testDB, cleanup := setupPhaseConfigTestDB(t)
	defer cleanup()

	router := newPhaseConfigTestRouter(NewService(testDB.Conn))
	req := httptest.NewRequest(http.MethodGet, "/api/course_phase/not-a-uuid/setup-config", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusBadRequest)
	}
}
