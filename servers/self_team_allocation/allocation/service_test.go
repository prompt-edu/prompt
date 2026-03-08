package allocation

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/testutils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AllocationServiceTestSuite struct {
	suite.Suite
	ctx     context.Context
	testDB  *testutils.TestDB
	cleanup func()
}

func (suite *AllocationServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(suite.ctx, "../database_dumps/base.sql")
	require.NoError(suite.T(), err)

	suite.testDB = testDB
	suite.cleanup = cleanup

	AllocationServiceSingleton = &AllocationService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
}

func (suite *AllocationServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *AllocationServiceTestSuite) TestGetAllAllocations() {
	coursePhaseID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	allocations, err := GetAllAllocations(suite.ctx, coursePhaseID)

	require.NoError(suite.T(), err)
	require.Len(suite.T(), allocations, 3)
}

func (suite *AllocationServiceTestSuite) TestGetAllocationByCourseParticipationID() {
	coursePhaseID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	participantID := uuid.MustParse("aaaa1111-1111-1111-1111-111111111111")

	teamID, err := GetAllocationByCourseParticipationID(suite.ctx, participantID, coursePhaseID)

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), teamID)
}

func (suite *AllocationServiceTestSuite) TestGetAllocationByCourseParticipationIDNotFound() {
	coursePhaseID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	participantID := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")

	_, err := GetAllocationByCourseParticipationID(suite.ctx, participantID, coursePhaseID)

	require.Error(suite.T(), err)
}

func TestAllocationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AllocationServiceTestSuite))
}
