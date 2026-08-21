package copy

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CopyServiceTestSuite struct {
	suite.Suite
	suiteCtx    context.Context
	cleanup     func()
	copyService CopyService
}

func (suite *CopyServiceTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/coursePhaseConfig.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.copyService = CopyService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	CopyServiceSingleton = &suite.copyService
}

func (suite *CopyServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *CopyServiceTestSuite) TestHandlePhaseCopy_Success() {
	// Create source course phase config
	sourceCoursePhaseID := uuid.New()
	targetCoursePhaseID := uuid.New()
	assessmentSchemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	selfEvalSchemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	peerEvalSchemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
	tutorEvalSchemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440003")

	// Create valid timestamps for NOT NULL fields
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}

	// Create source config
	err := suite.copyService.queries.CreateOrUpdateCoursePhaseConfig(suite.suiteCtx, db.CreateOrUpdateCoursePhaseConfigParams{
		AssessmentSchemaID:       assessmentSchemaID,
		CoursePhaseID:            sourceCoursePhaseID,
		Start:                    now,
		Deadline:                 pgtype.Timestamptz{Valid: false},
		SelfEvaluationEnabled:    true,
		SelfEvaluationSchema:     selfEvalSchemaID,
		SelfEvaluationStart:      now,
		SelfEvaluationDeadline:   now,
		PeerEvaluationEnabled:    true,
		PeerEvaluationSchema:     peerEvalSchemaID,
		PeerEvaluationStart:      now,
		PeerEvaluationDeadline:   now,
		TutorEvaluationEnabled:   true,
		TutorEvaluationSchema:    tutorEvalSchemaID,
		TutorEvaluationStart:     pgtype.Timestamptz{Valid: false},
		TutorEvaluationDeadline:  pgtype.Timestamptz{Valid: false},
		EvaluationResultsVisible: true,
		GradeSuggestionVisible:   pgtype.Bool{Bool: true, Valid: true},
		ActionItemsVisible:       pgtype.Bool{Bool: false, Valid: true},
		GradingSheetVisible:      pgtype.Bool{Bool: false, Valid: true},
	})
	assert.NoError(suite.T(), err)

	// Execute copy
	handler := &AssessmentCopyHandler{}
	req := promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: sourceCoursePhaseID,
		TargetCoursePhaseID: targetCoursePhaseID,
	}

	// Create a proper gin context for the test
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/copy", nil)

	err = handler.HandlePhaseCopy(c, req)
	assert.NoError(suite.T(), err)

	// Verify the target config was created
	targetConfig, err := suite.copyService.queries.GetCoursePhaseConfig(suite.suiteCtx, targetCoursePhaseID)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), assessmentSchemaID, targetConfig.AssessmentSchemaID)
	assert.Equal(suite.T(), targetCoursePhaseID, targetConfig.CoursePhaseID)
	assert.Equal(suite.T(), true, targetConfig.SelfEvaluationEnabled)
	assert.Equal(suite.T(), selfEvalSchemaID, targetConfig.SelfEvaluationSchema)
	assert.Equal(suite.T(), true, targetConfig.PeerEvaluationEnabled)
	assert.Equal(suite.T(), peerEvalSchemaID, targetConfig.PeerEvaluationSchema)
	assert.Equal(suite.T(), true, targetConfig.TutorEvaluationEnabled)
	assert.Equal(suite.T(), tutorEvalSchemaID, targetConfig.TutorEvaluationSchema)
	assert.Equal(suite.T(), true, targetConfig.EvaluationResultsVisible)
	assert.Equal(suite.T(), true, targetConfig.GradeSuggestionVisible)
	assert.Equal(suite.T(), false, targetConfig.ActionItemsVisible)
	assert.Equal(suite.T(), false, targetConfig.ResultsReleased)
}

