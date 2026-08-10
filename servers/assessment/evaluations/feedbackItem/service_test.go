package feedbackItem

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/feedbackItem/feedbackItemDTO"
)

type FeedbackItemServiceTestSuite struct {
	suite.Suite
	suiteCtx                  context.Context
	cleanup                   func()
	feedbackItemService       FeedbackItemService
	testCoursePhaseID         uuid.UUID
	testCourseParticipationID uuid.UUID
	testAuthorID              uuid.UUID
	testFeedbackItemID        uuid.UUID
}

func (suite *FeedbackItemServiceTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../../database_dumps/feedbackItems.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.feedbackItemService = FeedbackItemService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	FeedbackItemServiceSingleton = &suite.feedbackItemService

	// Use predefined test UUIDs from the test data that match feedbackItems.sql
	suite.testCoursePhaseID = uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	suite.testCourseParticipationID = uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")
	suite.testAuthorID = uuid.MustParse("da42e447-60f9-4fe0-b297-2dae3f924fd7")
	suite.testFeedbackItemID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
}

func (suite *FeedbackItemServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

// seedFeedbackItem inserts a self feedback item authored by the given participation
// so tests that mutate rows do not depend on each other's ordering.
func (suite *FeedbackItemServiceTestSuite) seedFeedbackItem(authorID uuid.UUID) uuid.UUID {
	feedbackItemID := uuid.New()

	err := suite.feedbackItemService.queries.CreateFeedbackItem(suite.suiteCtx, db.CreateFeedbackItemParams{
		ID:                          feedbackItemID,
		FeedbackType:                db.FeedbackTypePositive,
		FeedbackText:                "Original feedback text",
		CourseParticipationID:       authorID,
		CoursePhaseID:               suite.testCoursePhaseID,
		AuthorCourseParticipationID: authorID,
		Type:                        db.AssessmentTypeSelf,
	})
	assert.NoError(suite.T(), err)
	return feedbackItemID
}

func (suite *FeedbackItemServiceTestSuite) TestCreateFeedbackItem() {
	req := feedbackItemDTO.CreateFeedbackItemRequest{
		FeedbackType:                db.FeedbackTypePositive,
		FeedbackText:                "Great work on this task!",
		CourseParticipationID:       suite.testAuthorID, // self feedback: subject is the author
		CoursePhaseID:               suite.testCoursePhaseID,
		AuthorCourseParticipationID: suite.testAuthorID,
		Type:                        assessmentType.Self,
	}

	err := CreateFeedbackItem(suite.suiteCtx, req)
	assert.NoError(suite.T(), err)
}

func (suite *FeedbackItemServiceTestSuite) TestCreateFeedbackItemRejectsForeignSelfTarget() {
	req := feedbackItemDTO.CreateFeedbackItemRequest{
		FeedbackType:                db.FeedbackTypePositive,
		FeedbackText:                "Injected into a peer's self feedback",
		CourseParticipationID:       suite.testCourseParticipationID, // not the author
		CoursePhaseID:               suite.testCoursePhaseID,
		AuthorCourseParticipationID: suite.testAuthorID,
		Type:                        assessmentType.Self,
	}

	err := CreateFeedbackItem(suite.suiteCtx, req)
	assert.ErrorIs(suite.T(), err, evaluationCompletion.ErrSelfEvaluationTargetMismatch)
}

func (suite *FeedbackItemServiceTestSuite) TestUpdateFeedbackItem() {
	updateFeedbackItemID := suite.seedFeedbackItem(suite.testAuthorID)
	req := feedbackItemDTO.UpdateFeedbackItemRequest{
		FeedbackType: db.FeedbackTypeNegative, // Change from positive to negative
		FeedbackText: "Updated feedback text",
	}

	err := UpdateFeedbackItem(suite.suiteCtx, updateFeedbackItemID, suite.testCoursePhaseID, suite.testAuthorID, req)
	assert.NoError(suite.T(), err)

	updated, err := GetFeedbackItem(suite.suiteCtx, updateFeedbackItemID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated feedback text", updated.FeedbackText)
	assert.Equal(suite.T(), db.FeedbackTypeNegative, updated.FeedbackType)
}

func (suite *FeedbackItemServiceTestSuite) TestUpdateFeedbackItemRejectsWrongPhase() {
	updateFeedbackItemID := suite.seedFeedbackItem(suite.testAuthorID)
	otherPhaseID := uuid.MustParse("34561b6b-3c3a-4bc6-ba42-69eeb1514da9")
	req := feedbackItemDTO.UpdateFeedbackItemRequest{
		FeedbackType: db.FeedbackTypeNegative,
		FeedbackText: "Updated from another phase",
	}

	err := UpdateFeedbackItem(suite.suiteCtx, updateFeedbackItemID, otherPhaseID, suite.testAuthorID, req)
	assert.ErrorIs(suite.T(), err, ErrFeedbackItemNotFound)
}

func (suite *FeedbackItemServiceTestSuite) TestUpdateFeedbackItemRejectsNonAuthor() {
	// Feedback item 22222222 is authored by the lecturer, not testAuthorID
	updateFeedbackItemID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	req := feedbackItemDTO.UpdateFeedbackItemRequest{
		FeedbackType: db.FeedbackTypePositive,
		FeedbackText: "Rewritten by a non-author",
	}

	err := UpdateFeedbackItem(suite.suiteCtx, updateFeedbackItemID, suite.testCoursePhaseID, suite.testAuthorID, req)
	assert.ErrorIs(suite.T(), err, ErrNotFeedbackItemAuthor)

	unchanged, err := GetFeedbackItem(suite.suiteCtx, updateFeedbackItemID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Need to improve time management", unchanged.FeedbackText)
}

func (suite *FeedbackItemServiceTestSuite) TestGetFeedbackItem() {
	feedbackItem, err := GetFeedbackItem(suite.suiteCtx, suite.testFeedbackItemID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.testFeedbackItemID, feedbackItem.ID)
	assert.Equal(suite.T(), "Great teamwork and communication skills!", feedbackItem.FeedbackText)
	assert.Equal(suite.T(), db.FeedbackTypePositive, feedbackItem.FeedbackType)
}

func (suite *FeedbackItemServiceTestSuite) TestDeleteFeedbackItem() {
	// First create a feedback item to delete
	newID := uuid.New()
	req := feedbackItemDTO.CreateFeedbackItemRequest{
		FeedbackType:                db.FeedbackTypeNegative,
		FeedbackText:                "Feedback to be deleted",
		CourseParticipationID:       suite.testCourseParticipationID,
		CoursePhaseID:               suite.testCoursePhaseID,
		AuthorCourseParticipationID: suite.testAuthorID,
		Type:                        assessmentType.Self,
	}

	// Insert directly using the database
	err := suite.feedbackItemService.queries.CreateFeedbackItem(suite.suiteCtx, db.CreateFeedbackItemParams{
		ID:                          newID,
		FeedbackType:                req.FeedbackType,
		FeedbackText:                req.FeedbackText,
		CourseParticipationID:       req.CourseParticipationID,
		CoursePhaseID:               req.CoursePhaseID,
		AuthorCourseParticipationID: req.AuthorCourseParticipationID,
		Type:                        assessmentType.MapDTOtoDBAssessmentType(req.Type),
	})
	assert.NoError(suite.T(), err)

	// Now delete it
	err = DeleteFeedbackItem(suite.suiteCtx, newID)
	assert.NoError(suite.T(), err)

	// Verify it's deleted
	_, err = GetFeedbackItem(suite.suiteCtx, newID)
	assert.Error(suite.T(), err)
}

func (suite *FeedbackItemServiceTestSuite) TestListFeedbackItemsForCoursePhase() {
	feedbackItems, err := ListFeedbackItemsForCoursePhase(suite.suiteCtx, suite.testCoursePhaseID)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(feedbackItems), 2) // We have at least 2 in test data
}

func (suite *FeedbackItemServiceTestSuite) TestListFeedbackItemsForParticipantInPhase() {
	feedbackItems, err := ListFeedbackItemsForParticipantInPhase(suite.suiteCtx, suite.testCourseParticipationID, suite.testCoursePhaseID)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(feedbackItems), 1)
}

func (suite *FeedbackItemServiceTestSuite) TestIsFeedbackItemAuthor() {
	isAuthor := IsFeedbackItemAuthor(suite.suiteCtx, suite.testFeedbackItemID, suite.testAuthorID)
	assert.True(suite.T(), isAuthor)

	// Test with wrong author
	wrongAuthor := uuid.MustParse("03234567-1234-1234-1234-123456789012")
	isAuthor = IsFeedbackItemAuthor(suite.suiteCtx, suite.testFeedbackItemID, wrongAuthor)
	assert.False(suite.T(), isAuthor)
}

func TestFeedbackItemServiceTestSuite(t *testing.T) {
	suite.Run(t, new(FeedbackItemServiceTestSuite))
}
