package feedbackItem

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/feedbackItem/feedbackItemDTO"
)

type FeedbackItemServiceTestSuite struct {
	suite.Suite
	suiteCtx                  context.Context
	cleanup                   func()
	feedbackItemService       *FeedbackItemService
	testCoursePhaseID         uuid.UUID
	testCourseParticipationID uuid.UUID
	testAuthorID              uuid.UUID
	testFeedbackItemID        uuid.UUID
	testAuthHeader            string
}

func (suite *FeedbackItemServiceTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.suiteCtx, "../../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../../database_dumps/feedbackItems.sql")
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.feedbackItemService = NewFeedbackItemService(*testDB.Queries, testDB.Conn,
		evaluationCompletion.NewEvaluationCompletionService(
			*testDB.Queries,
			testDB.Conn,
			coursePhaseConfig.GetTeamsForCoursePhase,
			coursePhaseConfig.NewCoursePhaseConfigService(*testDB.Queries, testDB.Conn, assessmentSchemas.NewAssessmentSchemaService(*testDB.Queries, testDB.Conn)),
		))

	// Use predefined test UUIDs from the test data that match feedbackItems.sql
	suite.testCoursePhaseID = uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	suite.testCourseParticipationID = uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")
	suite.testAuthorID = uuid.MustParse("da42e447-60f9-4fe0-b297-2dae3f924fd7")
	suite.testFeedbackItemID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	suite.testAuthHeader = "Bearer feedback-item-service-test-token"
}

