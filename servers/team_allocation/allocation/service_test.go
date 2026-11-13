package allocation

import (
	"context"
	"testing"

	"github.com/google/uuid"
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
	allocationService *AllocationService
}

func (suite *AllocationServiceTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.suiteCtx, "../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../database_dumps/allocations.sql")
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.allocationService = NewAllocationService(*testDB.Queries, testDB.Conn)
	suite.allocationService.resolveParticipants = stubParticipants(tutorFreeParticip, deltaParticip, epsilonParticip, staffFreeParticip)
}

func (suite *AllocationServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *AllocationServiceTestSuite) TestGetAllAllocations() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")

	allocations, err := suite.allocationService.GetAllAllocations(suite.suiteCtx, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(allocations), 0, "Expected at least one allocation")

	for _, allocation := range allocations {
		assert.NotEmpty(suite.T(), allocation.CourseParticipationID, "CourseParticipationID should not be empty")
		assert.NotEmpty(suite.T(), allocation.TeamAllocation, "TeamAllocation should not be empty")
	}
}

func (suite *AllocationServiceTestSuite) TestGetAllAllocationsNonExistentCoursePhase() {
	nonExistentID := uuid.New()

	allocations, err := suite.allocationService.GetAllAllocations(suite.suiteCtx, nonExistentID)
	assert.NoError(suite.T(), err, "Should not error for non-existent course phase")
	assert.Equal(suite.T(), 0, len(allocations), "Should return empty list for non-existent course phase")
}

func (suite *AllocationServiceTestSuite) TestGetAllocationByCourseParticipationID() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	courseParticipationID := uuid.MustParse("99999999-9999-9999-9999-999999999991")
	expectedTeamID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	teamID, err := suite.allocationService.GetAllocationByCourseParticipationID(suite.suiteCtx, courseParticipationID, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTeamID, teamID, "Should return the correct team ID")
}

func (suite *AllocationServiceTestSuite) TestGetAllocationByCourseParticipationIDNotFound() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	nonExistentID := uuid.New()

	_, err := suite.allocationService.GetAllocationByCourseParticipationID(suite.suiteCtx, nonExistentID, coursePhaseID)
	assert.Error(suite.T(), err, "Should error for non-existent course participation")
}

func (suite *AllocationServiceTestSuite) TestGetAllocationByCourseParticipationIDWrongCoursePhase() {
	wrongCoursePhaseID := uuid.New()
	courseParticipationID := uuid.MustParse("99999999-9999-9999-9999-999999999991")

	_, err := suite.allocationService.GetAllocationByCourseParticipationID(suite.suiteCtx, courseParticipationID, wrongCoursePhaseID)
	assert.Error(suite.T(), err, "Should error for wrong course phase")
}

func (suite *AllocationServiceTestSuite) TestUpsertAllocationCreatesAndMoves() {
	phaseID := uuid.MustParse(writePhase)
	participantID := uuid.MustParse(staffFreeParticip)

	err := suite.allocationService.UpsertAllocation(suite.suiteCtx, "", phaseID, participantID, uuid.MustParse(teamDelta))
	assert.NoError(suite.T(), err)

	teamID, err := suite.allocationService.GetAllocationByCourseParticipationID(suite.suiteCtx, participantID, phaseID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uuid.MustParse(teamDelta), teamID)

	err = suite.allocationService.UpsertAllocation(suite.suiteCtx, "", phaseID, participantID, uuid.MustParse(teamEpsilon))
	assert.NoError(suite.T(), err)

	teamID, err = suite.allocationService.GetAllocationByCourseParticipationID(suite.suiteCtx, participantID, phaseID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uuid.MustParse(teamEpsilon), teamID)
}

func (suite *AllocationServiceTestSuite) TestUpsertAllocationRejectsUnknownTeamAndParticipant() {
	phaseID := uuid.MustParse(writePhase)

	err := suite.allocationService.UpsertAllocation(suite.suiteCtx, "", phaseID, uuid.MustParse(staffFreeParticip), uuid.New())
	assert.ErrorIs(suite.T(), err, ErrInvalidTeamForPhase)

	err = suite.allocationService.UpsertAllocation(suite.suiteCtx, "", phaseID, uuid.New(), uuid.MustParse(teamDelta))
	assert.ErrorIs(suite.T(), err, ErrParticipantNotInPhase)
}

func (suite *AllocationServiceTestSuite) TestDeleteAllocationRemovesRowAndReportsMissing() {
	phaseID := uuid.MustParse(writePhase)

	err := suite.allocationService.DeleteAllocation(suite.suiteCtx, phaseID, uuid.MustParse(deltaParticip))
	assert.NoError(suite.T(), err)

	err = suite.allocationService.DeleteAllocation(suite.suiteCtx, phaseID, uuid.MustParse(deltaParticip))
	assert.ErrorIs(suite.T(), err, ErrAllocationNotFound)
}

func TestAllocationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AllocationServiceTestSuite))
}
