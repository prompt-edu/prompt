package interviewSlot

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/google/uuid"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
	interviewSlotDTO "github.com/prompt-edu/prompt/servers/interview/interviewSlot/interviewSlotDTO"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type InterviewSlotServiceTestSuite struct {
	suite.Suite
	ctx           context.Context
	testDB        *sdkTestUtils.TestDB[*db.Queries]
	cleanup       func()
	service       *InterviewSlotService
	activePhaseID uuid.UUID
	futurePhaseID uuid.UUID
}

func (suite *InterviewSlotServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	require.NoError(suite.T(), err)
	suite.testDB = testDB
	suite.cleanup = cleanup
	suite.activePhaseID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	suite.futurePhaseID = uuid.MustParse("22222222-2222-2222-2222-222222222222")

	suite.service = NewInterviewSlotService(*testDB.Queries, testDB.Conn)
}

func (suite *InterviewSlotServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *InterviewSlotServiceTestSuite) TestCreateInterviewSlot() {
	location := "Test Room"
	req := interviewSlotDTO.CreateInterviewSlotRequest{
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(25 * time.Hour),
		Location:  &location,
		Capacity:  2,
	}

	slot, err := suite.service.CreateInterviewSlot(suite.ctx, suite.activePhaseID, req)

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), suite.activePhaseID, slot.CoursePhaseID)
	require.Equal(suite.T(), int32(2), slot.Capacity)

	// Verify it was saved to database
	dbSlot, err := suite.testDB.Queries.GetInterviewSlot(suite.ctx, slot.ID)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), slot.ID, dbSlot.ID)
}

func (suite *InterviewSlotServiceTestSuite) TestCreateInterviewSlotInvalidTimes() {
	req := interviewSlotDTO.CreateInterviewSlotRequest{
		StartTime: time.Now().Add(25 * time.Hour),
		EndTime:   time.Now().Add(24 * time.Hour), // End before start
		Capacity:  1,
	}

	_, err := suite.service.CreateInterviewSlot(suite.ctx, suite.activePhaseID, req)

	require.Error(suite.T(), err)
	var serviceErr *ServiceError
	require.ErrorAs(suite.T(), err, &serviceErr)
	require.Equal(suite.T(), 400, serviceErr.StatusCode)
}

func (suite *InterviewSlotServiceTestSuite) TestCreateInterviewSlotsBatch() {
	location := "Batch Room"
	start := time.Now().Add(48 * time.Hour)
	req := interviewSlotDTO.CreateInterviewSlotsBatchRequest{
		Slots: []interviewSlotDTO.CreateInterviewSlotRequest{
			{StartTime: start, EndTime: start.Add(30 * time.Minute), Location: &location, Capacity: 1},
			{StartTime: start.Add(30 * time.Minute), EndTime: start.Add(60 * time.Minute), Location: &location, Capacity: 1},
			{StartTime: start.Add(60 * time.Minute), EndTime: start.Add(90 * time.Minute), Location: &location, Capacity: 1},
		},
	}

	slots, err := suite.service.CreateInterviewSlotsBatch(suite.ctx, suite.futurePhaseID, req)

	require.NoError(suite.T(), err)
	require.Len(suite.T(), slots, 3)
	for _, slot := range slots {
		require.Equal(suite.T(), suite.futurePhaseID, slot.CoursePhaseID)
		dbSlot, err := suite.testDB.Queries.GetInterviewSlot(suite.ctx, slot.ID)
		require.NoError(suite.T(), err)
		require.Equal(suite.T(), slot.ID, dbSlot.ID)
	}
}

func (suite *InterviewSlotServiceTestSuite) TestCreateInterviewSlotsBatchRejectsInvalidTimesWithoutCreating() {
	before, err := suite.service.GetAllInterviewSlots(suite.ctx, suite.activePhaseID, "")
	require.NoError(suite.T(), err)

	start := time.Now().Add(72 * time.Hour)
	req := interviewSlotDTO.CreateInterviewSlotsBatchRequest{
		Slots: []interviewSlotDTO.CreateInterviewSlotRequest{
			{StartTime: start, EndTime: start.Add(30 * time.Minute), Capacity: 1},
			{StartTime: start.Add(60 * time.Minute), EndTime: start.Add(30 * time.Minute), Capacity: 1}, // end before start
		},
	}

	_, err = suite.service.CreateInterviewSlotsBatch(suite.ctx, suite.activePhaseID, req)

	require.Error(suite.T(), err)
	var serviceErr *ServiceError
	require.ErrorAs(suite.T(), err, &serviceErr)
	require.Equal(suite.T(), 400, serviceErr.StatusCode)

	// The valid slot preceding the rejected one must not have been created either.
	after, err := suite.service.GetAllInterviewSlots(suite.ctx, suite.activePhaseID, "")
	require.NoError(suite.T(), err)
	require.Len(suite.T(), after, len(before))
}

