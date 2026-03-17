package timeframe

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TimeframeServiceTestSuite struct {
	suite.Suite
	ctx     context.Context
	testDB  *sdkTestUtils.TestDB[*db.Queries]
	cleanup func()
}

func (suite *TimeframeServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	require.NoError(suite.T(), err)

	suite.testDB = testDB
	suite.cleanup = cleanup

	TimeframeServiceSingleton = NewTimeframeService(*testDB.Queries, testDB.Conn)
}

func (suite *TimeframeServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *TimeframeServiceTestSuite) TestGetTimeframeReturnsExistingRecord() {
	coursePhaseID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	result, err := GetTimeframe(suite.ctx, coursePhaseID)

	require.NoError(suite.T(), err)
	require.True(suite.T(), result.TimeframeSet)
}

func (suite *TimeframeServiceTestSuite) TestGetTimeframeReturnsNotSet() {
	coursePhaseID := uuid.New()

	result, err := GetTimeframe(suite.ctx, coursePhaseID)

	require.NoError(suite.T(), err)
	require.False(suite.T(), result.TimeframeSet)
}

func (suite *TimeframeServiceTestSuite) TestSetTimeframePersists() {
	coursePhaseID := uuid.New()
	start := time.Now().UTC()
	end := start.Add(2 * time.Hour)

	err := SetTimeframe(suite.ctx, coursePhaseID, start, end)
	require.NoError(suite.T(), err)

	record, err := suite.testDB.Queries.GetTimeframe(suite.ctx, coursePhaseID)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), start.UTC().Truncate(time.Second), record.Starttime.Time.UTC().Truncate(time.Second))
	require.Equal(suite.T(), end.UTC().Truncate(time.Second), record.Endtime.Time.UTC().Truncate(time.Second))
}

func (suite *TimeframeServiceTestSuite) TestSetTimeframeValidatesRange() {
	err := SetTimeframe(suite.ctx, uuid.New(), time.Now().UTC(), time.Now().UTC().Add(-time.Hour))
	require.Error(suite.T(), err)
}

func TestTimeframeServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TimeframeServiceTestSuite))
}
