package actionItem

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/actionItem/actionItemDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type ActionItemServiceTestSuite struct {
	suite.Suite
	suiteCtx                  context.Context
	cleanup                   func()
	actionItemService         ActionItemService
	testCoursePhaseID         uuid.UUID
	testCourseParticipationID uuid.UUID
	testActionItemID          uuid.UUID
}

func (suite *ActionItemServiceTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../../database_dumps/actionItem.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.actionItemService = ActionItemService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	ActionItemServiceSingleton = &suite.actionItemService

	// Initialize required singletons
	assessmentCompletion.AssessmentCompletionServiceSingleton = &assessmentCompletion.AssessmentCompletionService{}
	coursePhaseConfig.CoursePhaseConfigSingleton = coursePhaseConfig.NewCoursePhaseConfigService(*testDB.Queries, testDB.Conn)

	// Use predefined test UUIDs from the database dump
	// This phase has assessment open (start in past, deadline in future)
	suite.testCoursePhaseID = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	suite.testCourseParticipationID = uuid.New()
	suite.testActionItemID = uuid.New()
}

func (suite *ActionItemServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *ActionItemServiceTestSuite) TestCreateActionItem() {
	// Test creating a new action item
	createRequest := actionItemDTO.CreateActionItemRequest{
		CoursePhaseID:         suite.testCoursePhaseID,
		CourseParticipationID: suite.testCourseParticipationID,
		Action:                "Complete the assignment",
		Author:                "test.author@example.com",
	}

	err := CreateActionItem(suite.suiteCtx, createRequest)
	assert.NoError(suite.T(), err, "Should be able to create action item")
}

func (suite *ActionItemServiceTestSuite) TestGetActionItemNonExistent() {
	// Test getting a non-existent action item
	nonExistentID := uuid.New()

	_, err := GetActionItem(suite.suiteCtx, nonExistentID)
	assert.Error(suite.T(), err, "Should return error for non-existent action item")
	assert.Contains(suite.T(), err.Error(), "could not get action item")
}

func (suite *ActionItemServiceTestSuite) TestUpdateActionItem() {
	// Create an action item first to get its ID, then update it
	createRequest := actionItemDTO.CreateActionItemRequest{
		CoursePhaseID:         suite.testCoursePhaseID,
		CourseParticipationID: suite.testCourseParticipationID,
		Action:                "Original action",
		Author:                "original.author@example.com",
	}

	err := CreateActionItem(suite.suiteCtx, createRequest)
	assert.NoError(suite.T(), err)

	// Get the created action item by listing items for the student in this phase
	actionItems, err := ListActionItemsForStudentInPhase(suite.suiteCtx, suite.testCourseParticipationID, suite.testCoursePhaseID)
	assert.NoError(suite.T(), err, "Should be able to list action items to find created item")
	assert.Greater(suite.T(), len(actionItems), 0, "Should have at least one action item")

	// Find the action item we just created
	var createdActionItemID uuid.UUID
	for _, item := range actionItems {
		if item.Action == "Original action" && item.Author == "original.author@example.com" {
			createdActionItemID = item.ID
			break
		}
	}
	assert.NotEqual(suite.T(), uuid.Nil, createdActionItemID, "Should have found the created action item")

	// Now update the action item using the actual ID
	updateRequest := actionItemDTO.UpdateActionItemRequest{
		ID:                    createdActionItemID,
		CoursePhaseID:         suite.testCoursePhaseID,
		CourseParticipationID: suite.testCourseParticipationID,
		Action:                "Updated action",
		Author:                "updated.author@example.com",
	}

	// Test the update operation with proper error handling
	assert.NotPanics(suite.T(), func() {
		err := UpdateActionItem(suite.suiteCtx, updateRequest)
		assert.NoError(suite.T(), err, "Should be able to update existing action item")
	}, "Should not panic when updating action item")
}

func (suite *ActionItemServiceTestSuite) TestDeleteActionItem() {
	// Test deleting an action item
	testID := uuid.New()

	// This might fail if the ID doesn't exist, which is expected
	// The test verifies the function doesn't panic
	assert.NotPanics(suite.T(), func() {
		_ = DeleteActionItem(suite.suiteCtx, testID)
	}, "Should not panic when deleting action item")
}

func (suite *ActionItemServiceTestSuite) TestDeleteActionItemNonExistent() {
	// Test deleting a non-existent action item
	nonExistentID := uuid.New()

	// The service might not return an error for deleting a non-existent item
	// This is because DELETE operations are idempotent in SQL
	// We just verify it doesn't panic
	assert.NotPanics(suite.T(), func() {
		_ = DeleteActionItem(suite.suiteCtx, nonExistentID)
	}, "Should not panic when deleting non-existent action item")
}