func (suite *InterviewSlotServiceTestSuite) TestCreateInterviewSlotsBatchRejectsTooManySlots() {
	start := time.Now().Add(96 * time.Hour)
	slots := make([]interviewSlotDTO.CreateInterviewSlotRequest, interviewSlotDTO.MaxBatchInterviewSlots+1)
	for i := range slots {
		slotStart := start.Add(time.Duration(i) * time.Hour)
		slots[i] = interviewSlotDTO.CreateInterviewSlotRequest{
			StartTime: slotStart,
			EndTime:   slotStart.Add(30 * time.Minute),
			Capacity:  1,
		}
	}

	_, err := suite.service.CreateInterviewSlotsBatch(suite.ctx, suite.activePhaseID, interviewSlotDTO.CreateInterviewSlotsBatchRequest{Slots: slots})

	require.Error(suite.T(), err)
	var serviceErr *ServiceError
	require.ErrorAs(suite.T(), err, &serviceErr)
	require.Equal(suite.T(), 400, serviceErr.StatusCode)
}

func (suite *InterviewSlotServiceTestSuite) TestGetAllInterviewSlots() {
	slots, err := suite.service.GetAllInterviewSlots(suite.ctx, suite.activePhaseID, "")

	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), len(slots), 3)

	foundSlotWithAssignment := false
	for _, slot := range slots {
		if len(slot.Assignments) > 0 {
			foundSlotWithAssignment = true
			require.Equal(suite.T(), int64(len(slot.Assignments)), slot.AssignedCount)
			break
		}
	}
	require.True(suite.T(), foundSlotWithAssignment)
}

func (suite *InterviewSlotServiceTestSuite) TestGetInterviewSlot() {
	slotID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	slot, err := suite.service.GetInterviewSlot(suite.ctx, suite.activePhaseID, slotID, "")

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), slotID, slot.ID)
	require.Equal(suite.T(), suite.activePhaseID, slot.CoursePhaseID)
	require.GreaterOrEqual(suite.T(), len(slot.Assignments), 0)
}

func (suite *InterviewSlotServiceTestSuite) TestGetInterviewSlotNotFound() {
	nonExistentID := uuid.New()

	_, err := suite.service.GetInterviewSlot(suite.ctx, suite.activePhaseID, nonExistentID, "")

	require.Error(suite.T(), err)
	var serviceErr *ServiceError
	require.ErrorAs(suite.T(), err, &serviceErr)
	require.Equal(suite.T(), 404, serviceErr.StatusCode)
}

func (suite *InterviewSlotServiceTestSuite) TestUpdateInterviewSlot() {
	location := "Original Room"
	createReq := interviewSlotDTO.CreateInterviewSlotRequest{
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(25 * time.Hour),
		Location:  &location,
		Capacity:  2,
	}
	slot, err := suite.service.CreateInterviewSlot(suite.ctx, suite.activePhaseID, createReq)
	require.NoError(suite.T(), err)

	newLocation := "Updated Room"
	updateReq := interviewSlotDTO.UpdateInterviewSlotRequest{
		StartTime: time.Now().Add(26 * time.Hour),
		EndTime:   time.Now().Add(27 * time.Hour),
		Location:  &newLocation,
		Capacity:  3,
	}

	updated, err := suite.service.UpdateInterviewSlot(suite.ctx, suite.activePhaseID, slot.ID, updateReq)

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), slot.ID, updated.ID)
	require.Equal(suite.T(), int32(3), updated.Capacity)
}

