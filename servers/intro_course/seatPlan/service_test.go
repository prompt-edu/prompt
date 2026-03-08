package seatPlan

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prompt-edu/prompt/servers/intro_course/seatPlan/seatPlanDTO"
	"github.com/prompt-edu/prompt/servers/intro_course/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SeatPlanServiceTestSuite struct {
	suite.Suite
	ctx                 context.Context
	cleanup             func()
	coursePhaseID       uuid.UUID
	studentID           uuid.UUID
	tutorID             uuid.UUID
	seatPlanTestService SeatPlanService
}

func (suite *SeatPlanServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(suite.ctx, "../database_dumps/intro_course.sql")
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}

	suite.cleanup = cleanup
	suite.coursePhaseID = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	suite.studentID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	suite.tutorID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	suite.seatPlanTestService = SeatPlanService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	SeatPlanServiceSingleton = &suite.seatPlanTestService
}

func (suite *SeatPlanServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func TestSeatPlanServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SeatPlanServiceTestSuite))
}

func (suite *SeatPlanServiceTestSuite) TestValidateSeatNames() {
	assert.True(suite.T(), validateSeatNames([]string{"A", "B", "C"}))
	assert.False(suite.T(), validateSeatNames([]string{"A", "B", "A"}))
}

func (suite *SeatPlanServiceTestSuite) TestGetSeatPlan() {
	seats, err := GetSeatPlan(suite.ctx, suite.coursePhaseID)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), seats, 2)
	assert.Equal(suite.T(), "Seat-1", seats[0].SeatName)
}

func (suite *SeatPlanServiceTestSuite) TestCreateSeatPlan() {
	newCoursePhaseID := uuid.New()
	seatNames := []string{"A-1", "A-2", "A-3"}

	err := CreateSeatPlan(suite.ctx, newCoursePhaseID, seatNames)
	assert.NoError(suite.T(), err)

	createdSeats, err := GetSeatPlan(suite.ctx, newCoursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), createdSeats, len(seatNames))
}

func (suite *SeatPlanServiceTestSuite) TestUpdateSeatPlan() {
	updateCoursePhaseID := uuid.New()
	err := CreateSeatPlan(suite.ctx, updateCoursePhaseID, []string{"B-1"})
	assert.NoError(suite.T(), err)

	updatedSeat := seatPlanDTO.Seat{
		SeatName:        "B-1",
		HasMac:          true,
		DeviceID:        pgtype.Text{String: "Device-42", Valid: true},
		AssignedStudent: pgtype.UUID{Bytes: uuid.New(), Valid: true},
		AssignedTutor:   pgtype.UUID{}, // keep NULL to avoid FK issues
	}

	err = UpdateSeatPlan(suite.ctx, updateCoursePhaseID, []seatPlanDTO.Seat{updatedSeat})
	assert.NoError(suite.T(), err)

	seats, err := GetSeatPlan(suite.ctx, updateCoursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), seats, 1)
	assert.True(suite.T(), seats[0].HasMac)
	assert.Equal(suite.T(), updatedSeat.DeviceID, seats[0].DeviceID)
	assert.Equal(suite.T(), updatedSeat.AssignedStudent, seats[0].AssignedStudent)
}

func (suite *SeatPlanServiceTestSuite) TestDeleteSeatPlan() {
	deleteCoursePhaseID := uuid.New()
	err := CreateSeatPlan(suite.ctx, deleteCoursePhaseID, []string{"C-1", "C-2"})
	assert.NoError(suite.T(), err)

	err = DeleteSeatPlan(suite.ctx, deleteCoursePhaseID)
	assert.NoError(suite.T(), err)

	seats, err := GetSeatPlan(suite.ctx, deleteCoursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), seats)
}

func (suite *SeatPlanServiceTestSuite) TestGetOwnSeatAssignment() {
	assignment, err := GetOwnSeatAssignment(suite.ctx, suite.coursePhaseID, suite.studentID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Seat-1", assignment.SeatName)
	assert.Equal(suite.T(), "Alice", assignment.TutorFirstName)
	assert.Equal(suite.T(), "Tutor", assignment.TutorLastName)
}

func (suite *SeatPlanServiceTestSuite) TestGetOwnSeatAssignmentNotFound() {
	assignment, err := GetOwnSeatAssignment(suite.ctx, suite.coursePhaseID, uuid.New())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), seatPlanDTO.SeatAssignment{}, assignment)
}
