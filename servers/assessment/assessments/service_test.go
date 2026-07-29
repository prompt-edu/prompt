package assessments

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/categoryAssessment"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/utils"
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
	assessmentCompletion.InitAssessmentCompletionModule(gin.New().Group("/dummy"), *testDB.Queries, testDB.Conn)
	scoreLevel.InitScoreLevelModule(gin.New().Group("/dummy"), *testDB.Queries, testDB.Conn)
	categoryAssessment.InitCategoryAssessmentModule(gin.New().Group("/dummy"), *testDB.Queries, testDB.Conn)

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
	coursePhaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	err := DeleteAssessment(suite.suiteCtx, id, coursePhaseID)
	assert.ErrorIs(suite.T(), err, ErrAssessmentNotFound, "Deleting non-existent assessment should return not-found")
}

func (suite *AssessmentServiceTestSuite) TestCreateOrUpdateAssessmentWithEmptyScoreLevel() {
	req := assessmentDTO.CreateOrUpdateAssessmentRequest{
		CourseParticipationID: uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7"),
		CoursePhaseID:         uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9"),
		CompetencyID:          uuid.MustParse("01935143-5e85-7e1d-81bb-96fb3ebf34aa"),
		ScoreLevel:            "",
		Author:                "Test Author",
	}

	err := CreateOrUpdateAssessment(suite.suiteCtx, req)
	assert.Error(suite.T(), err, "Expected error for empty scoreLevel")
	assert.Equal(suite.T(), ErrInvalidScoreLevel, err, "Expected ErrInvalidScoreLevel")
}

func (suite *AssessmentServiceTestSuite) TestExportStudentAssessmentJSONIncludesAssessmentsCompletionAndActionItems() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")

	err := suite.service.queries.CreateOrUpdateAssessmentCompletion(suite.suiteCtx, db.CreateOrUpdateAssessmentCompletionParams{
		CourseParticipationID: partID,
		CoursePhaseID:         phaseID,
		CompletedAt:           pgtype.Timestamptz{Time: time.Date(2025, 5, 21, 10, 0, 0, 0, time.UTC), Valid: true},
		Author:                "Export Test Author",
		Comment:               "Export general remarks",
		GradeSuggestion:       utils.MapFloat64ToNumeric(4.5),
		Completed:             true,
	})
	assert.NoError(suite.T(), err)

	err = suite.service.queries.CreateActionItem(suite.suiteCtx, db.CreateActionItemParams{
		ID:                    uuid.New(),
		CoursePhaseID:         phaseID,
		CourseParticipationID: partID,
		Action:                "Export action item",
		Author:                "Export Test Author",
	})
	assert.NoError(suite.T(), err)

	export, err := ExportStudentAssessment(suite.suiteCtx, phaseID, partID, AssessmentExportFormatJSON)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), phaseID, export.CoursePhaseID)
	assert.Equal(suite.T(), partID, export.CourseParticipationID)
	assert.True(suite.T(), export.StudentAssessment.AssessmentCompletion.Completed)
	assert.Equal(suite.T(), "Export general remarks", export.StudentAssessment.AssessmentCompletion.Comment)
	assert.Equal(suite.T(), 4.5, export.StudentAssessment.AssessmentCompletion.GradeSuggestion)
	assert.True(suite.T(), export.StudentAssessment.AssessmentCompletion.CompletedAt.Valid)
	assert.NotEmpty(suite.T(), export.StudentAssessment.Assessments)
	assert.NotEmpty(suite.T(), export.ActionItems)
	foundExportActionItem := false
	for _, item := range export.ActionItems {
		if strings.Contains(item.Action, "Export action item") {
			foundExportActionItem = true
			break
		}
	}
	assert.True(suite.T(), foundExportActionItem)
}

func (suite *AssessmentServiceTestSuite) TestExportStudentAssessmentJSONWithoutOptionalData() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.New()

	export, err := ExportStudentAssessment(suite.suiteCtx, phaseID, partID, AssessmentExportFormatJSON)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), phaseID, export.CoursePhaseID)
	assert.Equal(suite.T(), partID, export.CourseParticipationID)
	assert.False(suite.T(), export.StudentAssessment.AssessmentCompletion.Completed)
	assert.False(suite.T(), export.StudentAssessment.AssessmentCompletion.CompletedAt.Valid)
	assert.Empty(suite.T(), export.StudentAssessment.AssessmentCompletion.Comment)
	assert.Empty(suite.T(), export.StudentAssessment.Assessments)
	assert.Empty(suite.T(), export.StudentAssessment.CategoryAssessments)
	assert.Empty(suite.T(), export.ActionItems)
}

func (suite *AssessmentServiceTestSuite) TestExportStudentAssessmentUnsupportedFormat() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")

	_, err := ExportStudentAssessment(suite.suiteCtx, phaseID, partID, "pdf")
	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, ErrUnsupportedAssessmentExportFormat))
}

func TestAssessmentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AssessmentServiceTestSuite))
}
