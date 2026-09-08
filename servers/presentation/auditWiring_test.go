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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/presentation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/presentation/presentation"
	"github.com/stretchr/testify/require"
)

const (
	auditActorID             = "44444444-4444-4444-4444-444444444444"
	auditCoursePhaseID       = "10000000-0000-0000-0000-000000000001"
	auditSourceCoursePhaseID = "10000000-0000-0000-0000-000000000002"
	auditCategoryID          = "10000000-0000-0000-0000-000000000003"
	auditSlotID              = "10000000-0000-0000-0000-000000000004"
	auditPresentationID      = "10000000-0000-0000-0000-000000000005"
	auditMaterialID          = "10000000-0000-0000-0000-000000000006"
	auditCoursePhaseRoute    = "/presentation/api/course_phase/" + auditCoursePhaseID
	auditCoursePhaseTemplate = "/presentation/api/course_phase/:coursePhaseID"
	auditCopyRoute           = "/presentation/api/copy"
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

// waitForEvents polls until at least n events arrived: the audit middleware delivers to the
// sink from a background goroutine.
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

// auditRouter rebuilds main.go's group layout, with the audit middleware attached to the
// api group before the course phase subgroup exists, because gin snapshots the handler
// chain when a subgroup is created. RegisterRoutes wires the real SDK auth middleware, so
// every request without a bearer token is denied, which is the outcome the audit
// middleware is meant to capture. No handler runs, so the service needs no live database.
func auditRouter(sink audit.Sink) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := router.Group("/presentation/api")
	api.Use(audit.Middleware(sink))
	api.Use(auditActorMiddleware())
	coursePhaseAPI := api.Group("/course_phase/:coursePhaseID")

	service := presentation.NewService(db.New(nil), nil, nil, "", 0, 0, 0, nil)
	presentation.RegisterRoutes(coursePhaseAPI, service)
	promptTypes.RegisterCopyEndpoint(
		api.Group("", audit.Describe(presentation.AuditCopyAction)),
		promptSDK.AuthenticationMiddleware(promptSDK.PromptAdmin, promptSDK.CourseLecturer),
		&presentation.CopyHandler{Service: service},
	)
	return router
}

func requireDeniedEvent(t *testing.T, method, url, action, actionKey, coursePhaseID string) {
	t.Helper()

	sink := &recordingSink{}
	router := auditRouter(sink)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(method, url, nil))
	require.Equal(t, http.StatusUnauthorized, response.Code)

	events := sink.waitForEvents(1)
	require.Len(t, events, 1)
	require.Equal(t, action, events[0].Action)
	require.Equal(t, actionKey, events[0].ActionKey)
	require.Equal(t, audit.OutcomeDenied, events[0].Outcome)
	require.Equal(t, http.StatusUnauthorized, events[0].HTTPStatus)
	require.Equal(t, coursePhaseID, events[0].CoursePhaseID)
	require.Equal(t, auditActorID, events[0].ActorID)
	require.Equal(t, "Ada Lovelace", events[0].ActorName)
}

