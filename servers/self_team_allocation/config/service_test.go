package config

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/testutils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ConfigServiceTestSuite struct {
	suite.Suite
	ctx     context.Context
	testDB  *testutils.TestDB
	cleanup func()
}

func (suite *ConfigServiceTestSuite) SetupSuite() {
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

func (suite *ConfigServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *ConfigServiceTestSuite) newContext(coursePhaseID uuid.UUID) *gin.Context {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/config", nil)
	c.Params = []gin.Param{{Key: "coursePhaseID", Value: coursePhaseID.String()}}
	return c
}

func (suite *ConfigServiceTestSuite) TestHandlePhaseConfigWithTimeframe() {
	handler := ConfigHandler{}
	c := suite.newContext(uuid.MustParse("11111111-1111-1111-1111-111111111111"))

	configMap, err := handler.HandlePhaseConfig(c)

	require.NoError(suite.T(), err)
	require.True(suite.T(), configMap["surveyTimeframe"])
}

func (suite *ConfigServiceTestSuite) TestHandlePhaseConfigWithoutTimeframe() {
	handler := ConfigHandler{}
	c := suite.newContext(uuid.New())

	configMap, err := handler.HandlePhaseConfig(c)

	require.NoError(suite.T(), err)
	require.False(suite.T(), configMap["surveyTimeframe"])
}

func TestConfigServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigServiceTestSuite))
}
