package categoryAssessment

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/categoryAssessment/categoryAssessmentDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// The stubbed core lists every participation in every phase except this one.
var outsideParticipationID = uuid.MustParse("6f1c2a4e-3b5d-4e7f-8a9b-0c1d2e3f4a5b")

func checkTestParticipant(_ context.Context, _ string, _, courseParticipationID uuid.UUID) error {
	if courseParticipationID == outsideParticipationID {
		return coursePhaseConfig.ErrParticipantNotInPhase
	}
	return nil
}

type CategoryAssessmentServiceTestSuite struct {
	suite.Suite
	suiteCtx context.Context
	cleanup  func()
	queries  db.Queries
	service  *CategoryAssessmentService
}

func (suite *CategoryAssessmentServiceTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.suiteCtx, "../../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../../database_dumps/assessments.sql")
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.queries = *testDB.Queries
	coursePhaseConfigService := coursePhaseConfig.NewCoursePhaseConfigService(*testDB.Queries, testDB.Conn, assessmentSchemas.NewAssessmentSchemaService(*testDB.Queries, testDB.Conn))
	suite.service = NewCategoryAssessmentService(*testDB.Queries, testDB.Conn, assessmentCompletion.NewAssessmentCompletionService(*testDB.Queries, testDB.Conn, coursePhaseConfigService), checkTestParticipant)
}

func (suite *CategoryAssessmentServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *CategoryAssessmentServiceTestSuite) TestCreateOrUpdateCategoryAssessment() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	req := categoryAssessmentDTO.CreateOrUpdateCategoryAssessmentRequest{
		CategoryID:            uuid.MustParse("25f1c984-ba31-4cf2-aa8e-5662721bf44e"),
		CourseParticipationID: uuid.New(),
		Comment:               "Strong understanding of version control workflows.",
		Author:                "Test Author",
		AuthorID:              "test-author-id",
	}

	err := suite.service.CreateOrUpdateCategoryAssessment(suite.suiteCtx, "", coursePhaseID, req)
	assert.NoError(suite.T(), err)

	items, err := suite.service.ListCategoryAssessmentsByStudentInPhase(suite.suiteCtx, req.CourseParticipationID, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), items, 1)
	assert.Equal(suite.T(), req.CategoryID, items[0].CategoryID)
	assert.Equal(suite.T(), req.Comment, items[0].Comment)
	assert.Equal(suite.T(), req.Author, items[0].Author)
	assert.Equal(suite.T(), req.AuthorID, items[0].AuthorID)
}

func (suite *CategoryAssessmentServiceTestSuite) TestListCategoryAssessmentsByStudentInPhase() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	req := categoryAssessmentDTO.CreateOrUpdateCategoryAssessmentRequest{
		CategoryID:            uuid.MustParse("815b159b-cab3-49b4-8060-c4722d59241d"),
		CourseParticipationID: uuid.New(),
		Comment:               "Clear UI reasoning and implementation notes.",
		Author:                "Test Author",
		AuthorID:              "test-author-id",
	}

	err := suite.service.CreateOrUpdateCategoryAssessment(suite.suiteCtx, "", coursePhaseID, req)
	assert.NoError(suite.T(), err)

	items, err := suite.service.ListCategoryAssessmentsByStudentInPhase(suite.suiteCtx, req.CourseParticipationID, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), items, 1)
	assert.Equal(suite.T(), req.CourseParticipationID, items[0].CourseParticipationID)
	assert.Equal(suite.T(), coursePhaseID, items[0].CoursePhaseID)
	assert.Equal(suite.T(), req.CategoryID, items[0].CategoryID)
}

func (suite *CategoryAssessmentServiceTestSuite) TestCreateOrUpdateRejectsCategoryOfOtherSchema() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	selfEvaluationCategoryID := uuid.New()
	err := suite.queries.CreateCategory(suite.suiteCtx, db.CreateCategoryParams{
		ID:                 selfEvaluationCategoryID,
		Name:               "Self Reflection",
		Weight:             1,
		AssessmentSchemaID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
	})
	suite.Require().NoError(err)

	for _, categoryID := range []uuid.UUID{selfEvaluationCategoryID, uuid.New()} {
		req := categoryAssessmentDTO.CreateOrUpdateCategoryAssessmentRequest{
			CategoryID:            categoryID,
			CourseParticipationID: uuid.New(),
			Comment:               "Comment on a category the phase does not assess",
			Author:                "Test Author",
			AuthorID:              "test-author-id",
		}

		err := suite.service.CreateOrUpdateCategoryAssessment(suite.suiteCtx, "", coursePhaseID, req)
		assert.ErrorIs(suite.T(), err, ErrCategoryNotInPhase)

		items, err := suite.service.ListCategoryAssessmentsByStudentInPhase(suite.suiteCtx, req.CourseParticipationID, coursePhaseID)
		assert.NoError(suite.T(), err)
		assert.Empty(suite.T(), items)
	}
}

func (suite *CategoryAssessmentServiceTestSuite) TestCreateOrUpdateRejectsParticipantOutsidePhase() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	req := categoryAssessmentDTO.CreateOrUpdateCategoryAssessmentRequest{
		CategoryID:            uuid.MustParse("25f1c984-ba31-4cf2-aa8e-5662721bf44e"),
		CourseParticipationID: outsideParticipationID,
		Comment:               "Comment on a student outside the phase",
		Author:                "Test Author",
		AuthorID:              "test-author-id",
	}

	err := suite.service.CreateOrUpdateCategoryAssessment(suite.suiteCtx, "", coursePhaseID, req)
	assert.ErrorIs(suite.T(), err, coursePhaseConfig.ErrParticipantNotInPhase)

	items, err := suite.service.ListCategoryAssessmentsByStudentInPhase(suite.suiteCtx, outsideParticipationID, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), items)
}

func TestCategoryAssessmentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CategoryAssessmentServiceTestSuite))
}
