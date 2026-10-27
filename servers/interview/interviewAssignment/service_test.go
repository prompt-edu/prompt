package interviewAssignment

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
	interviewAssignmentDTO "github.com/prompt-edu/prompt/servers/interview/interviewAssignment/interviewAssignmentDTO"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type InterviewAssignmentServiceTestSuite struct {
	suite.Suite
	ctx               context.Context
	testDB            *sdkTestUtils.TestDB[*db.Queries]
	cleanup           func()
	service           *InterviewAssignmentService
	activePhaseID     uuid.UUID
	participationUUID uuid.UUID
}

func (suite *InterviewAssignmentServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.ctx, "../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../database_dumps/base.sql")
	require.NoError(suite.T(), err)
	suite.testDB = testDB
	suite.cleanup = cleanup
	suite.activePhaseID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	suite.participationUUID = uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")

	suite.service = NewInterviewAssignmentService(*testDB.Queries, testDB.Conn)
}

func (suite *InterviewAssignmentServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *InterviewAssignmentServiceTestSuite) TestCreateInterviewAssignment() {
	slotID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	participationID := uuid.New()
	req := interviewAssignmentDTO.CreateInterviewAssignmentRequest{
		InterviewSlotID: slotID,
	}

	assignment, err := suite.service.CreateInterviewAssignment(suite.ctx, suite.activePhaseID, participationID, req)

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), slotID, assignment.InterviewSlotID)
	require.Equal(suite.T(), participationID, assignment.CourseParticipationID)

	// Verify it was saved to database
	dbAssignment, err := suite.testDB.Queries.GetInterviewAssignment(suite.ctx, assignment.ID)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), assignment.ID, dbAssignment.ID)
}

func (suite *InterviewAssignmentServiceTestSuite) TestCreateInterviewAssignmentExceedsCapacity() {
	slotID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc") // Capacity 1, already has 1
	participationID := uuid.New()

	req := interviewAssignmentDTO.CreateInterviewAssignmentRequest{
		InterviewSlotID: slotID,
	}

	_, err := suite.service.CreateInterviewAssignment(suite.ctx, suite.activePhaseID, participationID, req)

	require.Error(suite.T(), err)
	var serviceErr *ServiceError
	require.ErrorAs(suite.T(), err, &serviceErr)
	require.Equal(suite.T(), 409, serviceErr.StatusCode)
}

// Regression test for GHSA-w97v-vx5j-3gjj: parallel bookings of different slots by the
// same participation must leave exactly one assignment in the phase.
func (suite *InterviewAssignmentServiceTestSuite) TestCreateInterviewAssignmentConcurrentBookings() {
	const slots = 8
	slotIDs := make([]uuid.UUID, 0, slots)
	for i := range slots {
		slot, err := suite.testDB.Queries.CreateInterviewSlot(suite.ctx, db.CreateInterviewSlotParams{
			CoursePhaseID: suite.activePhaseID,
			StartTime:     pgtype.Timestamptz{Time: time.Now().Add(time.Duration(i+1) * time.Hour), Valid: true},
			EndTime:       pgtype.Timestamptz{Time: time.Now().Add(time.Duration(i+2) * time.Hour), Valid: true},
			Location:      pgtype.Text{String: "Race Room", Valid: true},
			Capacity:      5,
		})
		require.NoError(suite.T(), err)
		slotIDs = append(slotIDs, slot.ID)
	}

	participationID := uuid.New()
	var wg sync.WaitGroup
	errs := make(chan error, slots)
	for _, slotID := range slotIDs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := suite.service.CreateInterviewAssignment(suite.ctx, suite.activePhaseID, participationID, interviewAssignmentDTO.CreateInterviewAssignmentRequest{
				InterviewSlotID: slotID,
			})
			errs <- err
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
		var serviceErr *ServiceError
		require.ErrorAs(suite.T(), err, &serviceErr)
		require.Equal(suite.T(), 409, serviceErr.StatusCode)
	}
	require.Equal(suite.T(), 1, succeeded)

	booked := 0
	for _, slotID := range slotIDs {
		assignments, err := suite.testDB.Queries.GetInterviewAssignmentsBySlot(suite.ctx, slotID)
		require.NoError(suite.T(), err)
		booked += len(assignments)
	}
	require.Equal(suite.T(), 1, booked)
}

