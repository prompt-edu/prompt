package config

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/testutils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ConfigRouterTestSuite struct {
	suite.Suite
	ctx     context.Context
	router  *gin.Engine
	testDB  *testutils.TestDB
	cleanup func()
}

func (suite *ConfigRouterTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()

	testDB, cleanup, err := testutils.SetupTestDB(suite.ctx, "../database_dumps/base.sql")
	require.NoError(suite.T(), err)
	suite.testDB = testDB
	suite.cleanup = cleanup

	ConfigServiceSingleton = &ConfigService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}

	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	setupConfigRouter(api, testutils.DefaultMockAuthMiddleware())
}

func (suite *ConfigRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *ConfigRouterTestSuite) TestGetConfig() {
	req, _ := http.NewRequest("GET", "/api/course_phase/11111111-1111-1111-1111-111111111111/config", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)

	var body map[string]bool
	err := json.Unmarshal(resp.Body.Bytes(), &body)
	require.NoError(suite.T(), err)
	require.True(suite.T(), body["surveyTimeframe"])
}

func (suite *ConfigRouterTestSuite) TestGetConfigInvalidCoursePhase() {
	req, _ := http.NewRequest("GET", "/api/course_phase/not-a-uuid/config", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusInternalServerError, resp.Code)
}

func TestConfigRouterTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigRouterTestSuite))
}
