package teamScope

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/stretchr/testify/assert"
)

var (
	coursePhaseID = uuid.New()
	teamPhaseID   = uuid.New()
	member        = uuid.New()
	outsider      = uuid.New()
)

type upstream struct {
	hasTeamsInput  bool
	teamsStatus    int
	teamAuthHeader string
	teamCalls      int
}

func startUpstream(t *testing.T, u *upstream) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	t.Setenv("SERVER_CORE_HOST", server.URL)

	router.GET("/api/course_phases/:id/course_phase_data", func(c *gin.Context) {
		resolutions := []promptSDK.Resolution{}
		if u.hasTeamsInput {
			resolutions = append(resolutions, promptSDK.Resolution{
				DtoName:       teamsDtoName,
				BaseURL:       server.URL + "/team-phase/api",
				EndpointPath:  "/team",
				CoursePhaseID: teamPhaseID,
			})
		}
		c.JSON(http.StatusOK, gin.H{"prevData": gin.H{}, "resolutions": resolutions})
	})
	router.GET("/team-phase/api/course_phase/"+teamPhaseID.String()+"/team", func(c *gin.Context) {
		u.teamCalls++
		u.teamAuthHeader = c.GetHeader("Authorization")
		if u.teamsStatus != http.StatusOK {
			c.Status(u.teamsStatus)
			return
		}
		c.JSON(http.StatusOK, gin.H{"teams": []gin.H{{
			"id":      uuid.New(),
			"name":    "Team",
			"members": []gin.H{{"id": member, "firstName": "A", "lastName": "B"}},
		}}})
	})
}

func editor() keycloakTokenVerifier.TokenUser {
	return keycloakTokenVerifier.TokenUser{Roles: map[string]bool{}, IsEditor: true}
}

func lecturer() keycloakTokenVerifier.TokenUser {
	return keycloakTokenVerifier.TokenUser{Roles: map[string]bool{}, IsEditor: true, IsLecturer: true}
}

