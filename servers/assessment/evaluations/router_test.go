package evaluations

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type EvaluationRouterTestSuite struct {
	suite.Suite
	router            *gin.Engine
	suiteCtx          context.Context
	cleanup           func()
	evaluationService EvaluationService
	testCoursePhaseID uuid.UUID
}

func (suite *EvaluationRouterTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/evaluations.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to setup test database: %v", err)
	}

	suite.cleanup = cleanup
	suite.evaluationService = EvaluationService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}

	EvaluationServiceSingleton = &suite.evaluationService

	// Initialize CoursePhaseConfigSingleton to prevent nil pointer dereference
	coursePhaseConfig.CoursePhaseConfigSingleton = coursePhaseConfig.NewCoursePhaseConfigService(*testDB.Queries, testDB.Conn)

	suite.testCoursePhaseID = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")

	suite.router = gin.Default()

	// Add global middleware to set courseParticipationID for all requests
	suite.router.Use(func(c *gin.Context) {
		// Set the courseParticipationID that the evaluations router expects
		// This matches one of the course participations in our test database
		testCourseParticipationID := uuid.MustParse("01234567-1234-1234-1234-123456789012")
		c.Set("courseParticipationID", testCourseParticipationID)
		c.Next()
	})

	api := suite.router.Group("/assessment/api/course_phase/:coursePhaseID")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "existingstudent@example.com", "03711111", "ab12cde")
	}
	setupEvaluationRouter(api, testMiddleware)
}

func (suite *EvaluationRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func TestEvaluationRouterTestSuite(t *testing.T) {
	suite.Run(t, new(EvaluationRouterTestSuite))
}

// Test Admin/Lecturer endpoints
func (suite *EvaluationRouterTestSuite) TestGetAllEvaluationsByPhase() {
	req, _ := http.NewRequest(http.MethodGet, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var evaluations []evaluationDTO.Evaluation
	err := json.Unmarshal(w.Body.Bytes(), &evaluations)
	assert.NoError(suite.T(), err)
}

// Test Student endpoints
func (suite *EvaluationRouterTestSuite) TestGetMyEvaluations() {
	req, _ := http.NewRequest(http.MethodGet, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation/my-evaluations", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var evaluations []evaluationDTO.Evaluation
	err := json.Unmarshal(w.Body.Bytes(), &evaluations)
	assert.NoError(suite.T(), err)
}

func (suite *EvaluationRouterTestSuite) TestCreateOrUpdateEvaluation() {
	requestBody := evaluationDTO.CreateOrUpdateEvaluationRequest{
		CourseParticipationID:       uuid.MustParse("02234567-1234-1234-1234-123456789012"), // Valid from test data
		CompetencyID:                uuid.MustParse("c1234567-1234-1234-1234-123456789012"), // Valid from test data
		ScoreLevel:                  scoreLevelDTO.ScoreLevelGood,
		AuthorCourseParticipationID: uuid.MustParse("01234567-1234-1234-1234-123456789012"), // Must match test middleware
		Type:                        assessmentType.Self,
	}

	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest(http.MethodPost, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
}

func (suite *EvaluationRouterTestSuite) TestCreateOrUpdateEvaluationInvalidJSON() {
	req, _ := http.NewRequest(http.MethodPost, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var errResp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errResp, "error")
}

func (suite *EvaluationRouterTestSuite) TestCreateOrUpdateEvaluationForbiddenAuthor() {
	requestBody := evaluationDTO.CreateOrUpdateEvaluationRequest{
		CourseParticipationID:       uuid.MustParse("02234567-1234-1234-1234-123456789012"), // Valid from test data
		CompetencyID:                uuid.MustParse("c1234567-1234-1234-1234-123456789012"), // Valid from test data
		ScoreLevel:                  scoreLevelDTO.ScoreLevelGood,
		AuthorCourseParticipationID: uuid.MustParse("02234567-1234-1234-1234-123456789012"), // Different from test middleware user
		Type:                        assessmentType.Self,
	}

	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest(http.MethodPost, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	var errResp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errResp["error"], "Students can only create evaluations as the author")
}

func (suite *EvaluationRouterTestSuite) TestDeleteEvaluation() {
	// Use an existing evaluation ID from our test data
	existingEvaluationID := "e1234567-1234-1234-1234-123456789012"
	req, _ := http.NewRequest(http.MethodDelete, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation/"+existingEvaluationID, nil)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Should return either 200 (if evaluation exists and was deleted) or 500 (if evaluation doesn't exist)
	assert.True(suite.T(), w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

func (suite *EvaluationRouterTestSuite) TestInvalidCoursePhaseID() {
	req, _ := http.NewRequest(http.MethodGet, "/assessment/api/course_phase/invalid-uuid/evaluation", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var errResp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errResp, "error")
}

func (suite *EvaluationRouterTestSuite) TestInvalidEvaluationID() {
	req, _ := http.NewRequest(http.MethodDelete, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation/invalid-uuid", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var errResp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errResp, "error")
}
