package allocation

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AllocationServiceTestSuite struct {
	suite.Suite
	suiteCtx          context.Context
	cleanup           func()
	allocationService AllocationService
}

func (suite *AllocationServiceTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/allocations.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.allocationService = AllocationService{
		queries:             *testDB.Queries,
		conn:                testDB.Conn,
		resolveParticipants: stubParticipants(tutorFreeParticip, deltaParticip, epsilonParticip, staffFreeParticip),
	}
	AllocationServiceSingleton = &suite.allocationService
}

func (suite *AllocationServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *AllocationServiceTestSuite) TestGetAllAllocations() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")

	allocations, err := GetAllAllocations(suite.suiteCtx, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(allocations), 0, "Expected at least one allocation")

	for _, allocation := range allocations {
		assert.NotEmpty(suite.T(), allocation.CourseParticipationID, "CourseParticipationID should not be empty")
		assert.NotEmpty(suite.T(), allocation.TeamAllocation, "TeamAllocation should not be empty")
	}
}

func (suite *AllocationServiceTestSuite) TestGetAllAllocationsNonExistentCoursePhase() {
	nonExistentID := uuid.New()

	allocations, err := GetAllAllocations(suite.suiteCtx, nonExistentID)
	assert.NoError(suite.T(), err, "Should not error for non-existent course phase")
	assert.Equal(suite.T(), 0, len(allocations), "Should return empty list for non-existent course phase")
}

func (suite *AllocationServiceTestSuite) TestGetAllocationByCourseParticipationID() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	courseParticipationID := uuid.MustParse("99999999-9999-9999-9999-999999999991")
	expectedTeamID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	teamID, err := GetAllocationByCourseParticipationID(suite.suiteCtx, courseParticipationID, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTeamID, teamID, "Should return the correct team ID")
}

func (suite *AllocationServiceTestSuite) TestGetAllocationByCourseParticipationIDNotFound() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	nonExistentID := uuid.New()

	_, err := GetAllocationByCourseParticipationID(suite.suiteCtx, nonExistentID, coursePhaseID)
	assert.Error(suite.T(), err, "Should error for non-existent course participation")
}

func (suite *AllocationServiceTestSuite) TestGetAllocationByCourseParticipationIDWrongCoursePhase() {
	wrongCoursePhaseID := uuid.New()
	courseParticipationID := uuid.MustParse("99999999-9999-9999-9999-999999999991")

	_, err := GetAllocationByCourseParticipationID(suite.suiteCtx, courseParticipationID, wrongCoursePhaseID)
	assert.Error(suite.T(), err, "Should error for wrong course phase")
}

func (suite *AllocationServiceTestSuite) TestUpsertAllocationCreatesAndMoves() {
	phaseID := uuid.MustParse(writePhase)
	participantID := uuid.MustParse(staffFreeParticip)

	err := UpsertAllocation(suite.suiteCtx, "", phaseID, participantID, uuid.MustParse(teamDelta), pgtype.UUID{})
	assert.NoError(suite.T(), err)

	teamID, err := GetAllocationByCourseParticipationID(suite.suiteCtx, participantID, phaseID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uuid.MustParse(teamDelta), teamID)

	err = UpsertAllocation(suite.suiteCtx, "", phaseID, participantID, uuid.MustParse(teamEpsilon), pgtype.UUID{})
	assert.NoError(suite.T(), err)

	teamID, err = GetAllocationByCourseParticipationID(suite.suiteCtx, participantID, phaseID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uuid.MustParse(teamEpsilon), teamID)
}

func (suite *AllocationServiceTestSuite) TestUpsertAllocationScopedGuardBlocksForeignSourceTeam() {
	phaseID := uuid.MustParse(writePhase)
	guard := pgtype.UUID{Bytes: uuid.MustParse(teamDelta), Valid: true}

	err := UpsertAllocation(suite.suiteCtx, "", phaseID, uuid.MustParse(epsilonParticip), uuid.MustParse(teamDelta), guard)
	assert.ErrorIs(suite.T(), err, ErrTeamWriteDenied)

	teamID, err := GetAllocationByCourseParticipationID(suite.suiteCtx, uuid.MustParse(epsilonParticip), phaseID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uuid.MustParse(teamEpsilon), teamID)
}

func (suite *AllocationServiceTestSuite) TestUpsertAllocationScopedGuardAllowsUnallocatedAndOwnTeam() {
	phaseID := uuid.MustParse(writePhase)
	guard := pgtype.UUID{Bytes: uuid.MustParse(teamDelta), Valid: true}

	err := UpsertAllocation(suite.suiteCtx, "", phaseID, uuid.MustParse(tutorFreeParticip), uuid.MustParse(teamDelta), guard)
	assert.NoError(suite.T(), err, "an unallocated participant may be picked up")

	err = UpsertAllocation(suite.suiteCtx, "", phaseID, uuid.MustParse(tutorFreeParticip), uuid.MustParse(teamDelta), guard)
	assert.NoError(suite.T(), err, "a participant already in the scoped team may be rewritten")
}

func (suite *AllocationServiceTestSuite) TestUpsertAllocationRejectsUnknownTeamAndParticipant() {
	phaseID := uuid.MustParse(writePhase)

	err := UpsertAllocation(suite.suiteCtx, "", phaseID, uuid.MustParse(staffFreeParticip), uuid.New(), pgtype.UUID{})
	assert.ErrorIs(suite.T(), err, ErrInvalidTeamForPhase)

	err = UpsertAllocation(suite.suiteCtx, "", phaseID, uuid.New(), uuid.MustParse(teamDelta), pgtype.UUID{})
	assert.ErrorIs(suite.T(), err, ErrParticipantNotInPhase)
}

func (suite *AllocationServiceTestSuite) TestDeleteAllocationRemovesRowAndReportsMissing() {
	phaseID := uuid.MustParse(writePhase)

	err := DeleteAllocation(suite.suiteCtx, phaseID, uuid.MustParse(deltaParticip), pgtype.UUID{})
	assert.NoError(suite.T(), err)

	err = DeleteAllocation(suite.suiteCtx, phaseID, uuid.MustParse(deltaParticip), pgtype.UUID{})
	assert.ErrorIs(suite.T(), err, ErrAllocationNotFound)
}

func (suite *AllocationServiceTestSuite) TestDeleteAllocationScopedGuardBlocksForeignTeam() {
	phaseID := uuid.MustParse(writePhase)
	guard := pgtype.UUID{Bytes: uuid.MustParse(teamDelta), Valid: true}

	err := DeleteAllocation(suite.suiteCtx, phaseID, uuid.MustParse(epsilonParticip), guard)
	assert.ErrorIs(suite.T(), err, ErrAllocationNotFound)

	teamID, err := GetAllocationByCourseParticipationID(suite.suiteCtx, uuid.MustParse(epsilonParticip), phaseID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uuid.MustParse(teamEpsilon), teamID, "the foreign allocation must survive")
}

func TestAllocationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AllocationServiceTestSuite))
}