func scopedRouter(user keycloakTokenVerifier.TokenUser) *gin.Engine {
	router := gin.New()
	api := router.Group("/course_phase/:coursePhaseID", func(c *gin.Context) {
		keycloakTokenVerifier.SetTokenUser(c, user)
	})
	api.GET("/participants", Resolve(), func(c *gin.Context) {
		c.JSON(http.StatusOK, Filter(c, []uuid.UUID{member, outsider}, func(id uuid.UUID) uuid.UUID { return id }))
	})
	api.GET("/participants/:courseParticipationID", RequireParticipant(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	api.POST("/write", RequireBodyParticipants(), func(c *gin.Context) {
		var body map[string]any
		if err := c.BindJSON(&body); err != nil {
			return
		}
		c.Status(http.StatusOK)
	})
	return router
}

func get(router *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/course_phase/"+coursePhaseID.String()+path, nil)
	req.Header.Set("Authorization", "Bearer caller")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func post(router *gin.Engine, body any) *httptest.ResponseRecorder {
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/course_phase/"+coursePhaseID.String()+"/write", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer caller")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func listed(t *testing.T, resp *httptest.ResponseRecorder) []uuid.UUID {
	t.Helper()
	assert.Equal(t, http.StatusOK, resp.Code)
	var ids []uuid.UUID
	assert.NoError(t, json.Unmarshal(resp.Body.Bytes(), &ids))
	return ids
}

func TestTutorSeesOnlyTheirTeam(t *testing.T) {
	u := &upstream{hasTeamsInput: true, teamsStatus: http.StatusOK}
	startUpstream(t, u)
	router := scopedRouter(editor())

	assert.Equal(t, []uuid.UUID{member}, listed(t, get(router, "/participants")))
	assert.Equal(t, "Bearer caller", u.teamAuthHeader)

	assert.Equal(t, http.StatusOK, get(router, "/participants/"+member.String()).Code)
	assert.Equal(t, http.StatusForbidden, get(router, "/participants/"+outsider.String()).Code)

	assert.Equal(t, http.StatusOK, post(router, gin.H{"courseParticipationID": member}).Code)
	assert.Equal(t, http.StatusForbidden, post(router, gin.H{"courseParticipationID": outsider}).Code)
	assert.Equal(t, http.StatusOK, post(router, gin.H{"courseParticipationIDs": []uuid.UUID{member}}).Code)
	assert.Equal(t, http.StatusForbidden, post(router, gin.H{"courseParticipationIDs": []uuid.UUID{member, outsider}}).Code)
	assert.Equal(t, http.StatusForbidden, post(router, gin.H{"courseParticipationID": outsider, "courseParticipationIDs": []uuid.UUID{member}}).Code)
	assert.Equal(t, http.StatusOK, post(router, gin.H{"courseParticipationID": member, "courseParticipationIDs": []uuid.UUID{member}}).Code)
}

func TestEditorDeniedByTeamPhaseSeesNobody(t *testing.T) {
	startUpstream(t, &upstream{hasTeamsInput: true, teamsStatus: http.StatusForbidden})
	router := scopedRouter(editor())

	assert.Empty(t, listed(t, get(router, "/participants")))
	assert.Equal(t, http.StatusForbidden, get(router, "/participants/"+member.String()).Code)
	assert.Equal(t, http.StatusForbidden, post(router, gin.H{"courseParticipationID": member}).Code)
}

func TestEditorWithoutTeamsInputIsUnrestricted(t *testing.T) {
	u := &upstream{hasTeamsInput: false}
	startUpstream(t, u)
	router := scopedRouter(editor())

	assert.Equal(t, []uuid.UUID{member, outsider}, listed(t, get(router, "/participants")))
	assert.Equal(t, http.StatusOK, post(router, gin.H{"courseParticipationID": outsider}).Code)
	assert.Zero(t, u.teamCalls)
}

func TestTeamPhaseFailureIsBadGateway(t *testing.T) {
	startUpstream(t, &upstream{hasTeamsInput: true, teamsStatus: http.StatusInternalServerError})
	router := scopedRouter(editor())

	assert.Equal(t, http.StatusBadGateway, get(router, "/participants").Code)
	assert.Equal(t, http.StatusBadGateway, get(router, "/participants/"+member.String()).Code)
}

func TestLecturersAndAdminsAreUnrestricted(t *testing.T) {
	u := &upstream{hasTeamsInput: true, teamsStatus: http.StatusOK}
	startUpstream(t, u)
	admin := keycloakTokenVerifier.TokenUser{Roles: map[string]bool{promptSDK.PromptAdmin: true}, IsEditor: true}

	for _, user := range []keycloakTokenVerifier.TokenUser{lecturer(), admin} {
		router := scopedRouter(user)
		assert.Equal(t, []uuid.UUID{member, outsider}, listed(t, get(router, "/participants")))
		assert.Equal(t, http.StatusOK, get(router, "/participants/"+outsider.String()).Code)
	}
	assert.Zero(t, u.teamCalls)
}

func TestFilterWithoutMiddlewareAllowsNobody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.Empty(t, Filter(c, []uuid.UUID{member}, func(id uuid.UUID) uuid.UUID { return id }))
}

func TestResolutionURLRejectsUnsafeTargets(t *testing.T) {
	valid := promptSDK.Resolution{BaseURL: "http://team-allocation/api", EndpointPath: "/team", CoursePhaseID: teamPhaseID}
	got, err := resolutionURL(valid)
	assert.NoError(t, err)
	assert.Equal(t, "http://team-allocation/api/course_phase/"+teamPhaseID.String()+"/team", got)

	for _, resolution := range []promptSDK.Resolution{
		{BaseURL: "file:///etc", EndpointPath: "/team", CoursePhaseID: teamPhaseID},
		{BaseURL: "http://user:pass@host", EndpointPath: "/team", CoursePhaseID: teamPhaseID},
		{BaseURL: "http://host", EndpointPath: "/../admin", CoursePhaseID: teamPhaseID},
		{BaseURL: "http://host", EndpointPath: "/%2e%2e/admin", CoursePhaseID: teamPhaseID},
	} {
		_, err := resolutionURL(resolution)
		assert.Error(t, err, resolution.BaseURL+resolution.EndpointPath)
	}
}
