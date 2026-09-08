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
	"github.com/prompt-edu/prompt/servers/team_allocation/allocation"
	"github.com/prompt-edu/prompt/servers/team_allocation/config"
	"github.com/prompt-edu/prompt/servers/team_allocation/copy"
	"github.com/prompt-edu/prompt/servers/team_allocation/coursePhaseDeletion"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/privacy"
	"github.com/prompt-edu/prompt/servers/team_allocation/skills"
	"github.com/prompt-edu/prompt/servers/team_allocation/survey"
	teams "github.com/prompt-edu/prompt/servers/team_allocation/team"
	"github.com/prompt-edu/prompt/servers/team_allocation/tease"
	"github.com/stretchr/testify/require"
)

const (
	auditActorID             = "44444444-4444-4444-4444-444444444444"
	auditCoursePhaseID       = "10000000-0000-0000-0000-000000000001"
	auditSourceCoursePhaseID = "10000000-0000-0000-0000-000000000002"
	auditTeamID              = "10000000-0000-0000-0000-000000000003"
	auditSkillID             = "10000000-0000-0000-0000-000000000004"
	auditBase                = "/team-allocation/api"
	auditCoursePhaseRoute    = auditBase + "/course_phase/" + auditCoursePhaseID
	auditCoursePhaseTemplate = auditBase + "/course_phase/:coursePhaseID"
	auditTeaseRoute          = auditBase + "/tease/course_phase/" + auditCoursePhaseID
	auditTeaseTemplate       = auditBase + "/tease/course_phase/:coursePhaseID"
	auditCopyRoute           = auditBase + "/copy"
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

// auditRouter rebuilds main.go's group layout, with the audit middleware attached to the
// api group before the course phase subgroup exists, because gin snapshots the handler
// chain when a subgroup is created. The actor middleware stands in for a token that
// authenticates but lacks the role the route requires, which is the denial the audit
// middleware is meant to capture.
func auditRouter(sink audit.Sink, authMiddleware func(allowedRoles ...string) gin.HandlerFunc, copyService *copy.CopyService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := router.Group("/team-allocation/api")
	api.Use(audit.Middleware(sink))
	api.Use(auditActorMiddleware())
	coursePhaseApi := api.Group("/course_phase/:coursePhaseID")

	queries := db.New(nil)
	var conn *pgxpool.Pool
	skillsService := skills.NewSkillsService(*queries, conn)
	teamsService := teams.NewTeamsService(*queries, conn)
	surveyService := survey.NewSurveyService(*queries, conn)
	allocationService := allocation.NewAllocationService(*queries)
	teaseService := tease.NewTeaseService(*queries, conn)
	configService := config.NewConfigService(*queries, surveyService)
	privacyService := privacy.NewTeamsPrivacyService(*queries, conn)
	coursePhaseDeletionService := coursePhaseDeletion.NewCoursePhaseDeletionService(*queries, conn)

	skills.RegisterRoutes(coursePhaseApi, skillsService, authMiddleware)
	teams.RegisterRoutes(coursePhaseApi, teamsService, authMiddleware)
	survey.RegisterRoutes(coursePhaseApi, surveyService, authMiddleware)
	allocation.RegisterRoutes(coursePhaseApi, allocationService, authMiddleware)

	tease.RegisterRoutes(api, teaseService, authMiddleware)
	copy.RegisterRoutes(api, copyService, authMiddleware)

	config.RegisterRoutes(coursePhaseApi, configService, authMiddleware)

	privacy.RegisterRoutes(api, privacyService)
	coursePhaseDeletion.RegisterRoutes(coursePhaseApi, coursePhaseDeletionService)

	return router
}

func auditRouterWithoutDatabase(sink audit.Sink, authMiddleware func(allowedRoles ...string) gin.HandlerFunc) *gin.Engine {
	queries := db.New(nil)
	var conn *pgxpool.Pool
	return auditRouter(sink, authMiddleware, copy.NewCopyService(*queries, conn))
}

type auditRouteCase struct {
	name          string
	method        string
	url           string
	action        string
	actionKey     string
	coursePhaseID string
}

func requireDeniedEvent(t *testing.T, tc auditRouteCase) {
	t.Helper()

	sink := &recordingSink{}
	router := auditRouterWithoutDatabase(sink, promptSDK.AuthenticationMiddleware)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest(tc.method, tc.url, nil))
	require.Equal(t, http.StatusUnauthorized, resp.Code)

	events := sink.waitForEvents(1)
	require.Len(t, events, 1)
	require.Equal(t, tc.action, events[0].Action)
	require.Equal(t, tc.actionKey, events[0].ActionKey)
	require.Equal(t, audit.OutcomeDenied, events[0].Outcome)
	require.Equal(t, http.StatusUnauthorized, events[0].HTTPStatus)
	require.Equal(t, tc.coursePhaseID, events[0].CoursePhaseID)
	require.Equal(t, auditActorID, events[0].ActorID)
	require.Equal(t, "Ada Lovelace", events[0].ActorName)
}

