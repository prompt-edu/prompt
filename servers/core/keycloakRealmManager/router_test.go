package keycloakRealmManager

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// staffRouterCallerID is the Keycloak user every staff route test request is made as.
const staffRouterCallerID = "6f1c2a8e-3b4d-4e5f-9a0b-1c2d3e4f5a6b"

// recordingKeycloak issues service-account tokens, knows no users, and records every request it gets.
type recordingKeycloak struct {
	mu       sync.Mutex
	requests []string // as "<method> <path>"
}

func (k *recordingKeycloak) start(t *testing.T) *KeycloakRealmService {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		k.mu.Lock()
		k.requests = append(k.requests, r.Method+" "+r.URL.Path)
		k.mu.Unlock()
		if r.Method == http.MethodPost && r.URL.Path == "/realms/prompt/protocol/openid-connect/token" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"token","expires_in":60}`))
			return
		}
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
	}))
	t.Cleanup(server.Close)
	return NewKeycloakRealmService(server.URL, "prompt", "prompt-server", "secret", "client-uuid", db.Queries{})
}

func (k *recordingKeycloak) received() []string {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.requests
}

// staffRouter mounts the keycloak routes with every request made as staffRouterCallerID and every
// permission check passing.
func staffRouter(service *KeycloakRealmService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authMiddleware := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set(keycloakTokenVerifier.CtxUserID, staffRouterCallerID)
			c.Next()
		}
	}
	allowAll := func(...string) gin.HandlerFunc {
		return func(c *gin.Context) { c.Next() }
	}
	setupKeycloakRouter(router.Group("/api"), service, authMiddleware, allowAll)
	return router
}

func serveStaffRequest(router *gin.Engine, method, userID string) *httptest.ResponseRecorder {
	path := "/api/keycloak/" + uuid.NewString() + "/group/" + permissionValidation.CourseLecturer + "/members/" + userID
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(method, path, nil))
	return w
}

func TestCourseStaffRoutesRejectUserIDsThatAreNotUUIDs(t *testing.T) {
	valid := uuid.NewString()
	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		for _, userID := range []string{"not-a-uuid", valid + "-1", valid[:len(valid)-1]} {
			t.Run(method+" "+userID, func(t *testing.T) {
				keycloak := &recordingKeycloak{}
				w := serveStaffRequest(staffRouter(keycloak.start(t)), method, userID)

				assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
				assert.Empty(t, keycloak.received(), "the request must not reach Keycloak")
			})
		}
	}
}

func TestAddCourseStaffMemberPassesTheCanonicalUserIDToKeycloak(t *testing.T) {
	keycloak := &recordingKeycloak{}
	userID := uuid.NewString()
	w := serveStaffRequest(staffRouter(keycloak.start(t)), http.MethodPut, strings.ToUpper(userID))

	// The fake Keycloak knows no users, so the lookup of the target user ends the request.
	assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	assert.Contains(t, keycloak.received(), "GET /admin/realms/prompt/users/"+userID)
}

func TestRemoveCourseStaffMemberRefusesAnUppercaseSpellingOfTheCallersOwnID(t *testing.T) {
	keycloak := &recordingKeycloak{}
	w := serveStaffRequest(staffRouter(keycloak.start(t)), http.MethodDelete, strings.ToUpper(staffRouterCallerID))

	require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	var body sdkUtils.ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, ErrSelfRemoval.Error(), body.Error)
	assert.Empty(t, keycloak.received(), "the request must not reach Keycloak")
}
