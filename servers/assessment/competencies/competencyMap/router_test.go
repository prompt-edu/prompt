package competencyMap

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
	"github.com/prompt-edu/prompt/servers/assessment/competencies/competencyMap/competencyMapDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type CompetencyMapRouterTestSuite struct {
	suite.Suite
	router        *gin.Engine
	suiteCtx      context.Context
	cleanup       func()
	service       CompetencyMapService
	coursePhaseID uuid.UUID
}

func (suite *CompetencyMapRouterTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../../database_dumps/competencyMaps.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup

	suite.service = CompetencyMapService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	CompetencyMapServiceSingleton = &suite.service
	suite.coursePhaseID = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")

	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	setupCompetencyMapRouter(api, testMiddleware)
}

func (suite *CompetencyMapRouterTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *CompetencyMapRouterTestSuite) TestCreateCompetencyMapping() {
	// Use existing competency IDs from the database dump
	fromCompetencyID := uuid.MustParse("c5234567-1234-1234-1234-123456789012") // Teamwork
	toCompetencyID := uuid.MustParse("c7234567-1234-1234-1234-123456789012")   // Mentoring

	req := competencyMapDTO.CompetencyMapping{
		FromCompetencyID: fromCompetencyID,
		ToCompetencyID:   toCompetencyID,
	}

	jsonData, err := json.Marshal(req)
	assert.NoError(suite.T(), err)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/api/course_phase/4179d58a-d00d-4fa7-94a5-397bc69fab02/competency-mappings", bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, httpReq)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Competency mapping created successfully", response["message"])
}

func (suite *CompetencyMapRouterTestSuite) TestDeleteCompetencyMapping() {
	// First create a mapping using existing competency IDs
	fromCompetencyID := uuid.MustParse("c6234567-1234-1234-1234-123456789012") // Project Management
	toCompetencyID := uuid.MustParse("c3234567-1234-1234-1234-123456789012")   // Testing

	createReq := competencyMapDTO.CompetencyMapping{
		FromCompetencyID: fromCompetencyID,
		ToCompetencyID:   toCompetencyID,
	}
	err := CreateCompetencyMapping(suite.suiteCtx, suite.coursePhaseID, createReq)
	assert.NoError(suite.T(), err)

	// Now delete it via API
	deleteReq := competencyMapDTO.CompetencyMapping{
		FromCompetencyID: fromCompetencyID,
		ToCompetencyID:   toCompetencyID,
	}

	jsonData, err := json.Marshal(deleteReq)
	assert.NoError(suite.T(), err)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", "/api/course_phase/4179d58a-d00d-4fa7-94a5-397bc69fab02/competency-mappings", bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, httpReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Competency mapping deleted successfully", response["message"])
}

func (suite *CompetencyMapRouterTestSuite) TestGetAllCompetencyMappings() {
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/course_phase/4179d58a-d00d-4fa7-94a5-397bc69fab02/competency-mappings", nil)

	suite.router.ServeHTTP(w, httpReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var mappings []competencyMapDTO.CompetencyMapping
	err := json.Unmarshal(w.Body.Bytes(), &mappings)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(mappings), 6) // We have 6 test mappings in the dump
}

func (suite *CompetencyMapRouterTestSuite) TestGetCompetencyMappings() {
	// Use the Programming Skills competency from our test data
	fromCompetencyID := "c1234567-1234-1234-1234-123456789012"

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/course_phase/4179d58a-d00d-4fa7-94a5-397bc69fab02/competency-mappings/from/"+fromCompetencyID, nil)

	suite.router.ServeHTTP(w, httpReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var mappings []competencyMapDTO.CompetencyMapping
	err := json.Unmarshal(w.Body.Bytes(), &mappings)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(mappings)) // Programming maps to System Design and Testing
}

func (suite *CompetencyMapRouterTestSuite) TestGetReverseCompetencyMappings() {
	// Use the System Design competency from our test data (mapped to by Programming)
	toCompetencyID := "c2234567-1234-1234-1234-123456789012"

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/course_phase/4179d58a-d00d-4fa7-94a5-397bc69fab02/competency-mappings/to/"+toCompetencyID, nil)

	suite.router.ServeHTTP(w, httpReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var mappings []competencyMapDTO.CompetencyMapping
	err := json.Unmarshal(w.Body.Bytes(), &mappings)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(mappings)) // Only Programming maps to System Design
}

func (suite *CompetencyMapRouterTestSuite) TestGetEvaluationsByMappedCompetency() {
	// Use the Programming Skills competency from our test data
	fromCompetencyID := "c1234567-1234-1234-1234-123456789012"

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/course_phase/4179d58a-d00d-4fa7-94a5-397bc69fab02/competency-mappings/evaluations/"+fromCompetencyID, nil)

	suite.router.ServeHTTP(w, httpReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// The response should contain evaluations for competencies mapped from Programming Skills
	var evaluations []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &evaluations)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(evaluations), 2) // Should have evaluations for mapped competencies
}

func (suite *CompetencyMapRouterTestSuite) TestCreateCompetencyMappingInvalidJSON() {
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/api/course_phase/4179d58a-d00d-4fa7-94a5-397bc69fab02/competency-mappings", bytes.NewBuffer([]byte("invalid json")))
	httpReq.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, httpReq)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *CompetencyMapRouterTestSuite) TestGetCompetencyMappingsInvalidUUID() {
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/course_phase/4179d58a-d00d-4fa7-94a5-397bc69fab02/competency-mappings/from/invalid-uuid", nil)

	suite.router.ServeHTTP(w, httpReq)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func TestCompetencyMapRouterTestSuite(t *testing.T) {
	suite.Run(t, new(CompetencyMapRouterTestSuite))
}
