package auditLog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseIngestKeys_SkipsEmptyAndSupportsRotation(t *testing.T) {
	keys := parseIngestKeys("interview:old, interview:new, assessment:sekret, broken, empty:")

	// A service listed twice keeps both keys so a key can be rotated without
	// downtime.
	require.Equal(t, []string{"old", "new"}, keys["interview"])
	require.Equal(t, []string{"sekret"}, keys["assessment"])
	// An empty key is dropped: subtle.ConstantTimeCompare("", "") returns 1, so a
	// blank configured key would otherwise authenticate every request.
	assert.NotContains(t, keys, "empty")
	// A pair without a ':' is ignored.
	assert.NotContains(t, keys, "broken")
}

func TestMatchesAnyKey(t *testing.T) {
	accepted := []string{"old", "new"}
	assert.True(t, matchesAnyKey("old", accepted))
	assert.True(t, matchesAnyKey("new", accepted))
	assert.False(t, matchesAnyKey("wrong", accepted))
	// No configured key (or an empty accepted list) can never authenticate, in
	// particular not with an empty token.
	assert.False(t, matchesAnyKey("", nil))
	assert.False(t, matchesAnyKey("", []string{}))
}

func TestGetAuditLogStatus_ReflectsFeatureToggle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, tt := range []struct {
		toggle string
		want   string
	}{
		{toggle: "true", want: `{"enabled":true}`},
		{toggle: "", want: `{"enabled":false}`},
	} {
		t.Setenv("AUDIT_ENABLED", tt.toggle)

		router := gin.New()
		router.GET("/audit-log/status", getAuditLogStatus)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/audit-log/status", nil))

		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, tt.want, w.Body.String())
	}
}

func TestRegisterRoutes_KeepsStatusMountedWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("AUDIT_ENABLED", "")

	router := gin.New()
	RegisterRoutes(router.Group("/api"), NewAuditLogService(db.Queries{}))

	mounted := make(map[string]bool)
	for _, route := range router.Routes() {
		mounted[route.Path] = true
	}

	assert.True(t, mounted["/api/audit-log/status"])
	assert.False(t, mounted["/api/audit-log"])
	assert.False(t, mounted["/api/courses/:uuid/audit-log"])
}
