package auditLog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func allowAnyCoursePermission(*gin.Context, uuid.UUID, ...string) (bool, error) {
	return true, nil
}

func TestRegisterRoutes_KeepsStatusMountedWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("AUDIT_ENABLED", "")

	router := gin.New()
	RegisterRoutes(router.Group("/api"), NewAuditLogService(db.Queries{}), func() gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddleware(nil)
	}, allowAnyCoursePermission)

	mounted := make(map[string]bool)
	for _, route := range router.Routes() {
		mounted[route.Path] = true
	}

	assert.True(t, mounted["/api/audit-log/status"])
	assert.False(t, mounted["/api/audit-log"])
	assert.False(t, mounted["/api/courses/:uuid/audit-log"])
}
