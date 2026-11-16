package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/bodyLimit"
	"github.com/prompt-edu/prompt/servers/assessment/copy"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/stretchr/testify/require"
)

const (
	auditActorID              = "44444444-4444-4444-4444-444444444444"
	auditCoursePhaseID        = "10000000-0000-0000-0000-000000000001"
	auditSourceCoursePhaseID  = "10000000-0000-0000-0000-000000000002"
	auditCourseParticipation  = "10000000-0000-0000-0000-000000000003"
	auditAssessmentID         = "10000000-0000-0000-0000-000000000004"
	auditBase                 = "/assessment/api"
	auditCoursePhaseRoute     = auditBase + "/course_phase/" + auditCoursePhaseID
	auditCoursePhaseTemplate  = auditBase + "/course_phase/:coursePhaseID"
	auditCopyRoute            = auditBase + "/copy"
	auditParticipationSegment = "/course-participation/" + auditCourseParticipation
	auditParticipationParam   = "/course-participation/:courseParticipationID"
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

// auditActorMiddleware populates the token user the default actor extractor reads. The
// SDK's MockAuthMiddleware is unusable here: it sets an empty ID, which the extractor
// rejects.
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

// newServiceRouter builds the router groups the way promptSDK.Bootstrap does, with the test sink
// in place of core's, and registers main.go's routes on them. The actor middleware stands in for a
// token that authenticates but lacks the role the route requires, which is the denial the audit
// middleware is meant to capture.
func newServiceRouter(t *testing.T, sink audit.Sink) (*gin.Engine, *gin.RouterGroup, *gin.RouterGroup) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(audit.Middleware(sink), auditActorMiddleware())
	api := router.Group("/assessment/api")
	coursePhaseApi := api.Group("/course_phase/:coursePhaseID")
	require.NoError(t, registerRoutes(api, coursePhaseApi, nil))
	return router, api, coursePhaseApi
}

type auditRouteCase struct {
	name      string
	method    string
	path      string
	action    string
	actionKey string
}

func requireDeniedEvent(t *testing.T, tc auditRouteCase) {
	t.Helper()

	sink := &recordingSink{}
	router, _, _ := newServiceRouter(t, sink)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest(tc.method, tc.path, nil))
	require.Equal(t, http.StatusUnauthorized, resp.Code)

	events := sink.waitForEvents(1)
	require.Len(t, events, 1)
	require.Equal(t, tc.action, events[0].Action)
	require.Equal(t, tc.actionKey, events[0].ActionKey)
	require.Equal(t, audit.OutcomeDenied, events[0].Outcome)
	require.Equal(t, http.StatusUnauthorized, events[0].HTTPStatus)
	require.Equal(t, auditCoursePhaseID, events[0].CoursePhaseID)
	require.Equal(t, auditActorID, events[0].ActorID)
	require.Equal(t, "Ada Lovelace", events[0].ActorName)
}

