package timeframe

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const timeframeTimestamptzMigration = "0003_timeframe_timestamptz.up.sql"

type TimeframeMigrationTestSuite struct {
	suite.Suite
	ctx     context.Context
	testDB  *sdkTestUtils.TestDB[*db.Queries]
	cleanup func()
}

func (suite *TimeframeMigrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../db/migration/0001_schema.up.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	require.NoError(suite.T(), err)

	suite.testDB = testDB
	suite.cleanup = cleanup
}

func (suite *TimeframeMigrationTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *TimeframeMigrationTestSuite) TestTimestamptzMigrationKeepsStoredInstants() {
	t := suite.T()

	conn, err := suite.testDB.Conn.Acquire(suite.ctx)
	require.NoError(t, err)
	defer conn.Release()

	_, err = conn.Exec(suite.ctx, "SET TIME ZONE 'Europe/Berlin'")
	require.NoError(t, err)

	for _, path := range migrationsBetweenSchemaAnd(t, timeframeTimestamptzMigration) {
		applyMigration(t, suite.ctx, conn, path)
	}

	coursePhaseID := uuid.New()
	_, err = conn.Exec(suite.ctx,
		"INSERT INTO timeframe (course_phase_id, starttime, endtime) VALUES ($1, '2026-01-15 12:34:56.123456', '2026-06-15 12:34:56.123456')",
		coursePhaseID)
	require.NoError(t, err)

	applyMigration(t, suite.ctx, conn, filepath.Join("../db/migration", timeframeTimestamptzMigration))

	var startType, endType string
	err = conn.QueryRow(suite.ctx,
		`SELECT
			(SELECT data_type FROM information_schema.columns WHERE table_name = 'timeframe' AND column_name = 'starttime'),
			(SELECT data_type FROM information_schema.columns WHERE table_name = 'timeframe' AND column_name = 'endtime')`).
		Scan(&startType, &endType)
	require.NoError(t, err)
	require.Equal(t, "timestamp with time zone", startType)
	require.Equal(t, "timestamp with time zone", endType)

	var startMatches, endMatches bool
	err = conn.QueryRow(suite.ctx,
		`SELECT starttime = TIMESTAMPTZ '2026-01-15 12:34:56.123456+00',
		        endtime = TIMESTAMPTZ '2026-06-15 12:34:56.123456+00'
		 FROM timeframe WHERE course_phase_id = $1`, coursePhaseID).
		Scan(&startMatches, &endMatches)
	require.NoError(t, err)
	require.True(t, startMatches)
	require.True(t, endMatches)
}

// The initial schema is applied by the suite's test database setup.
func migrationsBetweenSchemaAnd(t *testing.T, target string) []string {
	t.Helper()

	paths, err := filepath.Glob("../db/migration/*.up.sql")
	require.NoError(t, err)
	sort.Strings(paths)
	require.NotEmpty(t, paths)

	pending := make([]string, 0, len(paths))
	for _, path := range paths[1:] {
		if filepath.Base(path) >= target {
			break
		}
		pending = append(pending, path)
	}
	return pending
}

func applyMigration(t *testing.T, ctx context.Context, conn *pgxpool.Conn, path string) {
	t.Helper()

	statements, err := os.ReadFile(path)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, string(statements))
	require.NoError(t, err)
}

func TestTimeframeMigrationTestSuite(t *testing.T) {
	suite.Run(t, new(TimeframeMigrationTestSuite))
}
