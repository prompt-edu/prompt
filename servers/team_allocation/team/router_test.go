package teams

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/team/teamDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TeamRouterTestSuite struct {
	suite.Suite
	router      *gin.Engine
	suiteCtx    context.Context
	cleanup     func()
	teamService TeamsService
}

func (suite *TeamRouterTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/teams.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.teamService = TeamsService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	TeamsServiceSingleton = &suite.teamService
	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "lecturer@example.com", "03711111", "ab12cde")
	}
	setupTeamRouter(api, testMiddleware)
}

func (suite *TeamRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *TeamRouterTestSuite) TestGetAllTeams() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/team", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var response struct {
		Teams []promptTypes.Team `json:"teams"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(response.Teams), 0, "Should return a list of teams")
}

func (suite *TeamRouterTestSuite) TestGetAllTeamsInvalidCoursePhaseID() {
	req, _ := http.NewRequest("GET", "/api/course_phase/invalid-uuid/team", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *TeamRouterTestSuite) TestGetTeamByID() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	teamID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/team/"+teamID, nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var team promptTypes.Team
	err := json.Unmarshal(resp.Body.Bytes(), &team)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), team.ID, "Team ID should not be empty")
	assert.NotEmpty(suite.T(), team.Name, "Team name should not be empty")
}

func (suite *TeamRouterTestSuite) TestGetTeamByIDInvalidID() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/team/invalid-uuid", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *TeamRouterTestSuite) TestCreateTeams() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"

	createReq := teamDTO.CreateTeamsRequest{
		TeamNames: []string{"Team Epsilon", "Team Zeta"},
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+coursePhaseID+"/team", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
}

func (suite *TeamRouterTestSuite) TestCreateTeamsInvalidJSON() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	req, _ := http.NewRequest("POST", "/api/course_phase/"+coursePhaseID+"/team", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *TeamRouterTestSuite) TestUpdateTeamName() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	teamID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	updateReq := teamDTO.UpdateTeamRequest{
		NewTeamName: "Updated Team Alpha",
	}
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/course_phase/"+coursePhaseID+"/team/"+teamID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *TeamRouterTestSuite) TestUpdateTeamNameInvalidID() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"

	updateReq := teamDTO.UpdateTeamRequest{
		NewTeamName: "Updated Team",
	}
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/course_phase/"+coursePhaseID+"/team/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *TeamRouterTestSuite) TestDeleteTeam() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	teamID := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+coursePhaseID+"/team/"+teamID, nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *TeamRouterTestSuite) TestDeleteTeamInvalidID() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"

	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+coursePhaseID+"/team/invalid-uuid", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func TestTeamRouterTestSuite(t *testing.T) {
	suite.Run(t, new(TeamRouterTestSuite))
}
