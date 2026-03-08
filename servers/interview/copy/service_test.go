package copy

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/interview/testutils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CopyServiceTestSuite struct {
	suite.Suite
	ctx     context.Context
	testDB  *testutils.TestDB
	cleanup func()
}

func (suite *CopyServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(suite.ctx, "../database_dumps/base.sql")
	require.NoError(suite.T(), err)
	suite.testDB = testDB
	suite.cleanup = cleanup

	CopyServiceSingleton = &CopyService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
}

func (suite *CopyServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *CopyServiceTestSuite) TestHandlePhaseCopyReturnsNoError() {
	handler := &SelfTeamCopyHandler{}
	c, _ := gin.CreateTestContext(nil)
	req := promptTypes.PhaseCopyRequest{}

	err := handler.HandlePhaseCopy(c, req)
	require.NoError(suite.T(), err)
}

func TestCopyServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CopyServiceTestSuite))
}