func (suite *FeedbackItemServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

// seedFeedbackItem inserts a self feedback item authored by the given participation
// so tests that mutate rows do not depend on each other's ordering.
func (suite *FeedbackItemServiceTestSuite) seedFeedbackItem(authorID uuid.UUID) uuid.UUID {
	return suite.seedFeedbackItemInPhase(authorID, suite.testCoursePhaseID)
}

func (suite *FeedbackItemServiceTestSuite) seedFeedbackItemInPhase(authorID, coursePhaseID uuid.UUID) uuid.UUID {
	feedbackItemID := uuid.New()

	err := suite.feedbackItemService.queries.CreateFeedbackItem(suite.suiteCtx, db.CreateFeedbackItemParams{
		ID:                          feedbackItemID,
		FeedbackType:                db.FeedbackTypePositive,
		FeedbackText:                "Original feedback text",
		CourseParticipationID:       authorID,
		CoursePhaseID:               coursePhaseID,
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

	err := suite.feedbackItemService.CreateFeedbackItem(suite.suiteCtx, suite.testAuthHeader, req)
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

	err := suite.feedbackItemService.CreateFeedbackItem(suite.suiteCtx, suite.testAuthHeader, req)
	assert.ErrorIs(suite.T(), err, evaluationCompletion.ErrSelfEvaluationTargetMismatch)
}

func (suite *FeedbackItemServiceTestSuite) TestUpdateFeedbackItem() {
	updateFeedbackItemID := suite.seedFeedbackItem(suite.testAuthorID)
	req := feedbackItemDTO.UpdateFeedbackItemRequest{
		FeedbackType: db.FeedbackTypeNegative, // Change from positive to negative
		FeedbackText: "Updated feedback text",
	}

	err := suite.feedbackItemService.UpdateFeedbackItem(suite.suiteCtx, suite.testAuthHeader, updateFeedbackItemID, suite.testCoursePhaseID, suite.testAuthorID, req)
	assert.NoError(suite.T(), err)

	updated, err := suite.feedbackItemService.GetFeedbackItem(suite.suiteCtx, updateFeedbackItemID)
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

	err := suite.feedbackItemService.UpdateFeedbackItem(suite.suiteCtx, suite.testAuthHeader, updateFeedbackItemID, otherPhaseID, suite.testAuthorID, req)
	assert.ErrorIs(suite.T(), err, ErrFeedbackItemNotFound)
}

func (suite *FeedbackItemServiceTestSuite) TestUpdateFeedbackItemRejectsNonAuthor() {
	// Feedback item 22222222 is authored by the lecturer, not testAuthorID
	updateFeedbackItemID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	req := feedbackItemDTO.UpdateFeedbackItemRequest{
		FeedbackType: db.FeedbackTypePositive,
		FeedbackText: "Rewritten by a non-author",
	}

	err := suite.feedbackItemService.UpdateFeedbackItem(suite.suiteCtx, suite.testAuthHeader, updateFeedbackItemID, suite.testCoursePhaseID, suite.testAuthorID, req)
	assert.ErrorIs(suite.T(), err, ErrNotFeedbackItemAuthor)

	unchanged, err := suite.feedbackItemService.GetFeedbackItem(suite.suiteCtx, updateFeedbackItemID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Need to improve time management", unchanged.FeedbackText)
}

func (suite *FeedbackItemServiceTestSuite) TestGetFeedbackItem() {
	feedbackItem, err := suite.feedbackItemService.GetFeedbackItem(suite.suiteCtx, suite.testFeedbackItemID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.testFeedbackItemID, feedbackItem.ID)
	assert.Equal(suite.T(), "Great teamwork and communication skills!", feedbackItem.FeedbackText)
	assert.Equal(suite.T(), db.FeedbackTypePositive, feedbackItem.FeedbackType)
}

func (suite *FeedbackItemServiceTestSuite) TestDeleteFeedbackItem() {
	feedbackItemID := suite.seedFeedbackItem(suite.testAuthorID)

	err := suite.feedbackItemService.DeleteFeedbackItem(suite.suiteCtx, suite.testAuthHeader, feedbackItemID, suite.testCoursePhaseID, suite.testAuthorID)
	assert.NoError(suite.T(), err)

	_, err = suite.feedbackItemService.GetFeedbackItem(suite.suiteCtx, feedbackItemID)
	assert.Error(suite.T(), err)
}

// The tests below are regression tests for GHSA-h36r-xgjv-r8cx: deletion must be bound
// to the phase in the URL and respect the evaluation lock, exactly like updates.

func (suite *FeedbackItemServiceTestSuite) TestDeleteFeedbackItemRejectsWrongPhase() {
	feedbackItemID := suite.seedFeedbackItem(suite.testAuthorID)
	otherPhaseID := uuid.MustParse("34561b6b-3c3a-4bc6-ba42-69eeb1514da9")

	err := suite.feedbackItemService.DeleteFeedbackItem(suite.suiteCtx, suite.testAuthHeader, feedbackItemID, otherPhaseID, suite.testAuthorID)
	assert.ErrorIs(suite.T(), err, ErrFeedbackItemNotFound)

	_, err = suite.feedbackItemService.GetFeedbackItem(suite.suiteCtx, feedbackItemID)
	assert.NoError(suite.T(), err)
}

func (suite *FeedbackItemServiceTestSuite) TestDeleteFeedbackItemRejectsNonAuthor() {
	// Feedback item 22222222 is authored by the lecturer, not testAuthorID
	feedbackItemID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	err := suite.feedbackItemService.DeleteFeedbackItem(suite.suiteCtx, suite.testAuthHeader, feedbackItemID, suite.testCoursePhaseID, suite.testAuthorID)
	assert.ErrorIs(suite.T(), err, ErrNotFeedbackItemAuthor)

	_, err = suite.feedbackItemService.GetFeedbackItem(suite.suiteCtx, feedbackItemID)
	assert.NoError(suite.T(), err)
}

func (suite *FeedbackItemServiceTestSuite) TestDeleteFeedbackItemRejectsCompletedEvaluation() {
	// The seed marks the self evaluation of testCourseParticipationID in phase 2 as completed.
	lockedPhaseID := uuid.MustParse("34561b6b-3c3a-4bc6-ba42-69eeb1514da9")
	feedbackItemID := suite.seedFeedbackItemInPhase(suite.testCourseParticipationID, lockedPhaseID)

	err := suite.feedbackItemService.DeleteFeedbackItem(suite.suiteCtx, suite.testAuthHeader, feedbackItemID, lockedPhaseID, suite.testCourseParticipationID)
	assert.ErrorIs(suite.T(), err, evaluationCompletion.ErrEvaluationAlreadyCompleted)

	_, err = suite.feedbackItemService.GetFeedbackItem(suite.suiteCtx, feedbackItemID)
	assert.NoError(suite.T(), err)
}

func (suite *FeedbackItemServiceTestSuite) TestDeleteFeedbackItemRejectsEvaluationNotStarted() {
	notStartedPhaseID := uuid.MustParse("44561b6b-3c3a-4bc6-ba42-69eeb1514da9")
	feedbackItemID := suite.seedFeedbackItemInPhase(suite.testAuthorID, notStartedPhaseID)

	err := suite.feedbackItemService.DeleteFeedbackItem(suite.suiteCtx, suite.testAuthHeader, feedbackItemID, notStartedPhaseID, suite.testAuthorID)
	assert.ErrorIs(suite.T(), err, coursePhaseConfig.ErrNotStarted)

	_, err = suite.feedbackItemService.GetFeedbackItem(suite.suiteCtx, feedbackItemID)
	assert.NoError(suite.T(), err)
}

func (suite *FeedbackItemServiceTestSuite) TestDeleteFeedbackItemUnknownID() {
	err := suite.feedbackItemService.DeleteFeedbackItem(suite.suiteCtx, suite.testAuthHeader, uuid.New(), suite.testCoursePhaseID, suite.testAuthorID)
	assert.ErrorIs(suite.T(), err, ErrFeedbackItemNotFound)
}

func (suite *FeedbackItemServiceTestSuite) TestListFeedbackItemsForCoursePhase() {
	feedbackItems, err := suite.feedbackItemService.ListFeedbackItemsForCoursePhase(suite.suiteCtx, suite.testCoursePhaseID)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(feedbackItems), 2) // We have at least 2 in test data
}

func (suite *FeedbackItemServiceTestSuite) TestListFeedbackItemsForParticipantInPhase() {
	feedbackItems, err := suite.feedbackItemService.ListFeedbackItemsForParticipantInPhase(suite.suiteCtx, suite.testCourseParticipationID, suite.testCoursePhaseID)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(feedbackItems), 1)
}

func TestFeedbackItemServiceTestSuite(t *testing.T) {
	suite.Run(t, new(FeedbackItemServiceTestSuite))
}