func (suite *InterviewSlotServiceTestSuite) TestUpdateInterviewSlotReduceCapacityBelowAssignments() {
	// Create a slot with capacity 2
	location := "Test"
	createReq := interviewSlotDTO.CreateInterviewSlotRequest{
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(25 * time.Hour),
		Location:  &location,
		Capacity:  2,
	}
	slot, err := suite.service.CreateInterviewSlot(suite.ctx, suite.activePhaseID, createReq)
	require.NoError(suite.T(), err)

	// Add 2 assignments
	_, err = suite.testDB.Queries.CreateInterviewAssignment(suite.ctx, db.CreateInterviewAssignmentParams{
		InterviewSlotID:       slot.ID,
		CourseParticipationID: uuid.New(),
	})
	require.NoError(suite.T(), err)
	_, err = suite.testDB.Queries.CreateInterviewAssignment(suite.ctx, db.CreateInterviewAssignmentParams{
		InterviewSlotID:       slot.ID,
		CourseParticipationID: uuid.New(),
	})
	require.NoError(suite.T(), err)

	// Try to reduce capacity below 2
	updateReq := interviewSlotDTO.UpdateInterviewSlotRequest{
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(25 * time.Hour),
		Location:  &location,
		Capacity:  1,
	}

	_, err = suite.service.UpdateInterviewSlot(suite.ctx, suite.activePhaseID, slot.ID, updateReq)

	require.Error(suite.T(), err)
	var serviceErr *ServiceError
	require.ErrorAs(suite.T(), err, &serviceErr)
	require.Equal(suite.T(), 400, serviceErr.StatusCode)
}

func (suite *InterviewSlotServiceTestSuite) TestDeleteInterviewSlot() {
	location := "Disposable Room"
	createReq := interviewSlotDTO.CreateInterviewSlotRequest{
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(25 * time.Hour),
		Location:  &location,
		Capacity:  1,
	}
	slot, err := suite.service.CreateInterviewSlot(suite.ctx, suite.activePhaseID, createReq)
	require.NoError(suite.T(), err)

	err = suite.service.DeleteInterviewSlot(suite.ctx, suite.activePhaseID, slot.ID)

	require.NoError(suite.T(), err)

	_, err = suite.testDB.Queries.GetInterviewSlot(suite.ctx, slot.ID)
	require.Error(suite.T(), err)
}

func (suite *InterviewSlotServiceTestSuite) TestDeleteInterviewSlotCascadesAssignments() {
	// Create a new slot with assignments for testing cascade delete
	location := "Cascade Test Room"
	createReq := interviewSlotDTO.CreateInterviewSlotRequest{
		StartTime: time.Now().Add(48 * time.Hour),
		EndTime:   time.Now().Add(49 * time.Hour),
		Location:  &location,
		Capacity:  5,
	}
	slot, err := suite.service.CreateInterviewSlot(suite.ctx, suite.activePhaseID, createReq)
	require.NoError(suite.T(), err)

	// Create some assignments for this slot
	participationID1 := uuid.MustParse("aaaa1111-1111-1111-1111-111111111111")
	participationID2 := uuid.MustParse("bbbb1111-1111-1111-1111-111111111111")

	_, err = suite.testDB.Queries.CreateInterviewAssignment(suite.ctx, db.CreateInterviewAssignmentParams{
		InterviewSlotID:       slot.ID,
		CourseParticipationID: participationID1,
	})
	require.NoError(suite.T(), err)

	_, err = suite.testDB.Queries.CreateInterviewAssignment(suite.ctx, db.CreateInterviewAssignmentParams{
		InterviewSlotID:       slot.ID,
		CourseParticipationID: participationID2,
	})
	require.NoError(suite.T(), err)

	// Verify assignments exist
	assignments, err := suite.testDB.Queries.GetInterviewAssignmentsBySlot(suite.ctx, slot.ID)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 2, len(assignments))

	// Delete the slot
	err = suite.service.DeleteInterviewSlot(suite.ctx, suite.activePhaseID, slot.ID)
	require.NoError(suite.T(), err)

	// Verify assignments were cascaded (deleted)
	assignments, err = suite.testDB.Queries.GetInterviewAssignmentsBySlot(suite.ctx, slot.ID)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 0, len(assignments))
}

func TestInterviewSlotServiceTestSuite(t *testing.T) {
	suite.Run(t, new(InterviewSlotServiceTestSuite))
}
