package tease

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt/servers/team_allocation/bodyLimit"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/stretchr/testify/assert"
)

// TestTeaseWriteRoutesRaiseGlobalBodyLimit verifies the workspace routes lift the global body
// cap to their own 5 MiB limit, so a workspace above the 2 MiB default still reaches the
// handler while bodies above 5 MiB are still rejected.
func TestTeaseWriteRoutesRaiseGlobalBodyLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(bodyLimit.Middleware(bodyLimit.DefaultMaxBytes))
	allowAll := func(...string) gin.HandlerFunc { return func(c *gin.Context) { c.Next() } }
	RegisterRoutes(router.Group(""), NewTeaseService(db.Queries{}, nil), allowAll)

	base := "/tease/course_phase/10000000-0000-0000-0000-000000000001"
	// Unterminated JSON: the handler fails to bind it (400) without touching the database,
	// which only happens if the body was read past the 2 MiB default.
	aboveDefault := `{"padding":"` + strings.Repeat("x", int(bodyLimit.DefaultMaxBytes))
	aboveTeaseLimit := `{"padding":"` + strings.Repeat("x", maxTeaseWorkspaceBodyBytes)

	for _, route := range []struct{ method, path string }{
		{http.MethodPut, base + "/workspace"},
		{http.MethodPost, base + "/save"},
	} {
		req := httptest.NewRequest(route.method, route.path, strings.NewReader(aboveDefault))
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code, route.path)

		req = httptest.NewRequest(route.method, route.path, strings.NewReader(aboveTeaseLimit))
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusRequestEntityTooLarge, resp.Code, route.path)
	}
}
