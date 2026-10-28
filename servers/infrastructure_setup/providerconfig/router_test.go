package providerconfig

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/providerconfig/providerconfigDTO"
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
	var got []providerconfigDTO.ProviderConfigResponse
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

// The validate endpoint must not echo transport errors: they name the dialled address
// and distinguish a refused port from a timeout, which made the endpoint a probe for the
// internal network. Regression test for GHSA-g8fq-mjgm-x63v.
func TestValidationFailureMessageHidesNetworkDetails(t *testing.T) {
	refused := &url.Error{Op: "Get", URL: "https://10.0.0.5:5432/api/v4/user", Err: &net.OpError{
		Op: "dial", Net: "tcp", Addr: &net.TCPAddr{IP: net.ParseIP("10.0.0.5"), Port: 5432}, Err: errors.New("connect: connection refused"),
	}}
	if got := validationFailureMessage(fmt.Errorf("gitlab: %w", refused)); strings.Contains(got, "10.0.0.5") || strings.Contains(got, "refused") {
		t.Fatalf("message = %q, want no address or failure kind", got)
	}

	blocked := &url.Error{Op: "Get", URL: "https://metadata.internal/", Err: provider.ErrBlockedAddress}
	if got := validationFailureMessage(blocked); got != provider.BlockedAddressMessage {
		t.Fatalf("message = %q, want %q", got, provider.BlockedAddressMessage)
	}

	// An upstream status carries no network detail and tells the lecturer what to fix.
	status := provider.HTTPError("gitlab", http.MethodGet, "/api/v4/user", http.StatusUnauthorized, nil)
	if got := validationFailureMessage(status); got != status.Error() {
		t.Fatalf("message = %q, want the status error %q", got, status.Error())
	}
}
