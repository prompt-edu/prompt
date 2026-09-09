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
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/assessments"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/actionItem"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/categoryAssessment"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel"
	"github.com/prompt-edu/prompt/servers/assessment/categories"
	"github.com/prompt-edu/prompt/servers/assessment/competencies"
	"github.com/prompt-edu/prompt/servers/assessment/copy"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/feedbackItem"
	"github.com/prompt-edu/prompt/servers/assessment/privacy"
	"github.com/prompt-edu/prompt/servers/assessment/schemaModification"
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

// auditRouter rebuilds main.go's group layout, with the audit middleware attached to the
// api group before the course phase subgroup exists, because gin snapshots the handler
// chain when a subgroup is created. The actor middleware stands in for a token that
// authenticates but lacks the role the route requires, which is the denial the audit
// middleware is meant to capture.
func auditRouter(sink audit.Sink, authMiddleware func(allowedRoles ...string) gin.HandlerFunc, copyService *copy.CopyService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := router.Group("/assessment/api")
	api.Use(audit.Middleware(sink))
	api.Use(auditActorMiddleware())
	coursePhaseApi := api.Group("/course_phase/:coursePhaseID")

	queries := db.New(nil)
	var conn *pgxpool.Pool

	assessmentSchemaService := assessmentSchemas.NewAssessmentSchemaService(*queries, conn)
	coursePhaseConfigService := coursePhaseConfig.NewCoursePhaseConfigService(*queries, conn, assessmentSchemaService)
	schemaModificationService := schemaModification.NewSchemaModificationService(assessmentSchemaService, coursePhaseConfigService, *queries)

	competencyService := competencies.NewCompetencyService(*queries, conn, assessmentSchemaService, schemaModificationService)
	categoryService := categories.NewCategoryService(*queries, conn, assessmentSchemaService, schemaModificationService, coursePhaseConfigService)

	evaluationCompletionService := evaluationCompletion.NewEvaluationCompletionService(*queries, conn, coursePhaseConfig.GetTeamsForCoursePhase, coursePhaseConfigService)
	evaluationService := evaluations.NewEvaluationService(*queries, conn, evaluationCompletionService, coursePhaseConfigService)
	feedbackItemService := feedbackItem.NewFeedbackItemService(*queries, conn, evaluationCompletionService)

	assessmentCompletionService := assessmentCompletion.NewAssessmentCompletionService(*queries, conn, coursePhaseConfigService)
	categoryAssessmentService := categoryAssessment.NewCategoryAssessmentService(*queries, conn, assessmentCompletionService)
	actionItemService := actionItem.NewActionItemService(*queries, assessmentCompletionService, coursePhaseConfigService)
	scoreLevelService := scoreLevel.NewScoreLevelService(*queries)
	assessmentService := assessments.NewAssessmentService(*queries, conn, assessmentCompletionService, categoryAssessmentService, actionItemService, scoreLevelService, evaluationService, coursePhaseConfigService)

	competencies.RegisterRoutes(coursePhaseApi, competencyService, authMiddleware)
	categories.RegisterRoutes(coursePhaseApi, categoryService, authMiddleware)

	coursePhaseConfig.RegisterRoutes(coursePhaseApi, coursePhaseConfigService, authMiddleware)
	assessmentSchemas.RegisterRoutes(coursePhaseApi, assessmentSchemaService, authMiddleware)

	assessments.RegisterRoutes(coursePhaseApi, assessmentService, coursePhaseConfigService, authMiddleware)
	assessmentCompletion.RegisterRoutes(coursePhaseApi, assessmentCompletionService, coursePhaseConfigService, authMiddleware)
	categoryAssessment.RegisterRoutes(coursePhaseApi, categoryAssessmentService, coursePhaseConfigService, authMiddleware)
	actionItem.RegisterRoutes(coursePhaseApi, actionItemService, coursePhaseConfigService, authMiddleware)
	scoreLevel.RegisterRoutes(coursePhaseApi, scoreLevelService, authMiddleware)
	evaluations.RegisterRoutes(coursePhaseApi, evaluationService, authMiddleware)
	evaluationCompletion.RegisterRoutes(coursePhaseApi, evaluationCompletionService, authMiddleware)
	feedbackItem.RegisterRoutes(coursePhaseApi, feedbackItemService, authMiddleware)

	copy.RegisterRoutes(api, copyService, authMiddleware)
	privacy.RegisterRoutes(api, privacy.NewPrivacyService(*queries, conn))

	return router
}

