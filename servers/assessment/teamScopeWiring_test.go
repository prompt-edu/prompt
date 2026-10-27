package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/assessment/assessments"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/actionItem"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/categoryAssessment"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/feedbackItem"
	"github.com/stretchr/testify/assert"
)

var (
	scopeTeamPhaseID = uuid.New()
	scopeMember      = uuid.New()
	scopeOutsider    = uuid.New()
)

type scopedRoute struct {
	method string
	path   string
	body   string
}

func participantBody(id uuid.UUID) string {
	return `{"courseParticipationID":"` + id.String() + `"}`
}

func scopedRoutes(participant uuid.UUID) []scopedRoute {
	id := participant.String()
	single := participantBody(participant)
	batch := `{"courseParticipationIDs":["` + scopeMember.String() + `","` + id + `"]}`
	return []scopedRoute{
		{http.MethodGet, "/config/participations", ""},
		{http.MethodGet, "/student-assessment", ""},
		{http.MethodPost, "/student-assessment", single},
		{http.MethodPost, "/student-assessment/independent", single},
		{http.MethodGet, "/student-assessment/" + id + "/export", ""},
		{http.MethodGet, "/student-assessment/" + id, ""},
		{http.MethodGet, "/student-assessment/course-participation/" + id, ""},
		{http.MethodDelete, "/student-assessment/independent/" + uuid.NewString(), ""},
		{http.MethodPost, "/category-assessment", single},
		{http.MethodGet, "/student-assessment/scoreLevel", ""},
		{http.MethodGet, "/student-assessment/scoreLevel/" + id, ""},
		{http.MethodGet, "/student-assessment/action-item/action", ""},
		{http.MethodGet, "/student-assessment/action-item/action/course-participation/" + id, ""},
		{http.MethodGet, "/student-assessment/action-item/course-participation/" + id, ""},
		{http.MethodPost, "/student-assessment/action-item", single},
		{http.MethodPut, "/student-assessment/action-item/" + uuid.NewString(), single},
		{http.MethodDelete, "/student-assessment/action-item/" + uuid.NewString(), ""},
		{http.MethodGet, "/student-assessment/completed/grade", ""},
		{http.MethodGet, "/student-assessment/completed/grade/course-participation/" + id, ""},
		{http.MethodGet, "/student-assessment/completed", ""},
		{http.MethodPost, "/student-assessment/completed", single},
		{http.MethodPut, "/student-assessment/completed", single},
		{http.MethodPost, "/student-assessment/completed/mark-complete", single},
		{http.MethodPost, "/student-assessment/completed/mark-complete/batch", batch},
		{http.MethodPut, "/student-assessment/completed/unmark/batch", batch},
		{http.MethodGet, "/student-assessment/completed/course-participation/" + id, ""},
		{http.MethodPut, "/student-assessment/completed/course-participation/" + id + "/unmark", ""},
		{http.MethodDelete, "/student-assessment/completed/course-participation/" + id, ""},
		{http.MethodGet, "/evaluation/completed", ""},
		{http.MethodGet, "/evaluation/feedback-items/course-participation/" + id, ""},
	}
}

func startTeamPhase(t *testing.T, teamStatus int) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	t.Setenv("SERVER_CORE_HOST", server.URL)

	router.GET("/api/course_phases/:id/course_phase_data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"prevData": gin.H{}, "resolutions": []promptSDK.Resolution{{
			DtoName:       "teams",
			BaseURL:       server.URL + "/team-phase/api",
			EndpointPath:  "/team",
			CoursePhaseID: scopeTeamPhaseID,
		}}})
	})
	router.GET("/team-phase/api/course_phase/:coursePhaseID/team", func(c *gin.Context) {
		if teamStatus != http.StatusOK {
			c.Status(teamStatus)
			return
		}
		c.JSON(http.StatusOK, gin.H{"teams": []gin.H{{"id": uuid.New(), "name": "Team", "members": []gin.H{{"id": scopeMember}}}}})
	})
}

func tutorRouter() *gin.Engine {
	router := gin.New()
	api := router.Group("/assessment/api/course_phase/:coursePhaseID")
	tutorAuth := func(...string) gin.HandlerFunc {
		return func(c *gin.Context) {
			keycloakTokenVerifier.SetTokenUser(c, keycloakTokenVerifier.TokenUser{Roles: map[string]bool{}, IsEditor: true})
		}
	}
	guard := (*coursePhaseConfig.CoursePhaseConfigService)(nil)

	coursePhaseConfig.RegisterRoutes(api, nil, tutorAuth)
	assessments.RegisterRoutes(api, nil, guard, tutorAuth)
	assessmentCompletion.RegisterRoutes(api, nil, guard, tutorAuth)
	categoryAssessment.RegisterRoutes(api, nil, guard, tutorAuth)
	actionItem.RegisterRoutes(api, nil, guard, tutorAuth)
	scoreLevel.RegisterRoutes(api, nil, tutorAuth)
	evaluationCompletion.RegisterRoutes(api, nil, tutorAuth)
	feedbackItem.RegisterRoutes(api, nil, tutorAuth)
	return router
}

func serve(router *gin.Engine, route scopedRoute) int {
	return serveRecorded(router, route).Code
}

func serveRecorded(router *gin.Engine, route scopedRoute) *httptest.ResponseRecorder {
	req := httptest.NewRequest(route.method, "/assessment/api/course_phase/"+uuid.NewString()+route.path, strings.NewReader(route.body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func TestEveryEditorParticipantRouteResolvesTheTeamScope(t *testing.T) {
	startTeamPhase(t, http.StatusInternalServerError)
	router := tutorRouter()

	for _, route := range scopedRoutes(scopeOutsider) {
		assert.Equal(t, http.StatusBadGateway, serve(router, route), route.method+" "+route.path)
	}
}

func TestTutorCannotReachParticipantsOutsideTheirTeam(t *testing.T) {
	startTeamPhase(t, http.StatusOK)
	router := tutorRouter()

	for _, route := range scopedRoutes(scopeOutsider) {
		if route.body == "" && !strings.Contains(route.path, scopeOutsider.String()) {
			continue
		}
		assert.Equal(t, http.StatusForbidden, serve(router, route), route.method+" "+route.path)
	}
}

func TestTutorCannotSmuggleAnOutsiderNextToTheirTeam(t *testing.T) {
	startTeamPhase(t, http.StatusOK)
	router := tutorRouter()
	mixed := `{"courseParticipationID":"` + scopeOutsider.String() + `","courseParticipationIDs":["` + scopeMember.String() + `"]}`

	for _, route := range scopedRoutes(scopeOutsider) {
		if route.body != participantBody(scopeOutsider) {
			continue
		}
		route.body = mixed
		assert.Equal(t, http.StatusForbidden, serve(router, route), route.method+" "+route.path)
	}
}

func TestEditorDeniedByTeamPhaseGetsEmptyLists(t *testing.T) {
	startTeamPhase(t, http.StatusForbidden)
	router := tutorRouter()

	for _, path := range []string{"/config/participations", "/config/teams"} {
		resp := serveRecorded(router, scopedRoute{http.MethodGet, path, ""})
		assert.Equal(t, http.StatusOK, resp.Code, path)
		assert.JSONEq(t, "[]", resp.Body.String(), path)
	}
}