func TestAuditMiddlewareUsesDescribedLabels(t *testing.T) {
	tests := []auditRouteCase{
		{
			name:      "assessment completion created",
			method:    http.MethodPost,
			path:      auditCoursePhaseRoute + "/student-assessment/completed",
			action:    "Saved assessment completion",
			actionKey: "POST " + auditCoursePhaseTemplate + "/student-assessment/completed",
		},
		{
			name:      "assessment completion updated",
			method:    http.MethodPut,
			path:      auditCoursePhaseRoute + "/student-assessment/completed",
			action:    "Saved assessment completion",
			actionKey: "PUT " + auditCoursePhaseTemplate + "/student-assessment/completed",
		},
		{
			name:      "assessment marked complete",
			method:    http.MethodPost,
			path:      auditCoursePhaseRoute + "/student-assessment/completed/mark-complete",
			action:    "Marked assessment as completed",
			actionKey: "POST " + auditCoursePhaseTemplate + "/student-assessment/completed/mark-complete",
		},
		{
			name:      "assessments marked complete in a batch",
			method:    http.MethodPost,
			path:      auditCoursePhaseRoute + "/student-assessment/completed/mark-complete/batch",
			action:    "Marked assessments as completed",
			actionKey: "POST " + auditCoursePhaseTemplate + "/student-assessment/completed/mark-complete/batch",
		},
		{
			name:      "assessments unmarked in a batch",
			method:    http.MethodPut,
			path:      auditCoursePhaseRoute + "/student-assessment/completed/unmark/batch",
			action:    "Unmarked assessment completions",
			actionKey: "PUT " + auditCoursePhaseTemplate + "/student-assessment/completed/unmark/batch",
		},
		{
			name:      "assessment unmarked",
			method:    http.MethodPut,
			path:      auditCoursePhaseRoute + "/student-assessment/completed" + auditParticipationSegment + "/unmark",
			action:    "Unmarked assessment completion",
			actionKey: "PUT " + auditCoursePhaseTemplate + "/student-assessment/completed" + auditParticipationParam + "/unmark",
		},
		{
			name:      "assessment completion deleted",
			method:    http.MethodDelete,
			path:      auditCoursePhaseRoute + "/student-assessment/completed" + auditParticipationSegment,
			action:    "Deleted assessment completion",
			actionKey: "DELETE " + auditCoursePhaseTemplate + "/student-assessment/completed" + auditParticipationParam,
		},
		{
			name:      "my evaluation completion created",
			method:    http.MethodPost,
			path:      auditCoursePhaseRoute + "/evaluation/completed/my-completion",
			action:    "Saved own evaluation completion",
			actionKey: "POST " + auditCoursePhaseTemplate + "/evaluation/completed/my-completion",
		},
		{
			name:      "my evaluation completion updated",
			method:    http.MethodPut,
			path:      auditCoursePhaseRoute + "/evaluation/completed/my-completion",
			action:    "Saved own evaluation completion",
			actionKey: "PUT " + auditCoursePhaseTemplate + "/evaluation/completed/my-completion",
		},
		{
			name:      "my evaluation marked complete",
			method:    http.MethodPost,
			path:      auditCoursePhaseRoute + "/evaluation/completed/my-completion/mark-complete",
			action:    "Marked own evaluation as completed",
			actionKey: "POST " + auditCoursePhaseTemplate + "/evaluation/completed/my-completion/mark-complete",
		},
		{
			name:      "my evaluation unmarked",
			method:    http.MethodPut,
			path:      auditCoursePhaseRoute + "/evaluation/completed/my-completion/unmark",
			action:    "Unmarked own evaluation completion",
			actionKey: "PUT " + auditCoursePhaseTemplate + "/evaluation/completed/my-completion/unmark",
		},
		{
			name:      "configuration update",
			method:    http.MethodPut,
			path:      auditCoursePhaseRoute + "/config",
			action:    "Updated the assessment configuration",
			actionKey: "PUT " + auditCoursePhaseTemplate + "/config",
		},
		{
			name:      "results release",
			method:    http.MethodPost,
			path:      auditCoursePhaseRoute + "/config/release",
			action:    "Released assessment results",
			actionKey: "POST " + auditCoursePhaseTemplate + "/config/release",
		},
		{
			name:      "results unrelease",
			method:    http.MethodPost,
			path:      auditCoursePhaseRoute + "/config/unrelease",
			action:    "Withdrew released assessment results",
			actionKey: "POST " + auditCoursePhaseTemplate + "/config/unrelease",
		},
		{
			name:      "reminder dispatch",
			method:    http.MethodPost,
			path:      auditCoursePhaseRoute + "/config/reminders/send",
			action:    "Sent evaluation reminders",
			actionKey: "POST " + auditCoursePhaseTemplate + "/config/reminders/send",
		},
		{
			name:      "category assessment upsert",
			method:    http.MethodPost,
			path:      auditCoursePhaseRoute + "/category-assessment",
			action:    "Saved category assessment",
			actionKey: "POST " + auditCoursePhaseTemplate + "/category-assessment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requireDeniedEvent(t, tt)
		})
	}
}

// These routes are deliberately left on the SDK's derived label, so a change to the
// derivation must not silently rewrite them.
func TestAuditMiddlewareUsesDerivedLabels(t *testing.T) {
	tests := []auditRouteCase{
		{
			name:      "student assessment deletion",
			method:    http.MethodDelete,
			path:      auditCoursePhaseRoute + "/student-assessment/" + auditAssessmentID,
			action:    "Deleted student assessment",
			actionKey: "DELETE " + auditCoursePhaseTemplate + "/student-assessment/:assessmentID",
		},
		{
			name:      "action item creation",
			method:    http.MethodPost,
			path:      auditCoursePhaseRoute + "/student-assessment/action-item",
			action:    "Created action item",
			actionKey: "POST " + auditCoursePhaseTemplate + "/student-assessment/action-item",
		},
		{
			name:      "assessment schema creation",
			method:    http.MethodPost,
			path:      auditCoursePhaseRoute + "/assessment-schema",
			action:    "Created assessment schema",
			actionKey: "POST " + auditCoursePhaseTemplate + "/assessment-schema",
		},
		{
			name:      "feedback item creation",
			method:    http.MethodPost,
			path:      auditCoursePhaseRoute + "/evaluation/feedback-items",
			action:    "Created feedback item",
			actionKey: "POST " + auditCoursePhaseTemplate + "/evaluation/feedback-items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requireDeniedEvent(t, tt)
		})
	}
}

