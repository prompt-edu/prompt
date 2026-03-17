package coursePhaseParticipation

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseParticipation/coursePhaseParticipationDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RouterTestSuite struct {
	suite.Suite
	router                          *gin.Engine
	ctx                             context.Context
	cleanup                         func()
	coursePhaseParticipationService CoursePhaseParticipationService
}

func (suite *RouterTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Set up PostgreSQL container
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../../database_dumps/full_db.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}

	suite.cleanup = cleanup
	suite.coursePhaseParticipationService = CoursePhaseParticipationService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	CoursePhaseParticipationServiceSingleton = &suite.coursePhaseParticipationService

	resolution.InitResolutionModule("localhost:8080")

	suite.router = setupRouter()
}

func (suite *RouterTestSuite) TearDownSuite() {
	suite.cleanup()
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	setupCoursePhaseParticipationRouter(api, func() gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail([]string{"PROMPT_Admin", "ios24245-iPraktikum-Lecturer"}, "existingstudent@example.com", "1234567", "ab12cde")
	}, sdkTestUtils.MockPermissionMiddleware)
	return router
}

func (suite *RouterTestSuite) TestGetParticipationsForCoursePhase() {
	req := httptest.NewRequest(http.MethodGet, "/api/course_phases/4e736d05-c125-48f0-8fa0-848b03ca6908/participations", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var response coursePhaseParticipationDTO.CoursePhaseParticipationsWithResolutions
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(response.Participations), 0, "Expected participations to be returned")
}

func (suite *RouterTestSuite) TestUpdateCoursePhaseParticipation() {
	jsonData := `{"other-value": "some skills"}`
	var data meta.MetaData
	err := json.Unmarshal([]byte(jsonData), &data)
	assert.NoError(suite.T(), err)
	pass := db.PassStatusPassed

	updatedParticipation := coursePhaseParticipationDTO.UpdateCoursePhaseParticipationRequest{
		CourseParticipationID: uuid.MustParse("6a49b717-a8ca-4d16-bcd0-0bb059525269"),
		RestrictedData:        data,
		StudentReadableData:   data,
		PassStatus:            &pass,
	}
	body, _ := json.Marshal(updatedParticipation)

	// Send the update request
	req := httptest.NewRequest(http.MethodPut, "/api/course_phases/4e736d05-c125-48f0-8fa0-848b03ca6908/participations/6a49b717-a8ca-4d16-bcd0-0bb059525269", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assert the update request was successful
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	// Perform a GET request to verify the changes
	getReq := httptest.NewRequest(http.MethodGet, "/api/course_phases/4e736d05-c125-48f0-8fa0-848b03ca6908/participations/6a49b717-a8ca-4d16-bcd0-0bb059525269", nil)
	getW := httptest.NewRecorder()

	suite.router.ServeHTTP(getW, getReq)

	// Assert the GET request was successful
	assert.Equal(suite.T(), http.StatusOK, getW.Code)

	// Verify the returned data matches the expected updated data
	var updated coursePhaseParticipationDTO.CoursePhaseParticipationWithResolution
	err = json.Unmarshal(getW.Body.Bytes(), &updated)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), updatedParticipation.CourseParticipationID, updated.Participation.CourseParticipationID, "Course Participation ID should match")
	assert.Equal(suite.T(), "passed", updated.Participation.PassStatus, "PassStatus should match")
	assert.Equal(suite.T(), updatedParticipation.RestrictedData["other-value"], updated.Participation.RestrictedData["other-value"], "New Meta data should match")
	assert.Equal(suite.T(), updatedParticipation.StudentReadableData["other-value"], updated.Participation.StudentReadableData["other-value"], "New Meta data should match")

}

func (suite *RouterTestSuite) TestUpdateNewCoursePhaseParticipation() {
	jsonData := `{"other-value": "some skills"}`
	var data meta.MetaData
	err := json.Unmarshal([]byte(jsonData), &data)
	assert.NoError(suite.T(), err)
	pass := db.PassStatusPassed

	toBeCreatedParticipation := coursePhaseParticipationDTO.UpdateCoursePhaseParticipationRequest{
		CourseParticipationID: uuid.MustParse("f6744410-cfe2-456d-96fa-e857cf989569"),
		RestrictedData:        data,
		StudentReadableData:   data,
		PassStatus:            &pass,
	}
	body, _ := json.Marshal(toBeCreatedParticipation)

	// Send the update request
	req := httptest.NewRequest(http.MethodPut, "/api/course_phases/4e736d05-c125-48f0-8fa0-848b03ca6908/participations/f6744410-cfe2-456d-96fa-e857cf989569", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assert the update request was successful
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	// Perform a GET request to verify the changes
	getReq := httptest.NewRequest(http.MethodGet, "/api/course_phases/4e736d05-c125-48f0-8fa0-848b03ca6908/participations/f6744410-cfe2-456d-96fa-e857cf989569", nil)
	getW := httptest.NewRecorder()

	suite.router.ServeHTTP(getW, getReq)

	// Assert the GET request was successful
	assert.Equal(suite.T(), http.StatusOK, getW.Code)

	// Verify the returned data matches the expected updated data
	var createdParticipation coursePhaseParticipationDTO.CoursePhaseParticipationWithResolution
	err = json.Unmarshal(getW.Body.Bytes(), &createdParticipation)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "passed", createdParticipation.Participation.PassStatus, "PassStatus should match")
	assert.Equal(suite.T(), toBeCreatedParticipation.RestrictedData["other-value"], createdParticipation.Participation.RestrictedData["other-value"], "New Meta data should match")
	assert.Equal(suite.T(), toBeCreatedParticipation.StudentReadableData["other-value"], createdParticipation.Participation.StudentReadableData["other-value"], "New Meta data should match")
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}
