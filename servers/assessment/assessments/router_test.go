package assessments

import (
	"bytes"
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
	assessmentDTO "github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

// AssessmentRouterTestSuite tests HTTP routes for assessments
type AssessmentRouterTestSuite struct {
	suite.Suite
	router   *gin.Engine
	suiteCtx context.Context
	cleanup  func()
	service  AssessmentService
}

func (suite *AssessmentRouterTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/assessments.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test DB: %v", err)
	}
	suite.cleanup = cleanup

	// initialize service singleton
	suite.service = AssessmentService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	AssessmentServiceSingleton = &suite.service

	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "user@example.com", "id@example.com", "id")
	}
	// attach assessment routes
	setupAssessmentRouter(api, testMiddleware)
}

func (suite *AssessmentRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *AssessmentRouterTestSuite) TestListByCoursePhase() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	var items []assessmentDTO.Assessment
	err := json.Unmarshal(resp.Body.Bytes(), &items)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(items), 0)
}

func (suite *AssessmentRouterTestSuite) TestCreateInvalidJSON() {
	phaseID := uuid.New()
	// POST invalid JSON
	req, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/student-assessment", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *AssessmentRouterTestSuite) TestListByStudentInPhase() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")
	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/course-participation/"+partID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	var items []assessmentDTO.Assessment
	err := json.Unmarshal(resp.Body.Bytes(), &items)
	assert.NoError(suite.T(), err)
}

func (suite *AssessmentRouterTestSuite) TestInvalidUUIDs() {
	req, _ := http.NewRequest("GET", "/api/course_phase/invalid/student-assessment/invalid", nil)
	resp := httptest.NewRecorder()
	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func TestAssessmentRouterTestSuite(t *testing.T) {
	suite.Run(t, new(AssessmentRouterTestSuite))
}