// A booking that races an uncommitted booking of the same participation for another slot
// must fail with 409 once that booking commits, instead of creating a second assignment.
func (suite *InterviewAssignmentServiceTestSuite) TestCreateInterviewAssignmentWaitsForRacingBooking() {
	newSlot := func(offset time.Duration) uuid.UUID {
		slot, err := suite.testDB.Queries.CreateInterviewSlot(suite.ctx, db.CreateInterviewSlotParams{
			CoursePhaseID: suite.activePhaseID,
			StartTime:     pgtype.Timestamptz{Time: time.Now().Add(offset), Valid: true},
			EndTime:       pgtype.Timestamptz{Time: time.Now().Add(offset + time.Hour), Valid: true},
			Location:      pgtype.Text{String: "Race Room", Valid: true},
			Capacity:      5,
		})
		require.NoError(suite.T(), err)
		return slot.ID
	}
	firstSlot, secondSlot := newSlot(20*time.Hour), newSlot(22*time.Hour)
	participationID := uuid.New()

	// Play the first booking by hand and keep it uncommitted.
	tx, err := suite.testDB.Conn.Begin(suite.ctx)
	require.NoError(suite.T(), err)
	defer func() { _ = tx.Rollback(suite.ctx) }()
	_, err = suite.testDB.Queries.WithTx(tx).CreateInterviewAssignment(suite.ctx, db.CreateInterviewAssignmentParams{
		InterviewSlotID:       firstSlot,
		CourseParticipationID: participationID,
	})
	require.NoError(suite.T(), err)

	done := make(chan error, 1)
	go func() {
		_, err := suite.service.CreateInterviewAssignment(suite.ctx, suite.activePhaseID, participationID, interviewAssignmentDTO.CreateInterviewAssignmentRequest{
			InterviewSlotID: secondSlot,
		})
		done <- err
	}()

	select {
	case err := <-done:
		suite.FailNow("booking did not wait for the racing booking", "returned %v", err)
	case <-time.After(300 * time.Millisecond):
	}

	require.NoError(suite.T(), tx.Commit(suite.ctx))
	err = <-done
	var serviceErr *ServiceError
	require.ErrorAs(suite.T(), err, &serviceErr)
	require.Equal(suite.T(), 409, serviceErr.StatusCode)

	assignments, err := suite.testDB.Queries.GetInterviewAssignmentsBySlot(suite.ctx, secondSlot)
	require.NoError(suite.T(), err)
	require.Empty(suite.T(), assignments)
}

func (suite *InterviewAssignmentServiceTestSuite) TestCreateInterviewAssignmentSlotNotInPhase() {
	wrongPhaseSlotID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd") // Future phase
	participationID := uuid.New()

	req := interviewAssignmentDTO.CreateInterviewAssignmentRequest{
		InterviewSlotID: wrongPhaseSlotID,
	}

	_, err := suite.service.CreateInterviewAssignment(suite.ctx, suite.activePhaseID, participationID, req)

	require.Error(suite.T(), err)
	var serviceErr *ServiceError
	require.ErrorAs(suite.T(), err, &serviceErr)
	require.Equal(suite.T(), 404, serviceErr.StatusCode)
}

func (suite *InterviewAssignmentServiceTestSuite) TestGetMyInterviewAssignment() {
	participationID := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")

	assignment, err := suite.service.GetMyInterviewAssignment(suite.ctx, suite.activePhaseID, participationID)

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), participationID, assignment.CourseParticipationID)
	require.NotEqual(suite.T(), uuid.Nil, assignment.ID)
}

func (suite *InterviewAssignmentServiceTestSuite) TestGetMyInterviewAssignmentNotFound() {
	participationID := uuid.New()

	_, err := suite.service.GetMyInterviewAssignment(suite.ctx, suite.activePhaseID, participationID)

	require.Error(suite.T(), err)
	var serviceErr *ServiceError
	require.ErrorAs(suite.T(), err, &serviceErr)
	require.Equal(suite.T(), 404, serviceErr.StatusCode)
}

func (suite *InterviewAssignmentServiceTestSuite) TestDeleteInterviewAssignment() {
	// Create a dedicated slot for this test with large capacity
	slot, err := suite.testDB.Queries.CreateInterviewSlot(suite.ctx, db.CreateInterviewSlotParams{
		CoursePhaseID: suite.activePhaseID,
		StartTime:     pgtype.Timestamptz{Time: time.Now().Add(100 * time.Hour), Valid: true},
		EndTime:       pgtype.Timestamptz{Time: time.Now().Add(101 * time.Hour), Valid: true},
		Capacity:      10,
	})
	require.NoError(suite.T(), err)

	participationID := uuid.New()
	createReq := interviewAssignmentDTO.CreateInterviewAssignmentRequest{
		InterviewSlotID: slot.ID,
	}
	assignment, err := suite.service.CreateInterviewAssignment(suite.ctx, suite.activePhaseID, participationID, createReq)
	require.NoError(suite.T(), err)

	err = suite.service.DeleteInterviewAssignment(suite.ctx, suite.activePhaseID, assignment.ID, participationID, false)

	require.NoError(suite.T(), err)

	_, err = suite.testDB.Queries.GetInterviewAssignment(suite.ctx, assignment.ID)
	require.Error(suite.T(), err)
}

func (suite *InterviewAssignmentServiceTestSuite) TestDeleteInterviewAssignmentUnauthorized() {
	assignmentID := uuid.MustParse("11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	wrongParticipationID := uuid.New()

	err := suite.service.DeleteInterviewAssignment(suite.ctx, suite.activePhaseID, assignmentID, wrongParticipationID, false)

	require.Error(suite.T(), err)
	var serviceErr *ServiceError
	require.ErrorAs(suite.T(), err, &serviceErr)
	require.Equal(suite.T(), 403, serviceErr.StatusCode)
}

func TestInterviewAssignmentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(InterviewAssignmentServiceTestSuite))
}
