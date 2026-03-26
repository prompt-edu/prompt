package scoreLevel

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type ScoreLevelServiceTestSuite struct {
	suite.Suite
	suiteCtx context.Context
	cleanup  func()
	service  ScoreLevelService
}

func (suite *ScoreLevelServiceTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../../database_dumps/assessments.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.service = ScoreLevelService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	ScoreLevelServiceSingleton = &suite.service
}

func (suite *ScoreLevelServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *ScoreLevelServiceTestSuite) TestGetAllScoreLevels() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	levels, err := GetAllScoreLevels(suite.suiteCtx, phaseID)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(levels), 0, "Expected at least one score level")
	for _, lvl := range levels {
		assert.NotEmpty(suite.T(), lvl.CourseParticipationID, "Participation ID should not be empty")
		assert.NotEmpty(suite.T(), string(lvl.ScoreLevel), "Score level should not be empty")
	}
}

func (suite *ScoreLevelServiceTestSuite) TestGetScoreLevelByCourseParticipationID() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")
	lvl, err := GetScoreLevelByCourseParticipationID(suite.suiteCtx, partID, phaseID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), db.ScoreLevelVeryBad, lvl, "Expected very bad level")
}

func (suite *ScoreLevelServiceTestSuite) TestGetScoreLevelByCourseParticipationIDNotFound() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.New()
	_, err := GetScoreLevelByCourseParticipationID(suite.suiteCtx, partID, phaseID)
	assert.Error(suite.T(), err, "Expected error for non-existent participation")
}

func (suite *ScoreLevelServiceTestSuite) TestGetStudentScore() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.MustParse("e482ab63-c1c0-4943-9221-989b0c257559")
	score, err := GetStudentScore(suite.suiteCtx, partID, phaseID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), scoreLevelDTO.ScoreLevelBad, score.ScoreLevel, "Expected bad level")
	assert.GreaterOrEqual(suite.T(), score.ScoreNumeric.Float64, float64(1), "Score should be >= 1")
	assert.LessOrEqual(suite.T(), score.ScoreNumeric.Float64, float64(5), "Score should be <= 5")
}

func (suite *ScoreLevelServiceTestSuite) TestMapDBScoreLevelToDTOIntegration() {
	// Test that our DTO mapping works correctly within the service context
	validLevel := scoreLevelDTO.MapDBScoreLevelToDTO("bad")
	assert.Equal(suite.T(), scoreLevelDTO.ScoreLevelBad, validLevel, "Valid score level should map correctly")

	invalidLevel := scoreLevelDTO.MapDBScoreLevelToDTO("invalid_level")
	assert.Equal(suite.T(), scoreLevelDTO.ScoreLevelVeryBad, invalidLevel, "Invalid score level should return ScoreLevelVeryBad")
}

func TestScoreLevelServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ScoreLevelServiceTestSuite))
}