func (suite *CopyServiceTestSuite) TestHandlePhaseCopy_SameSourceAndTarget() {
	coursePhaseID := uuid.New()

	handler := &AssessmentCopyHandler{}
	req := promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: coursePhaseID,
		TargetCoursePhaseID: coursePhaseID,
	}

	// Create a proper gin context for the test
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/copy", nil)

	err := handler.HandlePhaseCopy(c, req)
	assert.NoError(suite.T(), err, "Should not error when source and target are the same")
}

func (suite *CopyServiceTestSuite) TestHandlePhaseCopy_NonExistentSource() {
	nonExistentSourceID := uuid.New()
	targetCoursePhaseID := uuid.New()

	handler := &AssessmentCopyHandler{}
	req := promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: nonExistentSourceID,
		TargetCoursePhaseID: targetCoursePhaseID,
	}

	// Create a proper gin context for the test
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/copy", nil)

	err := handler.HandlePhaseCopy(c, req)
	assert.Error(suite.T(), err, "Should error when source course phase config doesn't exist")
}

// seedCoursePhaseConfig writes a minimal config so a copy has something to read or overwrite.
func (suite *CopyServiceTestSuite) seedCoursePhaseConfig(coursePhaseID uuid.UUID, assessmentEnabled bool) {
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	err := suite.copyService.queries.CreateOrUpdateCoursePhaseConfig(suite.suiteCtx, db.CreateOrUpdateCoursePhaseConfigParams{
		AssessmentSchemaID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		CoursePhaseID:          coursePhaseID,
		Start:                  now,
		Deadline:               pgtype.Timestamptz{Valid: false},
		SelfEvaluationSchema:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		SelfEvaluationStart:    now,
		SelfEvaluationDeadline: now,
		PeerEvaluationSchema:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		PeerEvaluationStart:    now,
		PeerEvaluationDeadline: now,
		TutorEvaluationSchema:  uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
		AssessmentEnabled:      assessmentEnabled,
	})
	assert.NoError(suite.T(), err)
}

func (suite *CopyServiceTestSuite) copyPhase(sourceID, targetID uuid.UUID) error {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/copy", nil)

	return (&AssessmentCopyHandler{}).HandlePhaseCopy(c, promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: sourceID,
		TargetCoursePhaseID: targetID,
	})
}

func (suite *CopyServiceTestSuite) TestHandlePhaseCopy_DisabledSourceCannotDisableTargetWithData() {
	sourceCoursePhaseID := uuid.New()
	targetCoursePhaseID := uuid.New()

	suite.seedCoursePhaseConfig(sourceCoursePhaseID, false)
	suite.seedCoursePhaseConfig(targetCoursePhaseID, true)

	_, err := suite.copyService.conn.Exec(suite.suiteCtx,
		`INSERT INTO assessment_completion (course_participation_id, course_phase_id, completed_at, author)
		 VALUES ($1, $2, NOW(), $3)`,
		uuid.New(), targetCoursePhaseID, "Tutor")
	assert.NoError(suite.T(), err)

	err = suite.copyPhase(sourceCoursePhaseID, targetCoursePhaseID)
	assert.Equal(suite.T(), coursePhaseConfig.ErrCannotDisableAssessmentWithData, err)

	targetConfig, err := suite.copyService.queries.GetCoursePhaseConfig(suite.suiteCtx, targetCoursePhaseID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), targetConfig.AssessmentEnabled, "Target with assessment data should stay enabled")
}

func (suite *CopyServiceTestSuite) TestHandlePhaseCopy_DisabledSourceCopiesIntoEmptyTarget() {
	sourceCoursePhaseID := uuid.New()
	targetCoursePhaseID := uuid.New()

	suite.seedCoursePhaseConfig(sourceCoursePhaseID, false)

	assert.NoError(suite.T(), suite.copyPhase(sourceCoursePhaseID, targetCoursePhaseID))

	targetConfig, err := suite.copyService.queries.GetCoursePhaseConfig(suite.suiteCtx, targetCoursePhaseID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), targetConfig.AssessmentEnabled, "Empty target should adopt the source's disabled state")
}

func TestCopyServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CopyServiceTestSuite))
}