func TestAuditMiddlewareIgnoresReads(t *testing.T) {
	sink := &recordingSink{}
	router, _, _ := newServiceRouter(t, sink)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, auditCoursePhaseRoute+"/config", nil))
	require.Equal(t, http.StatusUnauthorized, resp.Code)

	time.Sleep(300 * time.Millisecond)
	require.Empty(t, sink.snapshot())
}

func postCopyRequest(t *testing.T, router *gin.Engine, source, target uuid.UUID) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(promptTypes.PhaseCopyRequest{SourceCoursePhaseID: source, TargetCoursePhaseID: target})
	require.NoError(t, err)

	resp := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, auditCopyRoute, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(resp, request)
	return resp
}

// auditCopyRouter reaches the copy handler itself, which the denied paths never
// do: copy.RegisterRoutes wires the real SDK auth middleware, so a request
// without a bearer token stops before the handler. The group layout and the
// audit label mirror what main.go and copy.RegisterRoutes build.
func auditCopyRouter(sink audit.Sink, copyService *copy.CopyService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := router.Group("/assessment/api")
	api.Use(audit.Middleware(sink))
	api.Use(auditActorMiddleware())
	promptTypes.RegisterCopyEndpoint(
		api.Group("", audit.Describe(copy.AuditCopyAction)),
		passThroughAuthMiddleware(),
		copyService,
	)

	return router
}

// Core detects copy support by posting a copy of a phase onto itself, forwarding the
// caller's token, so that probe must leave no trace in the audit log.
func TestHandlePhaseCopyProbeRecordsNothing(t *testing.T) {
	sink := &recordingSink{}
	router := auditCopyRouter(sink, copy.NewCopyService(*db.New(nil), nil))

	phaseID := uuid.MustParse(auditCoursePhaseID)
	require.Equal(t, http.StatusOK, postCopyRequest(t, router, phaseID, phaseID).Code)

	time.Sleep(300 * time.Millisecond)
	require.Empty(t, sink.snapshot())
}

func TestHandlePhaseCopyRecordsScopedEvent(t *testing.T) {
	ctx := context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(ctx, "db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "database_dumps/coursePhaseConfig.sql")
	require.NoError(t, err)
	defer cleanup()

	source := uuid.MustParse(auditSourceCoursePhaseID)
	target := uuid.MustParse(auditCoursePhaseID)
	_, err = testDB.Conn.Exec(ctx, `INSERT INTO course_phase_config (course_phase_id) VALUES ($1)`, source)
	require.NoError(t, err)

	sink := &recordingSink{}
	router := auditCopyRouter(sink, copy.NewCopyService(*testDB.Queries, testDB.Conn))

	require.Equal(t, http.StatusOK, postCopyRequest(t, router, source, target).Code)

	events := sink.waitForEvents(1)
	require.Len(t, events, 1)
	require.Equal(t, "Copied assessment phase", events[0].Action)
	require.Equal(t, audit.OutcomeSuccess, events[0].Outcome)
	require.Equal(t, "coursePhase", events[0].EntityType)
	require.Equal(t, auditCoursePhaseID, events[0].EntityID)
	require.Equal(t, auditCoursePhaseID, events[0].CoursePhaseID)
	require.Equal(t, auditSourceCoursePhaseID, events[0].Metadata["sourceCoursePhaseID"])
	require.Equal(t, auditActorID, events[0].ActorID)
}

// The grading and evaluation forms post on every interaction, so these two routes
// are silenced: auditing them would bury the log and start dropping events.
func TestAuditMiddlewareSkipsTheAutosaveRoutes(t *testing.T) {
	for _, path := range []string{"/student-assessment", "/evaluation"} {
		t.Run(path, func(t *testing.T) {
			sink := &recordingSink{}
			router, _, _ := newServiceRouter(t, sink)

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, httptest.NewRequest(http.MethodPost, auditCoursePhaseRoute+path, nil))

			time.Sleep(300 * time.Millisecond)
			require.Empty(t, sink.snapshot())
		})
	}
}

// Bootstrap creates the course phase group before registerRoutes runs, so the body cap has to be
// registered on both groups to reach every route.
func TestBodyCapCoversBothGroups(t *testing.T) {
	router, api, coursePhaseApi := newServiceRouter(t, &recordingSink{})
	readBody := func(c *gin.Context) {
		if _, err := io.ReadAll(c.Request.Body); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
		}
	}
	api.POST("/body-cap-probe", readBody)
	coursePhaseApi.POST("/body-cap-probe", readBody)

	for _, path := range []string{"/assessment/api/body-cap-probe", "/assessment/api/course_phase/" + auditCoursePhaseID + "/body-cap-probe"} {
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, httptest.NewRequest(http.MethodPost, path, bytes.NewReader(make([]byte, bodyLimit.DefaultMaxBytes+1))))
		require.Equal(t, http.StatusRequestEntityTooLarge, resp.Code, path)
	}
}
