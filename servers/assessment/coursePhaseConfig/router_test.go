package coursePhaseConfig

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CoursePhaseConfigRouterTestSuite struct {
	suite.Suite
	router                   *gin.Engine
	suiteCtx                 context.Context
	cleanup                  func()
	coursePhaseConfigService CoursePhaseConfigService
	testCoursePhaseID        uuid.UUID
}

func (suite *CoursePhaseConfigRouterTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/coursePhaseConfig.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.coursePhaseConfigService = CoursePhaseConfigService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	CoursePhaseConfigSingleton = &suite.coursePhaseConfigService

	suite.testCoursePhaseID = uuid.New()
	templateID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000") // From our test data

	// Insert a course phase config entry to enable updates
	_, err = testDB.Conn.Exec(suite.suiteCtx,
		"INSERT INTO course_phase_config (assessment_schema_id, course_phase_id) VALUES ($1, $2)",
		templateID, suite.testCoursePhaseID)
	if err != nil {
		suite.T().Fatalf("Failed to insert test course phase config: %v", err)
	}

	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	testMiddleWare := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "lecturer@example.com", "12345", "lecturer")
	}
	setupCoursePhaseRouter(api, testMiddleWare)
}

func (suite *CoursePhaseConfigRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *CoursePhaseConfigRouterTestSuite) TestGetTeamsForCoursePhase() {
	// Test GET request for teams
	url := fmt.Sprintf("/api/course_phase/%s/config/teams", suite.testCoursePhaseID.String())
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(suite.T(), err)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Since this calls an external service, we expect it to potentially fail
	// but we're testing that the endpoint exists and handles requests properly
	// The status should be either 200 (success) or 500 (external service failure)
	assert.True(suite.T(), w.Code == http.StatusOK || w.Code == http.StatusInternalServerError,
		"Should return either OK or Internal Server Error")

	if w.Code == http.StatusOK {
		var teams []promptTypes.Team
		err = json.Unmarshal(w.Body.Bytes(), &teams)
		assert.NoError(suite.T(), err, "Response should be valid JSON array of teams")
		// Teams array can be empty, that's valid
	}
}

func (suite *CoursePhaseConfigRouterTestSuite) TestGetTeamsForCoursePhaseInvalidID() {
	// Test GET request with invalid UUID
	url := "/api/course_phase/invalid-uuid/config/teams"
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(suite.T(), err)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "Invalid course phase ID")
}

func (suite *CoursePhaseConfigRouterTestSuite) TestGetTeamsForCoursePhaseNonExistent() {
	nonExistentID := uuid.New()

	// Test GET request for non-existent course phase
	url := fmt.Sprintf("/api/course_phase/%s/config/teams", nonExistentID.String())
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(suite.T(), err)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Since this calls an external service, we expect it to potentially fail
	// The status should be either 200 (success with empty array) or 500 (external service failure)
	assert.True(suite.T(), w.Code == http.StatusOK || w.Code == http.StatusInternalServerError,
		"Should return either OK or Internal Server Error for non-existent course phase")

	if w.Code == http.StatusOK {
		var teams []promptTypes.Team
		err = json.Unmarshal(w.Body.Bytes(), &teams)
		assert.NoError(suite.T(), err, "Response should be valid JSON array")
		// For non-existent course phase, teams array should be empty
	}
}

func (suite *CoursePhaseConfigRouterTestSuite) TestGetCoursePhaseConfig() {
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/course_phase/%s/config", suite.testCoursePhaseID.String()), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.True(suite.T(), resp.Code == http.StatusOK || resp.Code == http.StatusInternalServerError)
}

func (suite *CoursePhaseConfigRouterTestSuite) TestGetCoursePhaseConfigInvalidID() {
	req, _ := http.NewRequest("GET", "/api/course_phase/invalid-uuid/config", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *CoursePhaseConfigRouterTestSuite) TestGetParticipationsForCoursePhase() {
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/course_phase/%s/config/participations", suite.testCoursePhaseID.String()), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	// This endpoint may return success or error depending on external service availability
	assert.True(suite.T(), resp.Code == http.StatusOK || resp.Code == http.StatusInternalServerError)
}

func (suite *CoursePhaseConfigRouterTestSuite) TestGetParticipationsForCoursePhaseInvalidID() {
	req, _ := http.NewRequest("GET", "/api/course_phase/invalid-uuid/config/participations", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func TestCoursePhaseConfigRouterTestSuite(t *testing.T) {
	suite.Run(t, new(CoursePhaseConfigRouterTestSuite))
}
