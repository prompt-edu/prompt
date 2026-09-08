package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/example_server/config"
	"github.com/prompt-edu/prompt/servers/example_server/copy"
	db "github.com/prompt-edu/prompt/servers/example_server/db/sqlc"
	"github.com/prompt-edu/prompt/servers/example_server/example"
	"github.com/stretchr/testify/require"
)

const (
	auditActorID             = "44444444-4444-4444-4444-444444444444"
	auditCoursePhaseID       = "10000000-0000-0000-0000-000000000001"
	auditSourceCoursePhaseID = "10000000-0000-0000-0000-000000000002"
	auditProbePath           = "/audit-probe"
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

func passThroughAuthMiddleware(_ ...string) gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// auditRouter rebuilds main.go's group layout, with the audit middleware
// attached to the api group before the course phase subgroup exists, because
// gin snapshots the handler chain when a subgroup is created. The actor
// middleware stands in for a token that authenticates but lacks the role the
// route requires, which is the denial the audit middleware is meant to capture.
func auditRouter(sink audit.Sink, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := router.Group("example-service/api")
	api.Use(audit.Middleware(sink))
	api.Use(auditActorMiddleware())
	coursePhaseApi := api.Group("/course_phase/:coursePhaseID")

	queries := db.New(nil)
	var conn *pgxpool.Pool
	configService := config.NewConfigService(*queries, conn)
	copyService := copy.NewCopyService(*queries, conn)
	exampleService := example.NewExampleService(*queries, conn)

	config.RegisterRoutes(coursePhaseApi, configService, authMiddleware)
	example.RegisterRoutes(coursePhaseApi, exampleService, authMiddleware)
	copy.RegisterRoutes(api, copyService, authMiddleware)

	// The service has no mutating course phase route yet, so this stands in for one.
	coursePhaseApi.POST(auditProbePath, authMiddleware(promptSDK.PromptAdmin), func(c *gin.Context) {})

	return router
}

func TestAuditMiddlewareRecordsDeniedCopy(t *testing.T) {
	sink := &recordingSink{}
	router := auditRouter(sink, promptSDK.AuthenticationMiddleware)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest(http.MethodPost, "/example-service/api/copy", nil))
	require.Equal(t, http.StatusUnauthorized, resp.Code)

	events := sink.waitForEvents(1)
	require.Len(t, events, 1)
	require.Equal(t, "Copied course phase", events[0].Action)
	require.Equal(t, "POST /example-service/api/copy", events[0].ActionKey)
	require.Equal(t, audit.OutcomeDenied, events[0].Outcome)
	require.Equal(t, http.StatusUnauthorized, events[0].HTTPStatus)
	require.Equal(t, auditActorID, events[0].ActorID)
	require.Equal(t, "Ada Lovelace", events[0].ActorName)
}

// The probe route lives on a subgroup created after api.Use, so it only records
// while the audit middleware is registered before the subgroup exists.
func TestAuditMiddlewareRecordsCoursePhaseSubgroup(t *testing.T) {
	sink := &recordingSink{}
	router := auditRouter(sink, promptSDK.AuthenticationMiddleware)

	resp := httptest.NewRecorder()
	url := "/example-service/api/course_phase/" + auditCoursePhaseID + auditProbePath
	router.ServeHTTP(resp, httptest.NewRequest(http.MethodPost, url, nil))
	require.Equal(t, http.StatusUnauthorized, resp.Code)

	events := sink.waitForEvents(1)
	require.Len(t, events, 1)
	require.Equal(t, "POST /example-service/api/course_phase/:coursePhaseID"+auditProbePath, events[0].ActionKey)
	require.Equal(t, audit.OutcomeDenied, events[0].Outcome)
	require.Equal(t, auditCoursePhaseID, events[0].CoursePhaseID)
}

func TestAuditMiddlewareIgnoresReads(t *testing.T) {
	sink := &recordingSink{}
	router := auditRouter(sink, promptSDK.AuthenticationMiddleware)

	resp := httptest.NewRecorder()
	url := "/example-service/api/course_phase/" + auditCoursePhaseID + "/config"
	router.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, url, nil))
	require.Equal(t, http.StatusUnauthorized, resp.Code)

	time.Sleep(300 * time.Millisecond)
	require.Empty(t, sink.snapshot())
}

func TestHandlePhaseCopyRecordsExplicitEvent(t *testing.T) {
	sink := &recordingSink{}
	router := auditRouter(sink, passThroughAuthMiddleware)

	body, err := json.Marshal(promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: uuid.MustParse(auditSourceCoursePhaseID),
		TargetCoursePhaseID: uuid.MustParse(auditCoursePhaseID),
	})
	require.NoError(t, err)

	resp := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/example-service/api/copy", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(resp, request)
	require.Equal(t, http.StatusNotFound, resp.Code)

	events := sink.waitForEvents(1)
	require.Len(t, events, 1)
	require.Equal(t, "Copied course phase", events[0].Action)
	require.Equal(t, "coursePhase", events[0].EntityType)
	require.Equal(t, auditCoursePhaseID, events[0].EntityID)
	require.Equal(t, auditCoursePhaseID, events[0].CoursePhaseID)
	require.Equal(t, auditSourceCoursePhaseID, events[0].Metadata["sourceCoursePhaseID"])
	require.Equal(t, auditActorID, events[0].ActorID)
	require.Equal(t, audit.OutcomeError, events[0].Outcome)
	require.Equal(t, http.StatusNotFound, events[0].HTTPStatus)

	time.Sleep(300 * time.Millisecond)
	require.Len(t, sink.snapshot(), 1)
}
