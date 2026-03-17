package scoreLevel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	scoreLevelDTO "github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

// router reference from package setup
// filepath: assessments/scoreLevel/router.go

// ScoreLevelRouterTestSuite tests the HTTP endpoints for scoreLevel
type ScoreLevelRouterTestSuite struct {
	suite.Suite
	router   *gin.Engine
	suiteCtx context.Context
	cleanup  func()
	service  ScoreLevelService
}

func (suite *ScoreLevelRouterTestSuite) SetupSuite() {
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

	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "user@example.com", "1234", "id")
	}
	// attach routes
	setupScoreLevelRouter(api, testMiddleware)
}

func (suite *ScoreLevelRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *ScoreLevelRouterTestSuite) TestGetAllScoreLevels() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/scoreLevel", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var levels []scoreLevelDTO.ScoreLevelWithParticipation
	err := json.Unmarshal(resp.Body.Bytes(), &levels)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(levels), 0, "Expected at least one score level")
	for _, lvl := range levels {
		assert.NotEmpty(suite.T(), lvl.CourseParticipationID)
		assert.NotEmpty(suite.T(), string(lvl.ScoreLevel))
	}
}

func (suite *ScoreLevelRouterTestSuite) TestGetScoreLevelByCourseParticipationID() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")
	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/scoreLevel/"+partID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var lvl db.ScoreLevel
	err := json.Unmarshal(resp.Body.Bytes(), &lvl)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), db.ScoreLevelVeryBad, lvl)
}

func (suite *ScoreLevelRouterTestSuite) TestGetScoreLevelInvalidParticipationID() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/scoreLevel/invalid-uuid", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *ScoreLevelRouterTestSuite) TestGetScoreLevelInvalidPhaseID() {
	req, _ := http.NewRequest("GET", "/api/course_phase/invalid-phase/student-assessment/scoreLevel/"+uuid.New().String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func TestScoreLevelRouterTestSuite(t *testing.T) {
	suite.Run(t, new(ScoreLevelRouterTestSuite))
}
