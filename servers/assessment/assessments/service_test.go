package assessments

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type AssessmentServiceTestSuite struct {
	suite.Suite
	suiteCtx context.Context
	cleanup  func()
	service  AssessmentService
}

func (suite *AssessmentServiceTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	// initialize test database from dump
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/assessments.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.service = AssessmentService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	AssessmentServiceSingleton = &suite.service

	// Initialize CoursePhaseConfigSingleton to prevent nil pointer dereference
	coursePhaseConfig.CoursePhaseConfigSingleton = coursePhaseConfig.NewCoursePhaseConfigService(*testDB.Queries, testDB.Conn)
}

func (suite *AssessmentServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *AssessmentServiceTestSuite) TestListAssessmentsByCoursePhase() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	items, err := ListAssessmentsByCoursePhase(suite.suiteCtx, phaseID)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(items), 0, "Expected at least one assessment for phase")
}

func (suite *AssessmentServiceTestSuite) TestGetAssessment() {
	id := uuid.MustParse("1950fdb7-d736-4fe6-81f9-b8b1cf7c85df")
	a, err := GetAssessment(suite.suiteCtx, id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), id, a.ID, "Assessment ID should match")
}

func (suite *AssessmentServiceTestSuite) TestGetAssessmentNotFound() {
	id := uuid.New()
	_, err := GetAssessment(suite.suiteCtx, id)
	assert.Error(suite.T(), err, "Expected error for non-existent assessment")
}

func (suite *AssessmentServiceTestSuite) TestListAssessmentsByStudentInPhase() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")
	items, err := ListAssessmentsByStudentInPhase(suite.suiteCtx, partID, phaseID)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(items), 0, "Expected at least one assessment for student in phase")
}

func (suite *AssessmentServiceTestSuite) TestDeleteAssessmentNonExisting() {
	id := uuid.New()
	err := DeleteAssessment(suite.suiteCtx, id)
	assert.NoError(suite.T(), err, "Deleting non-existent assessment should not error")
}

func (suite *AssessmentServiceTestSuite) TestCreateOrUpdateAssessmentWithEmptyScoreLevel() {
	req := assessmentDTO.CreateOrUpdateAssessmentRequest{
		CourseParticipationID: uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7"),
		CoursePhaseID:         uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9"),
		CompetencyID:          uuid.MustParse("01935143-5e85-7e1d-81bb-96fb3ebf34aa"),
		ScoreLevel:            "", // Empty scoreLevel should be rejected
		Comment:               "",
		Examples:              "",
		Author:                "Test Author",
	}

	err := CreateOrUpdateAssessment(suite.suiteCtx, req)
	assert.Error(suite.T(), err, "Expected error for empty scoreLevel")
	assert.Equal(suite.T(), ErrInvalidScoreLevel, err, "Expected ErrInvalidScoreLevel")
}

func (suite *AssessmentServiceTestSuite) TestCreateOrUpdateAssessmentLowScoreLevelWithoutComment() {
	req := assessmentDTO.CreateOrUpdateAssessmentRequest{
		CourseParticipationID: uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7"),
		CoursePhaseID:         uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9"),
		CompetencyID:          uuid.MustParse("01935143-5e85-7e1d-81bb-96fb3ebf34aa"),
		ScoreLevel:            scoreLevelDTO.ScoreLevelBad,
		Comment:               "", // Empty comment should fail for low scoreLevel
		Examples:              "Some example",
		Author:                "Test Author",
	}

	err := CreateOrUpdateAssessment(suite.suiteCtx, req)
	assert.Error(suite.T(), err, "Expected error for low scoreLevel without comment")
	assert.Equal(suite.T(), ErrValidationFailed, err, "Expected ErrValidationFailed")
}

func (suite *AssessmentServiceTestSuite) TestCreateOrUpdateAssessmentLowScoreLevelWithoutExamples() {
	req := assessmentDTO.CreateOrUpdateAssessmentRequest{
		CourseParticipationID: uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7"),
		CoursePhaseID:         uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9"),
		CompetencyID:          uuid.MustParse("01935143-5e85-7e1d-81bb-96fb3ebf34aa"),
		ScoreLevel:            scoreLevelDTO.ScoreLevelOk,
		Comment:               "Some comment",
		Examples:              "", // Empty examples should fail for low scoreLevel
		Author:                "Test Author",
	}

	err := CreateOrUpdateAssessment(suite.suiteCtx, req)
	assert.Error(suite.T(), err, "Expected error for low scoreLevel without examples")
	assert.Equal(suite.T(), ErrValidationFailed, err, "Expected ErrValidationFailed")
}

func TestAssessmentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AssessmentServiceTestSuite))
}