func TestAuditMiddlewareUsesDescribedLabels(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		template string
		action   string
	}{
		{
			name:     "update config",
			method:   http.MethodPut,
			path:     "/config",
			template: "/config",
			action:   "Updated the presentation configuration",
		},
		{
			name:     "create category",
			method:   http.MethodPost,
			path:     "/categories",
			template: "/categories",
			action:   "Created a feedback category",
		},
		{
			name:     "update category",
			method:   http.MethodPut,
			path:     "/categories/" + auditCategoryID,
			template: "/categories/:categoryID",
			action:   "Updated a feedback category",
		},
		{
			name:     "delete category",
			method:   http.MethodDelete,
			path:     "/categories/" + auditCategoryID,
			template: "/categories/:categoryID",
			action:   "Deleted a feedback category",
		},
		{
			name:     "create slot",
			method:   http.MethodPost,
			path:     "/slots",
			template: "/slots",
			action:   "Created a presentation slot",
		},
		{
			name:     "create slots in bulk",
			method:   http.MethodPost,
			path:     "/slots/batch",
			template: "/slots/batch",
			action:   "Created presentation slots in bulk",
		},
		{
			name:     "update slot",
			method:   http.MethodPut,
			path:     "/slots/" + auditSlotID,
			template: "/slots/:slotID",
			action:   "Updated a presentation slot",
		},
		{
			name:     "delete slot",
			method:   http.MethodDelete,
			path:     "/slots/" + auditSlotID,
			template: "/slots/:slotID",
			action:   "Deleted a presentation slot",
		},
		{
			name:     "assign slot",
			method:   http.MethodPut,
			path:     "/slots/" + auditSlotID + "/assignment",
			template: "/slots/:slotID/assignment",
			action:   "Assigned a presentation target to a slot",
		},
		{
			name:     "unassign slot",
			method:   http.MethodDelete,
			path:     "/slots/" + auditSlotID + "/assignment",
			template: "/slots/:slotID/assignment",
			action:   "Unassigned a presentation target from a slot",
		},
		{
			name:     "presign material upload",
			method:   http.MethodPost,
			path:     "/presentations/" + auditPresentationID + "/materials/presign",
			template: "/presentations/:presentationID/materials/presign",
			action:   "Started a presentation material upload",
		},
		{
			name:     "complete material upload",
			method:   http.MethodPost,
			path:     "/presentations/" + auditPresentationID + "/materials/" + auditMaterialID + "/complete",
			template: "/presentations/:presentationID/materials/:materialID/complete",
			action:   "Uploaded presentation material",
		},
		{
			name:     "delete material",
			method:   http.MethodDelete,
			path:     "/presentations/" + auditPresentationID + "/materials/" + auditMaterialID,
			template: "/presentations/:presentationID/materials/:materialID",
			action:   "Deleted presentation material",
		},
		{
			name:     "update feedback answer",
			method:   http.MethodPut,
			path:     "/presentations/" + auditPresentationID + "/feedback/answers/" + auditCategoryID,
			template: "/presentations/:presentationID/feedback/answers/:categoryID",
			action:   "Updated a presentation feedback answer",
		},
		{
			name:     "submit feedback",
			method:   http.MethodPost,
			path:     "/presentations/" + auditPresentationID + "/feedback/submit",
			template: "/presentations/:presentationID/feedback/submit",
			action:   "Submitted presentation feedback",
		},
		{
			name:     "reopen feedback",
			method:   http.MethodPost,
			path:     "/presentations/" + auditPresentationID + "/feedback/reopen",
			template: "/presentations/:presentationID/feedback/reopen",
			action:   "Reopened presentation feedback",
		},
		{
			name:     "discard feedback draft",
			method:   http.MethodDelete,
			path:     "/presentations/" + auditPresentationID + "/feedback/draft",
			template: "/presentations/:presentationID/feedback/draft",
			action:   "Discarded a presentation feedback draft",
		},
		{
			name:     "release feedback",
			method:   http.MethodPost,
			path:     "/presentations/" + auditPresentationID + "/feedback/release",
			template: "/presentations/:presentationID/feedback/release",
			action:   "Released presentation feedback",
		},
		{
			name:     "unrelease feedback",
			method:   http.MethodDelete,
			path:     "/presentations/" + auditPresentationID + "/feedback/release",
			template: "/presentations/:presentationID/feedback/release",
			action:   "Unreleased presentation feedback",
		},
		{
			name:     "reset feedback",
			method:   http.MethodDelete,
			path:     "/presentations/" + auditPresentationID + "/feedback",
			template: "/presentations/:presentationID/feedback",
			action:   "Reset presentation feedback",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requireDeniedEvent(t, test.method,
				auditCoursePhaseRoute+test.path,
				test.action,
				test.method+" "+auditCoursePhaseTemplate+test.template,
				auditCoursePhaseID)
		})
	}
}

func TestAuditMiddlewareLabelsTheCopyRoute(t *testing.T) {
	requireDeniedEvent(t, http.MethodPost, auditCopyRoute,
		presentation.AuditCopyAction,
		"POST "+auditCopyRoute,
		"")
}

