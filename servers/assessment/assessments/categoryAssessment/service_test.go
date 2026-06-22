package categoryAssessment

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/categoryAssessment/categoryAssessmentDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CategoryAssessmentServiceTestSuite struct {
	suite.Suite
	suiteCtx context.Context
	cleanup  func()
	service  CategoryAssessmentService
}

func (suite *CategoryAssessmentServiceTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../../database_dumps/assessments.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.service = CategoryAssessmentService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	CategoryAssessmentServiceSingleton = &suite.service
	coursePhaseConfig.CoursePhaseConfigSingleton = coursePhaseConfig.NewCoursePhaseConfigService(*testDB.Queries, testDB.Conn)
}

func (suite *CategoryAssessmentServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *CategoryAssessmentServiceTestSuite) TestCreateOrUpdateCategoryAssessment() {
	req := categoryAssessmentDTO.CreateOrUpdateCategoryAssessmentRequest{
		CategoryID:            uuid.MustParse("25f1c984-ba31-4cf2-aa8e-5662721bf44e"),
		CoursePhaseID:         uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02"),
		CourseParticipationID: uuid.New(),
		Comment:               "Strong understanding of version control workflows.",
		Author:                "Test Author",
		AuthorID:              "test-author-id",
	}

	err := CreateOrUpdateCategoryAssessment(suite.suiteCtx, req)
	assert.NoError(suite.T(), err)

	items, err := ListCategoryAssessmentsByStudentInPhase(suite.suiteCtx, req.CourseParticipationID, req.CoursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), items, 1)
	assert.Equal(suite.T(), req.CategoryID, items[0].CategoryID)
	assert.Equal(suite.T(), req.Comment, items[0].Comment)
	assert.Equal(suite.T(), req.Author, items[0].Author)
	assert.Equal(suite.T(), req.AuthorID, items[0].AuthorID)
}

func (suite *CategoryAssessmentServiceTestSuite) TestListCategoryAssessmentsByStudentInPhase() {
	req := categoryAssessmentDTO.CreateOrUpdateCategoryAssessmentRequest{
		CategoryID:            uuid.MustParse("815b159b-cab3-49b4-8060-c4722d59241d"),
		CoursePhaseID:         uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02"),
		CourseParticipationID: uuid.New(),
		Comment:               "Clear UI reasoning and implementation notes.",
		Author:                "Test Author",
		AuthorID:              "test-author-id",
	}

	err := CreateOrUpdateCategoryAssessment(suite.suiteCtx, req)
	assert.NoError(suite.T(), err)

	items, err := ListCategoryAssessmentsByStudentInPhase(suite.suiteCtx, req.CourseParticipationID, req.CoursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), items, 1)
	assert.Equal(suite.T(), req.CourseParticipationID, items[0].CourseParticipationID)
	assert.Equal(suite.T(), req.CoursePhaseID, items[0].CoursePhaseID)
	assert.Equal(suite.T(), req.CategoryID, items[0].CategoryID)
}

func TestCategoryAssessmentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CategoryAssessmentServiceTestSuite))
}
