package config

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ConfigServiceTestSuite struct {
	suite.Suite
	ctx     context.Context
	testDB  *sdkTestUtils.TestDB[*db.Queries]
	cleanup func()
}

func (suite *ConfigServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
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

func (suite *ConfigServiceTestSuite) TestHandlePhaseConfigReturnsEmptyMap() {
	handler := &ConfigHandler{}
	config, err := handler.HandlePhaseConfig(nil)

	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), config)
	require.Equal(suite.T(), map[string]bool{}, config)
}

func TestConfigServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigServiceTestSuite))
}