func auditRouterWithoutDatabase(sink audit.Sink, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) *gin.Engine {
	queries := db.New(nil)
	var conn *pgxpool.Pool
	return auditRouter(sink, authMiddleware, copy.NewCopyService(*queries, conn))
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
	router := auditRouterWithoutDatabase(sink, promptSDK.AuthenticationMiddleware)

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
			action:    "Saved an assessment completion",
			actionKey: "POST " + auditCoursePhaseTemplate + "/student-assessment/completed",
		},
		{
			name:      "assessment completion updated",
			method:    http.MethodPut,
			path:      auditCoursePhaseRoute + "/student-assessment/completed",
			action:    "Saved an assessment completion",
			actionKey: "PUT " + auditCoursePhaseTemplate + "/student-assessment/completed",
		},
		{
			name:      "assessment marked complete",
			method:    http.MethodPost,
			path:      auditCoursePhaseRoute + "/student-assessment/completed/mark-complete",
			action:    "Marked an assessment as complete",
			actionKey: "POST " + auditCoursePhaseTemplate + "/student-assessment/completed/mark-complete",
		},
		{
			name:      "assessment unmarked",
			method:    http.MethodPut,
			path:      auditCoursePhaseRoute + "/student-assessment/completed" + auditParticipationSegment + "/unmark",
			action:    "Unmarked an assessment as complete",
			actionKey: "PUT " + auditCoursePhaseTemplate + "/student-assessment/completed" + auditParticipationParam + "/unmark",
		},
		{
			name:      "assessment completion deleted",
			method:    http.MethodDelete,
			path:      auditCoursePhaseRoute + "/student-assessment/completed" + auditParticipationSegment,
			action:    "Deleted an assessment completion",
			actionKey: "DELETE " + auditCoursePhaseTemplate + "/student-assessment/completed" + auditParticipationParam,
		},
		{
			name:      "my evaluation completion created",
			method:    http.MethodPost,
			path:      auditCoursePhaseRoute + "/evaluation/completed/my-completion",
			action:    "Saved my evaluation completion",
			actionKey: "POST " + auditCoursePhaseTemplate + "/evaluation/completed/my-completion",
		},
		{
			name:      "my evaluation completion updated",
			method:    http.MethodPut,
			path:      auditCoursePhaseRoute + "/evaluation/completed/my-completion",
			action:    "Saved my evaluation completion",
			actionKey: "PUT " + auditCoursePhaseTemplate + "/evaluation/completed/my-completion",
		},
		{
			name:      "my evaluation marked complete",
			method:    http.MethodPost,
			path:      auditCoursePhaseRoute + "/evaluation/completed/my-completion/mark-complete",
			action:    "Marked my evaluation as complete",
			actionKey: "POST " + auditCoursePhaseTemplate + "/evaluation/completed/my-completion/mark-complete",
		},
		{
			name:      "my evaluation unmarked",
			method:    http.MethodPut,
			path:      auditCoursePhaseRoute + "/evaluation/completed/my-completion/unmark",
			action:    "Unmarked my evaluation as complete",
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
			action:    "Released the assessment results",
			actionKey: "POST " + auditCoursePhaseTemplate + "/config/release",
		},
		{
			name:      "results unrelease",
			method:    http.MethodPost,
			path:      auditCoursePhaseRoute + "/config/unrelease",
			action:    "Unreleased the assessment results",
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
			action:    "Saved a category assessment",
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
	router := auditRouterWithoutDatabase(sink, promptSDK.AuthenticationMiddleware)

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

// Core detects copy support by posting a copy of a phase onto itself, forwarding the
// caller's token, so that probe must leave no trace in the audit log.
func TestHandlePhaseCopyProbeRecordsNothing(t *testing.T) {
	sink := &recordingSink{}
	router := auditRouterWithoutDatabase(sink, passThroughAuthMiddleware)

	phaseID := uuid.MustParse(auditCoursePhaseID)
	require.Equal(t, http.StatusOK, postCopyRequest(t, router, phaseID, phaseID).Code)

	time.Sleep(300 * time.Millisecond)
	require.Empty(t, sink.snapshot())
}

func TestHandlePhaseCopyRecordsScopedEvent(t *testing.T) {
	ctx := context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(ctx, "database_dumps/coursePhaseConfig.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	require.NoError(t, err)
	defer cleanup()

	source := uuid.MustParse(auditSourceCoursePhaseID)
	target := uuid.MustParse(auditCoursePhaseID)
	_, err = testDB.Conn.Exec(ctx, `INSERT INTO course_phase_config (course_phase_id) VALUES ($1)`, source)
	require.NoError(t, err)

	sink := &recordingSink{}
	router := auditRouter(sink, passThroughAuthMiddleware, copy.NewCopyService(*testDB.Queries, testDB.Conn))

	require.Equal(t, http.StatusOK, postCopyRequest(t, router, source, target).Code)

	events := sink.waitForEvents(1)
	require.Len(t, events, 1)
	require.Equal(t, "Copied course phase", events[0].Action)
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
			router := auditRouterWithoutDatabase(sink, promptSDK.AuthenticationMiddleware)

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, httptest.NewRequest(http.MethodPost, auditCoursePhaseRoute+path, nil))

			time.Sleep(300 * time.Millisecond)
			require.Empty(t, sink.snapshot())
		})
	}
}