func TestAuditMiddlewareIgnoresReads(t *testing.T) {
	sink := &recordingSink{}
	router := auditRouter(sink)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, auditCoursePhaseRoute+"/slots", nil))
	require.Equal(t, http.StatusUnauthorized, response.Code)

	time.Sleep(300 * time.Millisecond)
	require.Empty(t, sink.snapshot())
}

// noRowsDB stands in for the database so the real copy handler can run without one: it
// reports an unconfigured source phase, which is the handler's earliest return.
type noRowsDB struct{}

func (noRowsDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, pgx.ErrNoRows
}

func (noRowsDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, pgx.ErrNoRows
}

func (noRowsDB) QueryRow(context.Context, string, ...any) pgx.Row { return noRowsRow{} }

type noRowsRow struct{}

func (noRowsRow) Scan(...any) error { return pgx.ErrNoRows }

// auditCopyRouter reaches the copy handler itself, which the denied path never does.
func auditCopyRouter(sink audit.Sink) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := router.Group("/presentation/api")
	api.Use(audit.Middleware(sink))
	api.Use(auditActorMiddleware())
	service := presentation.NewService(db.New(noRowsDB{}), nil, nil, "", 0, 0, 0, nil)
	promptTypes.RegisterCopyEndpoint(
		api.Group("", audit.Describe(presentation.AuditCopyAction)),
		func(c *gin.Context) { c.Next() },
		&presentation.CopyHandler{Service: service},
	)
	return router
}

func postCopyRequest(t *testing.T, router *gin.Engine, body []byte) *httptest.ResponseRecorder {
	t.Helper()

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, auditCopyRoute, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)
	return response
}

func copyRequestBody(t *testing.T, source, target string) []byte {
	t.Helper()

	body, err := json.Marshal(promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: uuid.MustParse(source),
		TargetCoursePhaseID: uuid.MustParse(target),
	})
	require.NoError(t, err)
	return body
}

func TestHandlePhaseCopyRecordsExplicitEvent(t *testing.T) {
	sink := &recordingSink{}
	response := postCopyRequest(t, auditCopyRouter(sink),
		copyRequestBody(t, auditSourceCoursePhaseID, auditCoursePhaseID))
	require.Equal(t, http.StatusOK, response.Code)

	events := sink.waitForEvents(1)
	require.Len(t, events, 1)
	require.Equal(t, presentation.AuditCopyAction, events[0].Action)
	require.Equal(t, "coursePhase", events[0].EntityType)
	require.Equal(t, auditCoursePhaseID, events[0].EntityID)
	require.Equal(t, auditCoursePhaseID, events[0].CoursePhaseID)
	require.Equal(t, auditSourceCoursePhaseID, events[0].Metadata["sourceCoursePhaseID"])
	require.Equal(t, auditActorID, events[0].ActorID)
	require.Equal(t, audit.OutcomeSuccess, events[0].Outcome)
	require.Equal(t, http.StatusOK, events[0].HTTPStatus)

	time.Sleep(300 * time.Millisecond)
	require.Len(t, sink.snapshot(), 1)
}

// Core detects whether a phase service supports copying by posting a request whose source
// and target are the same phase, which is no copy and belongs in no audit log.
func TestHandlePhaseCopyIgnoresTheCopyabilityProbe(t *testing.T) {
	sink := &recordingSink{}
	response := postCopyRequest(t, auditCopyRouter(sink),
		copyRequestBody(t, auditCoursePhaseID, auditCoursePhaseID))
	require.Equal(t, http.StatusOK, response.Code)

	time.Sleep(300 * time.Millisecond)
	require.Empty(t, sink.snapshot())
}

// An empty body binds cleanly, leaving both phases blank, so there is no phase to pin an
// append-only entry to.
func TestHandlePhaseCopyIgnoresABlankBody(t *testing.T) {
	sink := &recordingSink{}
	response := postCopyRequest(t, auditCopyRouter(sink), []byte("{}"))
	require.Equal(t, http.StatusOK, response.Code)

	time.Sleep(300 * time.Millisecond)
	require.Empty(t, sink.snapshot())
}
