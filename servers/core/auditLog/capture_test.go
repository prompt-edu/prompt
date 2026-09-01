package auditLog

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt/servers/core/auditLog/auditLogDTO"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/stretchr/testify/require"
)

const seededActorID = "44444444-4444-4444-4444-444444444444"

// fakeCoreAuth populates the flat context keys core's own Keycloak middleware
// sets, which is what CoreActorExtractor reads.
func fakeCoreAuth(roles map[string]bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(keycloakTokenVerifier.CtxUserID, seededActorID)
		c.Set(keycloakTokenVerifier.CtxFirstName, "Ada")
		c.Set(keycloakTokenVerifier.CtxLastName, "Lovelace")
		c.Set(keycloakTokenVerifier.CtxUserEmail, "ada@tum.de")
		c.Set(keycloakTokenVerifier.CtxUserRoles, roles)
		c.Next()
	}
}

func denyingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied"})
	}
}

// captureRouter mirrors main.go's wiring: the capture middleware goes onto the
// /api group before any module registers its routes.
func (s *AuditLogTestSuite) captureRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	RegisterCapture(api, s.service)

	ok := func(c *gin.Context) { c.Status(http.StatusOK) }
	api.PUT("/courses/:uuid/archive", fakeCoreAuth(map[string]bool{"PROMPT_Admin": true}), ok)
	api.DELETE("/students/:uuid", fakeCoreAuth(map[string]bool{"PROMPT_Admin": true}), ok)
	api.PUT("/courses/:uuid/denied", fakeCoreAuth(map[string]bool{"ios2425-x-Student": true}), denyingMiddleware(), ok)
	api.GET("/courses/:uuid", fakeCoreAuth(map[string]bool{"PROMPT_Admin": true}), ok)
	api.PUT("/courses/:uuid/anonymous", ok)

	// Gin snapshots the middleware chain when a route or subgroup is registered,
	// so a subgroup created after the api.Use() call must still be captured.
	phases := api.Group("/course_phases")
	phases.POST("/course/:courseID", fakeCoreAuth(map[string]bool{"PROMPT_Lecturer": true}), func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	return router
}

func serve(router *gin.Engine, method, path string) int {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(method, path, nil))
	return w.Code
}

// waitForEntries polls until at least n entries are stored: the capture
// middleware delivers to the sink from a background goroutine.
func (s *AuditLogTestSuite) waitForEntries(n int) []auditLogDTO.AuditEntry {
	deadline := time.Now().Add(5 * time.Second)
	for {
		page, err := s.service.ListAuditLog(s.ctx, auditLogDTO.ListFilters{})
		require.NoError(s.T(), err)
		if len(page.Entries) >= n || time.Now().After(deadline) {
			return page.Entries
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func entryByActionKey(entries []auditLogDTO.AuditEntry, actionKey string) *auditLogDTO.AuditEntry {
	for i := range entries {
		if entries[i].ActionKey == actionKey {
			return &entries[i]
		}
	}
	return nil
}

func (s *AuditLogTestSuite) TestCaptureMiddleware_RecordsMutationsIncludingSubgroups() {
	router := s.captureRouter()

	require.Equal(s.T(), http.StatusOK, serve(router, "PUT", "/api/courses/"+seededCourseID+"/archive"))
	require.Equal(s.T(), http.StatusCreated, serve(router, "POST", "/api/course_phases/course/"+seededCourseID))

	entries := s.waitForEntries(2)
	require.Len(s.T(), entries, 2)

	archive := entryByActionKey(entries, "PUT /api/courses/:uuid/archive")
	require.NotNil(s.T(), archive)
	require.Equal(s.T(), "Updated archive", archive.Action)
	require.Equal(s.T(), "success", archive.Outcome)
	require.Equal(s.T(), "PROMPT_Admin", archive.ActorRole)
	require.Equal(s.T(), seededActorID, archive.ActorID)
	require.Equal(s.T(), "Ada Lovelace", archive.ActorName)
	require.Equal(s.T(), seededCourseID, archive.EntityID)
	require.Equal(s.T(), seededCourseID, archive.CourseID)
	require.Equal(s.T(), "core", archive.SourceService)
	require.Equal(s.T(), 200, archive.HTTPStatus)

	// Registered on a subgroup created after the middleware was attached.
	phase := entryByActionKey(entries, "POST /api/course_phases/course/:courseID")
	require.NotNil(s.T(), phase)
	require.Equal(s.T(), seededCourseID, phase.CourseID)
	require.Equal(s.T(), "PROMPT_Lecturer", phase.ActorRole)
}

func (s *AuditLogTestSuite) TestCaptureMiddleware_RecordsDeniedAttempt() {
	router := s.captureRouter()

	require.Equal(s.T(), http.StatusForbidden, serve(router, "PUT", "/api/courses/"+seededCourseID+"/denied"))

	entries := s.waitForEntries(1)
	require.Len(s.T(), entries, 1)
	require.Equal(s.T(), "denied", entries[0].Outcome)
	require.Equal(s.T(), 403, entries[0].HTTPStatus)
	require.Equal(s.T(), "Student", entries[0].ActorRole)
}

func (s *AuditLogTestSuite) TestCaptureMiddleware_IgnoresReadsAndUnauthenticatedRequests() {
	router := s.captureRouter()

	require.Equal(s.T(), http.StatusOK, serve(router, "GET", "/api/courses/"+seededCourseID))
	require.Equal(s.T(), http.StatusOK, serve(router, "PUT", "/api/courses/"+seededCourseID+"/anonymous"))

	// Give a stray delivery a chance to land before asserting nothing was stored.
	time.Sleep(300 * time.Millisecond)
	require.Equal(s.T(), 0, s.countRows())
}
