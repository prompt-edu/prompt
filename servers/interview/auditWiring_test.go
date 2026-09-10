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
	"github.com/prompt-edu/prompt/servers/interview/copy"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
	interview_assignment "github.com/prompt-edu/prompt/servers/interview/interviewAssignment"
	interview_review "github.com/prompt-edu/prompt/servers/interview/interviewReview"
	interview_slot "github.com/prompt-edu/prompt/servers/interview/interviewSlot"
	"github.com/stretchr/testify/require"
)

const (
	auditActorID             = "44444444-4444-4444-4444-444444444444"
	auditCoursePhaseID       = "10000000-0000-0000-0000-000000000001"
	auditSourceCoursePhaseID = "10000000-0000-0000-0000-000000000002"
	auditCourseParticipation = "10000000-0000-0000-0000-000000000003"
	auditCoursePhaseRoute    = "/interview/api/course_phase/" + auditCoursePhaseID
	auditCoursePhaseTemplate = "/interview/api/course_phase/:coursePhaseID"
	auditCopyRoute           = "/interview/api/copy"
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

	api := router.Group("interview/api")
	api.Use(audit.Middleware(sink))
	api.Use(auditActorMiddleware())
	coursePhaseApi := api.Group("/course_phase/:coursePhaseID")

	queries := db.New(nil)
	var conn *pgxpool.Pool
	interviewSlotService := interview_slot.NewInterviewSlotService(*queries, conn)
	interviewAssignmentService := interview_assignment.NewInterviewAssignmentService(*queries, conn)
	interviewReviewService := interview_review.NewInterviewReviewService(*queries)

	copy.RegisterRoutes(api, authMiddleware)
	interview_slot.RegisterRoutes(coursePhaseApi, interviewSlotService, authMiddleware)
	interview_assignment.RegisterRoutes(coursePhaseApi, interviewAssignmentService, authMiddleware)
	interview_review.RegisterRoutes(coursePhaseApi, interviewReviewService, authMiddleware)

	return router
}

func requireDeniedEvent(t *testing.T, method, url, action, actionKey string) {
	t.Helper()

	sink := &recordingSink{}
	router := auditRouter(sink, promptSDK.AuthenticationMiddleware)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest(method, url, nil))
	require.Equal(t, http.StatusUnauthorized, resp.Code)

	events := sink.waitForEvents(1)
	require.Len(t, events, 1)
	require.Equal(t, action, events[0].Action)
	require.Equal(t, actionKey, events[0].ActionKey)
	require.Equal(t, audit.OutcomeDenied, events[0].Outcome)
	require.Equal(t, http.StatusUnauthorized, events[0].HTTPStatus)
	require.Equal(t, auditCoursePhaseID, events[0].CoursePhaseID)
	require.Equal(t, auditActorID, events[0].ActorID)
	require.Equal(t, "Ada Lovelace", events[0].ActorName)
}

func TestAuditMiddlewareUsesDescribedLabels(t *testing.T) {
	t.Run("batch slot creation", func(t *testing.T) {
		requireDeniedEvent(t, http.MethodPost,
			auditCoursePhaseRoute+"/interview-slots/batch",
			"Created interview slots in bulk",
			"POST "+auditCoursePhaseTemplate+"/interview-slots/batch")
	})

	t.Run("admin assignment", func(t *testing.T) {
		requireDeniedEvent(t, http.MethodPost,
			auditCoursePhaseRoute+"/interview-assignments/admin",
			"Assigned a student to an interview slot",
			"POST "+auditCoursePhaseTemplate+"/interview-assignments/admin")
	})
}

func TestAuditMiddlewareUsesDerivedLabels(t *testing.T) {
	t.Run("slot creation", func(t *testing.T) {
		requireDeniedEvent(t, http.MethodPost,
			auditCoursePhaseRoute+"/interview-slots",
			"Created interview slot",
			"POST "+auditCoursePhaseTemplate+"/interview-slots")
	})

	t.Run("review upsert", func(t *testing.T) {
		requireDeniedEvent(t, http.MethodPut,
			auditCoursePhaseRoute+"/interview-review/"+auditCourseParticipation,
			"Updated interview review",
			"PUT "+auditCoursePhaseTemplate+"/interview-review/:courseParticipationID")
	})
}

func TestAuditMiddlewareIgnoresReads(t *testing.T) {
	sink := &recordingSink{}
	router := auditRouter(sink, promptSDK.AuthenticationMiddleware)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, auditCoursePhaseRoute+"/interview-slots", nil))
	require.Equal(t, http.StatusUnauthorized, resp.Code)

	time.Sleep(300 * time.Millisecond)
	require.Empty(t, sink.snapshot())
}

// Interview slots and reviews are tied to their phase, so a copy carries nothing
// over and must not appear in the audit log claiming that it did.
func TestHandlePhaseCopyRecordsNothing(t *testing.T) {
	sink := &recordingSink{}
	router := auditRouter(sink, passThroughAuthMiddleware)

	body, err := json.Marshal(promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: uuid.MustParse(auditSourceCoursePhaseID),
		TargetCoursePhaseID: uuid.MustParse(auditCoursePhaseID),
	})
	require.NoError(t, err)

	resp := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, auditCopyRoute, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(resp, request)
	require.Equal(t, http.StatusOK, resp.Code)

	time.Sleep(300 * time.Millisecond)
	require.Empty(t, sink.snapshot())
}
