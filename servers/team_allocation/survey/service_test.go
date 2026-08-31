package survey

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SurveyServiceTestSuite struct {
	suite.Suite
	suiteCtx      context.Context
	cleanup       func()
	surveyService *SurveyService
}

func (suite *SurveyServiceTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/complete_schema.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.surveyService = NewSurveyService(*testDB.Queries, testDB.Conn)
}

func (suite *SurveyServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *SurveyServiceTestSuite) TestGetSurveyForm() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")

	surveyForm, err := suite.surveyService.GetSurveyForm(suite.suiteCtx, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(surveyForm.Teams), 0, "Should return teams")
	assert.GreaterOrEqual(suite.T(), len(surveyForm.Skills), 0, "Should return skills")
	assert.NotZero(suite.T(), surveyForm.Deadline, "Should have a deadline")
}

func (suite *SurveyServiceTestSuite) TestGetSurveyFormNonExistentCoursePhase() {
	nonExistentID := uuid.New()

	_, err := suite.surveyService.GetSurveyForm(suite.suiteCtx, nonExistentID)
	assert.Error(suite.T(), err, "Should error for non-existent course phase")
}

func (suite *SurveyServiceTestSuite) TestGetSurveyTimeframe() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")

	timeframe, err := suite.surveyService.GetSurveyTimeframe(suite.suiteCtx, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), timeframe.TimeframeSet, "Timeframe should be set")
	assert.NotZero(suite.T(), timeframe.SurveyStart, "Should have a start time")
	assert.NotZero(suite.T(), timeframe.SurveyDeadline, "Should have a deadline")
}

func (suite *SurveyServiceTestSuite) TestGetSurveyTimeframeNonExistent() {
	nonExistentID := uuid.New()

	timeframe, err := suite.surveyService.GetSurveyTimeframe(suite.suiteCtx, nonExistentID)
	assert.NoError(suite.T(), err, "Should not error for non-existent timeframe")
	assert.False(suite.T(), timeframe.TimeframeSet, "Timeframe should not be set")
}

func (suite *SurveyServiceTestSuite) TestSetSurveyTimeframePreservesInstantForNonUTCZone() {
	berlin, err := time.LoadLocation("Europe/Berlin")
	assert.NoError(suite.T(), err)

	coursePhaseID := uuid.New()
	start := time.Date(2026, 1, 15, 12, 34, 56, 123456000, berlin)
	deadline := start.Add(48 * time.Hour)

	err = SetSurveyTimeframe(suite.suiteCtx, coursePhaseID, start, deadline)
	assert.NoError(suite.T(), err)

	timeframe, err := GetSurveyTimeframe(suite.suiteCtx, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), start.Equal(timeframe.SurveyStart))
	assert.True(suite.T(), deadline.Equal(timeframe.SurveyDeadline))
}

func TestSurveyServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SurveyServiceTestSuite))
}
