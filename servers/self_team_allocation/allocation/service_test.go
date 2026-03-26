package allocation

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AllocationServiceTestSuite struct {
	suite.Suite
	ctx     context.Context
	testDB  *sdkTestUtils.TestDB[*db.Queries]
	cleanup func()
}

func (suite *AllocationServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
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
