package interviewAssignment

import (
	"context"
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
	activePhaseID     uuid.UUID
	participationUUID uuid.UUID
}

func (suite *InterviewAssignmentServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	require.NoError(suite.T(), err)
	suite.testDB = testDB
	suite.cleanup = cleanup
	suite.activePhaseID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	suite.participationUUID = uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")

	InterviewAssignmentServiceSingleton = &InterviewAssignmentService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
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

	assignment, err := CreateInterviewAssignment(suite.ctx, suite.activePhaseID, participationID, req)

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

	_, err := CreateInterviewAssignment(suite.ctx, suite.activePhaseID, participationID, req)

	require.Error(suite.T(), err)
	var serviceErr *ServiceError
	require.ErrorAs(suite.T(), err, &serviceErr)
	require.Equal(suite.T(), 409, serviceErr.StatusCode)
}

func (suite *InterviewAssignmentServiceTestSuite) TestCreateInterviewAssignmentSlotNotInPhase() {
	wrongPhaseSlotID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd") // Future phase
	participationID := uuid.New()

	req := interviewAssignmentDTO.CreateInterviewAssignmentRequest{
		InterviewSlotID: wrongPhaseSlotID,
	}

	_, err := CreateInterviewAssignment(suite.ctx, suite.activePhaseID, participationID, req)

	require.Error(suite.T(), err)
	var serviceErr *ServiceError
	require.ErrorAs(suite.T(), err, &serviceErr)
	require.Equal(suite.T(), 404, serviceErr.StatusCode)
}

func (suite *InterviewAssignmentServiceTestSuite) TestGetMyInterviewAssignment() {
	participationID := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")

	assignment, err := GetMyInterviewAssignment(suite.ctx, suite.activePhaseID, participationID)

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), participationID, assignment.CourseParticipationID)
	require.NotEqual(suite.T(), uuid.Nil, assignment.ID)
}

func (suite *InterviewAssignmentServiceTestSuite) TestGetMyInterviewAssignmentNotFound() {
	participationID := uuid.New()

	_, err := GetMyInterviewAssignment(suite.ctx, suite.activePhaseID, participationID)

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
	assignment, err := CreateInterviewAssignment(suite.ctx, suite.activePhaseID, participationID, createReq)
	require.NoError(suite.T(), err)

	err = DeleteInterviewAssignment(suite.ctx, suite.activePhaseID, assignment.ID, participationID, false)

	require.NoError(suite.T(), err)

	_, err = suite.testDB.Queries.GetInterviewAssignment(suite.ctx, assignment.ID)
	require.Error(suite.T(), err)
}

func (suite *InterviewAssignmentServiceTestSuite) TestDeleteInterviewAssignmentUnauthorized() {
	assignmentID := uuid.MustParse("11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	wrongParticipationID := uuid.New()

	err := DeleteInterviewAssignment(suite.ctx, suite.activePhaseID, assignmentID, wrongParticipationID, false)

	require.Error(suite.T(), err)
	var serviceErr *ServiceError
	require.ErrorAs(suite.T(), err, &serviceErr)
	require.Equal(suite.T(), 403, serviceErr.StatusCode)
}

func TestInterviewAssignmentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(InterviewAssignmentServiceTestSuite))
}
