package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/certificate/config"
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
	"github.com/prompt-edu/prompt/servers/certificate/generator"
	"github.com/prompt-edu/prompt/servers/certificate/participants"
	"github.com/stretchr/testify/require"
)

const (
	auditActorID       = "44444444-4444-4444-4444-444444444444"
	auditCoursePhaseID = "10000000-0000-0000-0000-000000000001"
)

type recordingSink struct {
	mutex  sync.Mutex
	events []audit.Event
}

func (s *recordingSink) Record(_ context.Context, e audit.Event) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.events = append(s.events, e)
	return nil
}

func (s *recordingSink) snapshot() []audit.Event {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return append([]audit.Event(nil), s.events...)
}

// waitForEvents polls until at least n events arrived: the audit middleware
// delivers to the sink from a background goroutine.
func (s *recordingSink) waitForEvents(n int) []audit.Event {
	deadline := time.Now().Add(5 * time.Second)
	for {
		events := s.snapshot()
		if len(events) >= n || time.Now().After(deadline) {
			return events
		}
		time.Sleep(25 * time.Millisecond)
	}
}

// auditActorMiddleware populates the token user the default actor extractor
// reads. The SDK's MockAuthMiddleware is unusable here: it sets an empty ID,
// which the extractor rejects.
func auditActorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		keycloakTokenVerifier.SetTokenUser(c, keycloakTokenVerifier.TokenUser{
			ID:        auditActorID,
			FirstName: "Ada",
			LastName:  "Lovelace",
			Email:     "ada@tum.de",
			Roles:     map[string]bool{promptSDK.PromptAdmin: true},
		})
		c.Next()
	}
}

// auditRouter rebuilds main.go's group layout, with the audit middleware
// attached to the api group before the course phase subgroup exists, because
// gin snapshots the handler chain when a subgroup is created. The actor
// middleware stands in for a token that authenticates but lacks the role the
// route requires, which is the denial the audit middleware is meant to capture.
func auditRouter(sink audit.Sink) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := router.Group("certificate/api")
	api.Use(audit.Middleware(sink))
	api.Use(auditActorMiddleware())
	coursePhaseApi := api.Group("/course_phase/:coursePhaseID")

	queries := db.New(nil)
	configService := config.NewConfigService(*queries)
	participantsService := participants.NewParticipantsService(*queries, "http://localhost:8080")
	generatorService := generator.NewGeneratorService(*queries, configService, participantsService, participantsService)

	config.RegisterRoutes(coursePhaseApi, configService, promptSDK.AuthenticationMiddleware)
	participants.RegisterRoutes(coursePhaseApi, participantsService, promptSDK.AuthenticationMiddleware)
	generator.RegisterRoutes(coursePhaseApi, generatorService, promptSDK.AuthenticationMiddleware)

	return router
}

func TestAuditMiddlewareRecordsMutatingConfigRoutes(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		action    string
		actionKey string
	}{
		{
			name:      "config update",
			path:      "/config",
			action:    "Updated the certificate configuration",
			actionKey: "PUT /certificate/api/course_phase/:coursePhaseID/config",
		},
		{
			name:      "release date update",
			path:      "/config/release-date",
			action:    "Updated the certificate release date",
			actionKey: "PUT /certificate/api/course_phase/:coursePhaseID/config/release-date",
		},
		{
			name:      "student page text update",
			path:      "/config/student-page-text",
			action:    "Updated the certificate student page text",
			actionKey: "PUT /certificate/api/course_phase/:coursePhaseID/config/student-page-text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sink := &recordingSink{}
			router := auditRouter(sink)

			resp := httptest.NewRecorder()
			url := "/certificate/api/course_phase/" + auditCoursePhaseID + tt.path
			router.ServeHTTP(resp, httptest.NewRequest(http.MethodPut, url, nil))
			require.Equal(t, http.StatusUnauthorized, resp.Code)

			events := sink.waitForEvents(1)
			require.Len(t, events, 1)
			require.Equal(t, tt.action, events[0].Action)
			require.Equal(t, tt.actionKey, events[0].ActionKey)
			require.Equal(t, audit.OutcomeDenied, events[0].Outcome)
			require.Equal(t, http.StatusUnauthorized, events[0].HTTPStatus)
			require.Equal(t, auditCoursePhaseID, events[0].CoursePhaseID)
			require.Equal(t, auditActorID, events[0].ActorID)
			require.Equal(t, "Ada Lovelace", events[0].ActorName)
		})
	}
}

func TestAuditMiddlewareRecordsOneEventPerRequest(t *testing.T) {
	sink := &recordingSink{}
	router := auditRouter(sink)

	resp := httptest.NewRecorder()
	url := "/certificate/api/course_phase/" + auditCoursePhaseID + "/config"
	router.ServeHTTP(resp, httptest.NewRequest(http.MethodPut, url, nil))
	require.Equal(t, http.StatusUnauthorized, resp.Code)

	require.Len(t, sink.waitForEvents(1), 1)
	time.Sleep(300 * time.Millisecond)
	require.Len(t, sink.snapshot(), 1)
}

func TestAuditMiddlewareIgnoresReads(t *testing.T) {
	sink := &recordingSink{}
	router := auditRouter(sink)

	resp := httptest.NewRecorder()
	url := "/certificate/api/course_phase/" + auditCoursePhaseID + "/config"
	router.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, url, nil))
	require.Equal(t, http.StatusUnauthorized, resp.Code)

	time.Sleep(300 * time.Millisecond)
	require.Empty(t, sink.snapshot())
}
