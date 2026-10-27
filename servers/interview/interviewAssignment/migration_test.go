package interviewAssignment

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const assignmentPhaseUniqueMigration = "0003_interview_assignment_phase_unique.up.sql"

type AssignmentPhaseMigrationTestSuite struct {
	suite.Suite
	ctx     context.Context
	testDB  *sdkTestUtils.TestDB[*db.Queries]
	cleanup func()
}

func (suite *AssignmentPhaseMigrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../db/migration/0001_interview_slots.up.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	require.NoError(suite.T(), err)

	suite.testDB = testDB
	suite.cleanup = cleanup
}

func (suite *AssignmentPhaseMigrationTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

// Bookings that already slipped through the race must not block the migration: the earliest
// booking per participation and phase is kept, and the phase is backfilled from the slot.
func (suite *AssignmentPhaseMigrationTestSuite) TestMigrationKeepsEarliestBookingPerPhase() {
	t := suite.T()

	conn, err := suite.testDB.Conn.Acquire(suite.ctx)
	require.NoError(t, err)
	defer conn.Release()

	applyMigration(t, suite.ctx, conn, "0002_interview_review.up.sql")

	phaseID := uuid.New()
	otherPhaseID := uuid.New()
	slotA, slotB, slotOther := uuid.New(), uuid.New(), uuid.New()
	for slotID, slotPhase := range map[uuid.UUID]uuid.UUID{slotA: phaseID, slotB: phaseID, slotOther: otherPhaseID} {
		_, err = conn.Exec(suite.ctx,
			"INSERT INTO interview_slot (id, course_phase_id, start_time, end_time) VALUES ($1, $2, now(), now() + interval '1 hour')",
			slotID, slotPhase)
		require.NoError(t, err)
	}

	participationID := uuid.New()
	earliest, later, otherPhase := uuid.New(), uuid.New(), uuid.New()
	_, err = conn.Exec(suite.ctx, `
		INSERT INTO interview_assignment (id, interview_slot_id, course_participation_id, assigned_at) VALUES
			($1, $4, $7, '2026-01-01 10:00:00+00'),
			($2, $5, $7, '2026-01-01 10:00:01+00'),
			($3, $6, $7, '2026-01-01 10:00:02+00')`,
		earliest, later, otherPhase, slotA, slotB, slotOther, participationID)
	require.NoError(t, err)

	applyMigration(t, suite.ctx, conn, assignmentPhaseUniqueMigration)

	rows, err := conn.Query(suite.ctx,
		"SELECT id, course_phase_id FROM interview_assignment WHERE course_participation_id = $1 ORDER BY assigned_at",
		participationID)
	require.NoError(t, err)
	remaining := map[uuid.UUID]uuid.UUID{}
	for rows.Next() {
		var id, coursePhaseID uuid.UUID
		require.NoError(t, rows.Scan(&id, &coursePhaseID))
		remaining[id] = coursePhaseID
	}
	require.NoError(t, rows.Err())
	require.Equal(t, map[uuid.UUID]uuid.UUID{earliest: phaseID, otherPhase: otherPhaseID}, remaining)

	_, err = conn.Exec(suite.ctx,
		"INSERT INTO interview_assignment (interview_slot_id, course_participation_id, course_phase_id) VALUES ($1, $2, $3)",
		slotB, participationID, phaseID)
	require.Error(t, err, "a second booking in the same phase must violate the new constraint")
}

func applyMigration(t *testing.T, ctx context.Context, conn *pgxpool.Conn, name string) {
	t.Helper()

	statements, err := os.ReadFile(filepath.Join("../db/migration", name))
	require.NoError(t, err)

	_, err = conn.Exec(ctx, string(statements))
	require.NoError(t, err)
}

func TestAssignmentPhaseMigrationTestSuite(t *testing.T) {
	suite.Run(t, new(AssignmentPhaseMigrationTestSuite))
}