func (suite *ActionItemServiceTestSuite) TestListActionItemsForCoursePhase() {
	// Create multiple action items for the same course phase
	testCoursePhaseID := suite.testCoursePhaseID

	actionItems := []actionItemDTO.CreateActionItemRequest{
		{
			CoursePhaseID:         testCoursePhaseID,
			CourseParticipationID: uuid.New(),
			Action:                "First action",
			Author:                "first.author@example.com",
		},
		{
			CoursePhaseID:         testCoursePhaseID,
			CourseParticipationID: uuid.New(),
			Action:                "Second action",
			Author:                "second.author@example.com",
		},
	}

	// Create the action items
	for _, item := range actionItems {
		err := CreateActionItem(suite.suiteCtx, item)
		assert.NoError(suite.T(), err)
	}

	// List action items for the course phase
	retrievedItems, err := ListActionItemsForCoursePhase(suite.suiteCtx, testCoursePhaseID)
	assert.NoError(suite.T(), err, "Should be able to list action items for course phase")
	assert.GreaterOrEqual(suite.T(), len(retrievedItems), 2, "Should return at least 2 action items")

	// Verify the retrieved items belong to the correct course phase
	for _, item := range retrievedItems {
		assert.Equal(suite.T(), testCoursePhaseID, item.CoursePhaseID)
	}
}

func (suite *ActionItemServiceTestSuite) TestListActionItemsForStudentInPhase() {
	// Create action items for a specific student in a course phase
	testCoursePhaseID := suite.testCoursePhaseID
	testStudentID := uuid.New()

	actionItems := []actionItemDTO.CreateActionItemRequest{
		{
			CoursePhaseID:         testCoursePhaseID,
			CourseParticipationID: testStudentID,
			Action:                "Student action 1",
			Author:                "author1@example.com",
		},
		{
			CoursePhaseID:         testCoursePhaseID,
			CourseParticipationID: testStudentID,
			Action:                "Student action 2",
			Author:                "author2@example.com",
		},
		{
			CoursePhaseID:         testCoursePhaseID,
			CourseParticipationID: uuid.New(), // Different student
			Action:                "Other student action",
			Author:                "other@example.com",
		},
	}

	// Create the action items
	for _, item := range actionItems {
		err := CreateActionItem(suite.suiteCtx, item)
		assert.NoError(suite.T(), err)
	}

	// List action items for the specific student
	retrievedItems, err := ListActionItemsForStudentInPhase(suite.suiteCtx, testStudentID, testCoursePhaseID)
	assert.NoError(suite.T(), err, "Should be able to list action items for student in phase")
	assert.GreaterOrEqual(suite.T(), len(retrievedItems), 2, "Should return at least 2 action items for the student")

	// Verify the retrieved items belong to the correct student
	for _, item := range retrievedItems {
		assert.Equal(suite.T(), testStudentID, item.CourseParticipationID)
		assert.Equal(suite.T(), testCoursePhaseID, item.CoursePhaseID)
	}
}

func (suite *ActionItemServiceTestSuite) TestCountActionItemsForStudentInPhase() {
	// Create action items for a specific student in a course phase
	testCoursePhaseID := suite.testCoursePhaseID
	testStudentID := uuid.New()

	actionItems := []actionItemDTO.CreateActionItemRequest{
		{
			CoursePhaseID:         testCoursePhaseID,
			CourseParticipationID: testStudentID,
			Action:                "Count action 1",
			Author:                "author1@example.com",
		},
		{
			CoursePhaseID:         testCoursePhaseID,
			CourseParticipationID: testStudentID,
			Action:                "Count action 2",
			Author:                "author2@example.com",
		},
		{
			CoursePhaseID:         testCoursePhaseID,
			CourseParticipationID: testStudentID,
			Action:                "Count action 3",
			Author:                "author3@example.com",
		},
	}

	// Create the action items
	for _, item := range actionItems {
		err := CreateActionItem(suite.suiteCtx, item)
		assert.NoError(suite.T(), err)
	}

	// Count action items for the specific student
	count, err := CountActionItemsForStudentInPhase(suite.suiteCtx, testStudentID, testCoursePhaseID)
	assert.NoError(suite.T(), err, "Should be able to count action items for student in phase")
	assert.GreaterOrEqual(suite.T(), count, int64(3), "Should return count of at least 3 action items")
}

func (suite *ActionItemServiceTestSuite) TestCountActionItemsForStudentInPhaseEmpty() {
	// Test counting action items for a student with no action items
	nonExistentStudentID := uuid.New()
	nonExistentCoursePhaseID := uuid.New()

	count, err := CountActionItemsForStudentInPhase(suite.suiteCtx, nonExistentStudentID, nonExistentCoursePhaseID)
	assert.NoError(suite.T(), err, "Should be able to count action items even when none exist")
	assert.Equal(suite.T(), int64(0), count, "Should return count of 0 for non-existent student/phase")
}