func TestAuditMiddlewareUsesDescribedLabels(t *testing.T) {
	tests := []auditRouteCase{
		{
			name:          "team creation",
			method:        http.MethodPost,
			url:           auditCoursePhaseRoute + "/team",
			action:        "Created teams",
			actionKey:     "POST " + auditCoursePhaseTemplate + "/team",
			coursePhaseID: auditCoursePhaseID,
		},
		{
			name:          "student name import",
			method:        http.MethodPost,
			url:           auditCoursePhaseRoute + "/team/student-names",
			action:        "Added student names to team allocations",
			actionKey:     "POST " + auditCoursePhaseTemplate + "/team/student-names",
			coursePhaseID: auditCoursePhaseID,
		},
		{
			name:          "tutor import",
			method:        http.MethodPost,
			url:           auditCoursePhaseRoute + "/team/tutors",
			action:        "Imported tutors",
			actionKey:     "POST " + auditCoursePhaseTemplate + "/team/tutors",
			coursePhaseID: auditCoursePhaseID,
		},
		{
			name:          "tutor reassignment",
			method:        http.MethodPut,
			url:           auditCoursePhaseRoute + "/team/tutors/ab12cde",
			action:        "Updated a tutor's team",
			actionKey:     "PUT " + auditCoursePhaseTemplate + "/team/tutors/:universityLogin",
			coursePhaseID: auditCoursePhaseID,
		},
		{
			name:          "skill creation",
			method:        http.MethodPost,
			url:           auditCoursePhaseRoute + "/skill",
			action:        "Created skills",
			actionKey:     "POST " + auditCoursePhaseTemplate + "/skill",
			coursePhaseID: auditCoursePhaseID,
		},
		{
			name:          "survey submission",
			method:        http.MethodPost,
			url:           auditCoursePhaseRoute + "/survey/answers",
			action:        "Submitted survey answers",
			actionKey:     "POST " + auditCoursePhaseTemplate + "/survey/answers",
			coursePhaseID: auditCoursePhaseID,
		},
		{
			name:          "survey timeframe update",
			method:        http.MethodPut,
			url:           auditCoursePhaseRoute + "/survey/timeframe",
			action:        "Updated the survey timeframe",
			actionKey:     "PUT " + auditCoursePhaseTemplate + "/survey/timeframe",
			coursePhaseID: auditCoursePhaseID,
		},
		{
			name:          "phase data deletion",
			method:        http.MethodDelete,
			url:           auditCoursePhaseRoute,
			action:        "Deleted the team allocation phase data",
			actionKey:     "DELETE " + auditCoursePhaseTemplate,
			coursePhaseID: auditCoursePhaseID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requireDeniedEvent(t, tt)
		})
	}
}

// The TEASE routes hang off their own subgroup of api rather than off coursePhaseApi, so
// they only reach the middleware because main.go consolidated the three sibling api groups
// into one.
func TestAuditMiddlewareCoversTeaseRoutes(t *testing.T) {
	tests := []auditRouteCase{
		{
			name:          "workspace update",
			method:        http.MethodPut,
			url:           auditTeaseRoute + "/workspace",
			action:        "Updated the TEASE workspace",
			actionKey:     "PUT " + auditTeaseTemplate + "/workspace",
			coursePhaseID: auditCoursePhaseID,
		},
		{
			name:          "allocation publication",
			method:        http.MethodPost,
			url:           auditTeaseRoute + "/save",
			action:        "Published TEASE allocations",
			actionKey:     "POST " + auditTeaseTemplate + "/save",
			coursePhaseID: auditCoursePhaseID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requireDeniedEvent(t, tt)
		})
	}
}

func TestAuditMiddlewareUsesDerivedLabels(t *testing.T) {
	tests := []auditRouteCase{
		{
			name:          "team update",
			method:        http.MethodPut,
			url:           auditCoursePhaseRoute + "/team/" + auditTeamID,
			action:        "Updated team",
			actionKey:     "PUT " + auditCoursePhaseTemplate + "/team/:teamID",
			coursePhaseID: auditCoursePhaseID,
		},
		{
			name:          "team deletion",
			method:        http.MethodDelete,
			url:           auditCoursePhaseRoute + "/team/" + auditTeamID,
			action:        "Deleted team",
			actionKey:     "DELETE " + auditCoursePhaseTemplate + "/team/:teamID",
			coursePhaseID: auditCoursePhaseID,
		},
		{
			name:          "skill update",
			method:        http.MethodPut,
			url:           auditCoursePhaseRoute + "/skill/" + auditSkillID,
			action:        "Updated skill",
			actionKey:     "PUT " + auditCoursePhaseTemplate + "/skill/:skillID",
			coursePhaseID: auditCoursePhaseID,
		},
		{
			name:          "skill deletion",
			method:        http.MethodDelete,
			url:           auditCoursePhaseRoute + "/skill/" + auditSkillID,
			action:        "Deleted skill",
			actionKey:     "DELETE " + auditCoursePhaseTemplate + "/skill/:skillID",
			coursePhaseID: auditCoursePhaseID,
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
	router.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, auditCoursePhaseRoute+"/team", nil))
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
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(ctx, "database_dumps/skills.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	require.NoError(t, err)
	defer cleanup()

	sink := &recordingSink{}
	router := auditRouter(sink, passThroughAuthMiddleware, copy.NewCopyService(*testDB.Queries, testDB.Conn))

	source := uuid.MustParse(auditSourceCoursePhaseID)
	target := uuid.MustParse(auditCoursePhaseID)
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
