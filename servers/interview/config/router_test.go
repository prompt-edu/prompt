package config

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt/servers/interview/testutils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ConfigRouterTestSuite struct {
	suite.Suite
	ctx     context.Context
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
}

func (suite *ConfigRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *ConfigRouterTestSuite) newRouter(identity testutils.MockIdentity) *gin.Engine {
	router := gin.Default()
	api := router.Group("/api/course_phase/:coursePhaseID")
	setupConfigRouter(api, testutils.NewMockAuthMiddleware(identity))
	return router
}

func (suite *ConfigRouterTestSuite) TestGetPhaseConfigRoute() {
	router := suite.newRouter(testutils.MockIdentity{})
	req, _ := http.NewRequest("GET", "/api/course_phase/11111111-1111-1111-1111-111111111111/config", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)

	var config map[string]bool
	err := json.Unmarshal(resp.Body.Bytes(), &config)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), config)
}

func TestConfigRouterTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigRouterTestSuite))
}