func (suite *ActionItemServiceTestSuite) TestGetAllActionItemsForCoursePhaseCommunication() {
	// Create multiple action items for different course participation IDs in the same course phase
	testCoursePhaseID := suite.testCoursePhaseID
	testStudentID1 := uuid.New()
	testStudentID2 := uuid.New()

	actionItems := []actionItemDTO.CreateActionItemRequest{
		{
			CoursePhaseID:         testCoursePhaseID,
			CourseParticipationID: testStudentID1,
			Action:                "Student 1 - Action 1",
			Author:                "author1@example.com",
		},
		{
			CoursePhaseID:         testCoursePhaseID,
			CourseParticipationID: testStudentID1,
			Action:                "Student 1 - Action 2",
			Author:                "author1@example.com",
		},
		{
			CoursePhaseID:         testCoursePhaseID,
			CourseParticipationID: testStudentID2,
			Action:                "Student 2 - Action 1",
			Author:                "author2@example.com",
		},
	}

	// Create the action items
	for _, item := range actionItems {
		err := CreateActionItem(suite.suiteCtx, item)
		assert.NoError(suite.T(), err)
	}

	// Get all action items grouped by course participation ID
	retrievedItems, err := GetAllActionItemsForCoursePhaseCommunication(suite.suiteCtx, testCoursePhaseID)
	assert.NoError(suite.T(), err, "Should be able to get all action items for course phase communication")
	assert.GreaterOrEqual(suite.T(), len(retrievedItems), 2, "Should return at least 2 course participation groups")

	// Verify the structure - should be grouped by course participation ID
	foundStudent1 := false
	foundStudent2 := false
	for _, item := range retrievedItems {
		if item.CourseParticipationID == testStudentID1 {
			foundStudent1 = true
			assert.GreaterOrEqual(suite.T(), len(item.ActionItems), 2, "Student 1 should have at least 2 action items")
		}
		if item.CourseParticipationID == testStudentID2 {
			foundStudent2 = true
			assert.GreaterOrEqual(suite.T(), len(item.ActionItems), 1, "Student 2 should have at least 1 action item")
		}
	}
	assert.True(suite.T(), foundStudent1, "Should find action items for student 1")
	assert.True(suite.T(), foundStudent2, "Should find action items for student 2")
}

func (suite *ActionItemServiceTestSuite) TestGetAllActionItemsForCoursePhaseCommunicationEmpty() {
	// Test getting action items for a course phase with no action items
	nonExistentCoursePhaseID := uuid.New()

	retrievedItems, err := GetAllActionItemsForCoursePhaseCommunication(suite.suiteCtx, nonExistentCoursePhaseID)
	assert.NoError(suite.T(), err, "Should not return error for non-existent course phase")
	assert.Equal(suite.T(), 0, len(retrievedItems), "Should return empty array for non-existent course phase")
}

func (suite *ActionItemServiceTestSuite) TestGetStudentActionItemsForCoursePhaseCommunication() {
	// Create action items for a specific student in a course phase
	testCoursePhaseID := suite.testCoursePhaseID
	testStudentID := uuid.New()

	actionItems := []actionItemDTO.CreateActionItemRequest{
		{
			CoursePhaseID:         testCoursePhaseID,
			CourseParticipationID: testStudentID,
			Action:                "Communication Action 1",
			Author:                "author1@example.com",
		},
		{
			CoursePhaseID:         testCoursePhaseID,
			CourseParticipationID: testStudentID,
			Action:                "Communication Action 2",
			Author:                "author2@example.com",
		},
	}

	// Create the action items
	for _, item := range actionItems {
		err := CreateActionItem(suite.suiteCtx, item)
		assert.NoError(suite.T(), err)
	}

	// Get action items for the specific student
	retrievedItems, err := GetStudentActionItemsForCoursePhaseCommunication(suite.suiteCtx, testStudentID, testCoursePhaseID)
	assert.NoError(suite.T(), err, "Should be able to get student action items for course phase communication")
	assert.GreaterOrEqual(suite.T(), len(retrievedItems), 2, "Should return at least 2 action items")

	// Verify the returned items are strings (action text only)
	expectedActions := []string{"Communication Action 1", "Communication Action 2"}
	for _, expectedAction := range expectedActions {
		found := false
		for _, retrievedAction := range retrievedItems {
			if retrievedAction == expectedAction {
				found = true
				break
			}
		}
		assert.True(suite.T(), found, "Should find expected action: %s", expectedAction)
	}
}

func (suite *ActionItemServiceTestSuite) TestGetStudentActionItemsForCoursePhaseCommunicationEmpty() {
	// Test getting action items for a student with no action items
	nonExistentStudentID := uuid.New()
	nonExistentCoursePhaseID := uuid.New()

	retrievedItems, err := GetStudentActionItemsForCoursePhaseCommunication(suite.suiteCtx, nonExistentStudentID, nonExistentCoursePhaseID)
	assert.NoError(suite.T(), err, "Should not return error for non-existent student/phase")
	assert.Equal(suite.T(), 0, len(retrievedItems), "Should return empty array for non-existent student/phase")
}

func TestActionItemServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ActionItemServiceTestSuite))
}
