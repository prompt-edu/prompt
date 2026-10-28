package teams

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AssignmentServiceTestSuite struct {
	suite.Suite
	ctx               context.Context
	testDB            *sdkTestUtils.TestDB[*db.Queries]
	cleanup           func()
	assignmentService *AssignmentService
}

func (suite *AssignmentServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.ctx, "../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../database_dumps/base.sql")
	require.NoError(suite.T(), err)
	suite.testDB = testDB
	suite.cleanup = cleanup
	suite.assignmentService = NewAssignmentService(*testDB.Queries, testDB.Conn)
}

func (suite *AssignmentServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

// Regression test for GHSA-w97v-vx5j-3gjj: parallel joins cannot push a team past
// MaxTeamSize.
func (suite *AssignmentServiceTestSuite) TestConcurrentJoinsRespectTeamSize() {
	coursePhaseID := uuid.New()
	teamID := suite.createTeam(coursePhaseID)
	for range MaxTeamSize - 1 {
		require.NoError(suite.T(), suite.assignmentService.AssignTeam(suite.ctx, coursePhaseID, teamID, uuid.New(), "Early", "Member"))
	}

	const joiners = 10
	var wg sync.WaitGroup
	errs := make(chan error, joiners)
	for range joiners {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- suite.assignmentService.AssignTeam(suite.ctx, coursePhaseID, teamID, uuid.New(), "Racing", "Member")
		}()
	}
	wg.Wait()
	close(errs)

	succeeded := 0
	for err := range errs {
		if err == nil {
			succeeded++
			continue
		}
		require.ErrorIs(suite.T(), err, ErrTeamFull)
	}
	require.Equal(suite.T(), 1, succeeded)

	members, err := suite.testDB.Queries.CountTeamMembers(suite.ctx, teamID)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), int64(MaxTeamSize), members)
}

// A join that arrives while another transaction holds the team row must wait for it and
// then count the member that transaction added, instead of counting a stale snapshot.
func (suite *AssignmentServiceTestSuite) TestJoinWaitsForConcurrentJoin() {
	coursePhaseID := uuid.New()
	teamID := suite.createTeam(coursePhaseID)
	for range MaxTeamSize - 1 {
		require.NoError(suite.T(), suite.assignmentService.AssignTeam(suite.ctx, coursePhaseID, teamID, uuid.New(), "Early", "Member"))
	}

	// Play the first of two racing joins by hand: lock the team and take the last seat.
	tx, err := suite.testDB.Conn.Begin(suite.ctx)
	require.NoError(suite.T(), err)
	defer func() { _ = tx.Rollback(suite.ctx) }()
	qtx := suite.testDB.Queries.WithTx(tx)
	locked, err := qtx.LockTeamForAssignment(suite.ctx, db.LockTeamForAssignmentParams{ID: teamID, CoursePhaseID: coursePhaseID})
	require.NoError(suite.T(), err)
	require.Len(suite.T(), locked, 1)
	require.NoError(suite.T(), qtx.CreateOrUpdateAssignment(suite.ctx, db.CreateOrUpdateAssignmentParams{
		ID:                    uuid.New(),
		CourseParticipationID: uuid.New(),
		StudentFirstName:      "First",
		StudentLastName:       "Racer",
		TeamID:                teamID,
		CoursePhaseID:         coursePhaseID,
	}))

	done := make(chan error, 1)
	go func() {
		done <- suite.assignmentService.AssignTeam(suite.ctx, coursePhaseID, teamID, uuid.New(), "Second", "Racer")
	}()

	select {
	case err := <-done:
		suite.FailNow("join did not wait for the team lock", "returned %v", err)
	case <-time.After(300 * time.Millisecond):
	}

	require.NoError(suite.T(), tx.Commit(suite.ctx))
	require.ErrorIs(suite.T(), <-done, ErrTeamFull)

	members, err := suite.testDB.Queries.CountTeamMembers(suite.ctx, teamID)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), int64(MaxTeamSize), members)
}

func (suite *AssignmentServiceTestSuite) TestAssignTeamUnknownTeam() {
	err := suite.assignmentService.AssignTeam(suite.ctx, uuid.New(), uuid.New(), uuid.New(), "Lost", "Student")

	require.ErrorIs(suite.T(), err, ErrTeamNotFound)
}

func (suite *AssignmentServiceTestSuite) TestAssignTeamOtherPhase() {
	teamID := suite.createTeam(uuid.New())

	err := suite.assignmentService.AssignTeam(suite.ctx, uuid.New(), teamID, uuid.New(), "Wrong", "Phase")

	require.ErrorIs(suite.T(), err, ErrTeamNotFound)
}

func (suite *AssignmentServiceTestSuite) createTeam(coursePhaseID uuid.UUID) uuid.UUID {
	teamID := uuid.New()
	err := suite.testDB.Queries.CreateTeam(suite.ctx, db.CreateTeamParams{
		ID:            teamID,
		Name:          "Team " + teamID.String()[0:8],
		CoursePhaseID: coursePhaseID,
	})
	require.NoError(suite.T(), err)
	return teamID
}

func TestAssignmentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AssignmentServiceTestSuite))
}
