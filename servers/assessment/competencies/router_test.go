package competencies

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
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/competencies/competencyDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CompetencyRouterTestSuite struct {
	suite.Suite
	router            *gin.Engine
	ctx               context.Context
	cleanup           func()
	mockCoreCleanup   func()
	competencyService CompetencyService
}

const testCoursePhaseID = "4179d58a-d00d-4fa7-94a5-397bc69fab02"

func (suite *CompetencyRouterTestSuite) SetupTest() {
	suite.ctx = context.Background()

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/categories.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup

	// Set up mock core service
	_, mockCleanup := testutils.SetupMockCoreService()
	suite.mockCoreCleanup = mockCleanup

	suite.competencyService = CompetencyService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}

	CompetencyServiceSingleton = &suite.competencyService

	// Initialize service modules needed for schema copy logic
	group := gin.New().Group("")
	assessmentSchemas.InitAssessmentSchemaModule(group, *testDB.Queries, testDB.Conn)
	coursePhaseConfig.InitCoursePhaseConfigModule(group, *testDB.Queries, testDB.Conn)

	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	testMiddleWare := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "existingstudent@example.com", "03711111", "ab12cde")
	}
	setupCompetencyRouter(api, testMiddleWare)
}

func (suite *CompetencyRouterTestSuite) TearDownTest() {
	if suite.mockCoreCleanup != nil {
		suite.mockCoreCleanup()
	}
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *CompetencyRouterTestSuite) TestListCompetencies() {
	req, _ := http.NewRequest("GET", "/api/course_phase/"+testCoursePhaseID+"/competency", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var competencies []competencyDTO.Competency
	err := json.Unmarshal(resp.Body.Bytes(), &competencies)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(competencies), 0, "Should return a list of competencies")
}

func (suite *CompetencyRouterTestSuite) TestGetCompetency() {
	// First create a competency to get
	categoryID := uuid.MustParse("25f1c984-ba31-4cf2-aa8e-5662721bf44e")
	createReq := competencyDTO.CreateCompetencyRequest{
		CategoryID:          categoryID,
		Name:                "Test Competency",
		Description:         "Test Description",
		DescriptionVeryBad:  "Very Bad Description",
		DescriptionBad:      "Bad Description",
		DescriptionOk:       "Ok Description",
		DescriptionGood:     "Good Description",
		DescriptionVeryGood: "Very Good Description",
		Weight:              10,
	}

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+testCoursePhaseID+"/competency", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusCreated, resp.Code)

	// Get list of competencies to find the created one
	req, _ = http.NewRequest("GET", "/api/course_phase/"+testCoursePhaseID+"/competency", nil)
	resp = httptest.NewRecorder()
	suite.router.ServeHTTP(resp, req)

	var competencies []competencyDTO.Competency
	err := json.Unmarshal(resp.Body.Bytes(), &competencies)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(competencies), 0, "Should have at least one competency")

	// Find the created competency by unique attributes
	var competencyID uuid.UUID
	for _, competency := range competencies {
		if competency.Name == "Test Competency" { // Replace with actual unique attribute
			competencyID = competency.ID
			break
		}
	}
	assert.NotZero(suite.T(), competencyID, "Created competency not found in list")

	// Now test getting the specific competency
	req, _ = http.NewRequest("GET", "/api/course_phase/"+testCoursePhaseID+"/competency/"+competencyID.String(), nil)
	resp = httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var competency competencyDTO.Competency
	err = json.Unmarshal(resp.Body.Bytes(), &competency)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), competencyID, competency.ID)
}

func (suite *CompetencyRouterTestSuite) TestGetCompetencyInvalidID() {
	req, _ := http.NewRequest("GET", "/api/course_phase/"+testCoursePhaseID+"/competency/invalid-uuid", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)

	var errorResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorResp["error"], "invalid UUID")
}

func (suite *CompetencyRouterTestSuite) TestListCompetenciesByCategory() {
	categoryID := uuid.New()

	req, _ := http.NewRequest("GET", "/api/course_phase/"+testCoursePhaseID+"/competency/category/"+categoryID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, resp.Code)

	var errResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errResp, "error")
}

func (suite *CompetencyRouterTestSuite) TestListCompetenciesByCategoryInvalidID() {
	req, _ := http.NewRequest("GET", "/api/course_phase/"+testCoursePhaseID+"/competency/category/invalid-uuid", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)

	var errorResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorResp["error"], "invalid UUID")
}

func (suite *CompetencyRouterTestSuite) TestCreateCompetency() {
	categoryID := uuid.MustParse("25f1c984-ba31-4cf2-aa8e-5662721bf44e")
	createReq := competencyDTO.CreateCompetencyRequest{
		CategoryID:          categoryID,
		Name:                "Test Competency",
		Description:         "Test Description",
		DescriptionVeryBad:  "Very Bad Description",
		DescriptionBad:      "Bad Description",
		DescriptionOk:       "Ok Description",
		DescriptionGood:     "Good Description",
		DescriptionVeryGood: "Very Good Description",
		Weight:              10,
	}

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+testCoursePhaseID+"/competency", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
}

func (suite *CompetencyRouterTestSuite) TestCreateCompetencyInvalidJSON() {
	req, _ := http.NewRequest("POST", "/api/course_phase/"+testCoursePhaseID+"/competency", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)

	var errorResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorResp, "error")
}

func (suite *CompetencyRouterTestSuite) TestUpdateCompetency() {
	// First create a competency to update
	categoryID := uuid.MustParse("25f1c984-ba31-4cf2-aa8e-5662721bf44e")
	createReq := competencyDTO.CreateCompetencyRequest{
		CategoryID:          categoryID,
		Name:                "Original Competency",
		Description:         "Original Description",
		DescriptionVeryBad:  "Original Very Bad Description",
		DescriptionBad:      "Original Bad Description",
		DescriptionOk:       "Original Ok Description",
		DescriptionGood:     "Original Good Description",
		DescriptionVeryGood: "Original Very Good Description",
		Weight:              5,
	}

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+testCoursePhaseID+"/competency", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusCreated, resp.Code)

	// Get list of competencies to find the created one
	req, _ = http.NewRequest("GET", "/api/course_phase/"+testCoursePhaseID+"/competency", nil)
	resp = httptest.NewRecorder()
	suite.router.ServeHTTP(resp, req)

	var competencies []competencyDTO.Competency
	err := json.Unmarshal(resp.Body.Bytes(), &competencies)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(competencies), 0, "Should have at least one competency")

	// Find the created competency by unique attributes
	var competencyID uuid.UUID
	for _, competency := range competencies {
		if competency.Name == "Original Competency" { // Replace with actual unique attribute
			competencyID = competency.ID
			break
		}
	}
	assert.NotZero(suite.T(), competencyID, "Created competency not found in list")

	// Now update the competency
	updateReq := competencyDTO.UpdateCompetencyRequest{
		CategoryID:          categoryID,
		Name:                "Updated Competency",
		Description:         "Updated Description",
		DescriptionVeryBad:  "Updated Very Bad Description",
		DescriptionBad:      "Updated Bad Description",
		DescriptionOk:       "Updated Ok Description",
		DescriptionGood:     "Updated Good Description",
		DescriptionVeryGood: "Updated Very Good Description",
		Weight:              15,
	}

	body, _ = json.Marshal(updateReq)
	req, _ = http.NewRequest("PUT", "/api/course_phase/"+testCoursePhaseID+"/competency/"+competencyID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *CompetencyRouterTestSuite) TestUpdateCompetencyInvalidID() {
	updateReq := competencyDTO.UpdateCompetencyRequest{
		CategoryID:          uuid.New(),
		Name:                "Updated Competency",
		Description:         "Updated Description",
		DescriptionVeryBad:  "Updated Very Bad Description",
		DescriptionBad:      "Updated Bad Description",
		DescriptionOk:       "Updated Ok Description",
		DescriptionGood:     "Updated Good Description",
		DescriptionVeryGood: "Updated Very Good Description",
		Weight:              15,
	}

	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/course_phase/"+testCoursePhaseID+"/competency/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)

	var errorResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorResp["error"], "invalid UUID")
}

func (suite *CompetencyRouterTestSuite) TestUpdateCompetencyInvalidJSON() {
	competencyID := uuid.New()

	req, _ := http.NewRequest("PUT", "/api/course_phase/"+testCoursePhaseID+"/competency/"+competencyID.String(), bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)

	var errorResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorResp, "error")
}

func (suite *CompetencyRouterTestSuite) TestDeleteCompetency() {
	// First create a competency to delete
	categoryID := uuid.MustParse("25f1c984-ba31-4cf2-aa8e-5662721bf44e")
	createReq := competencyDTO.CreateCompetencyRequest{
		CategoryID:          categoryID,
		Name:                "Competency to Delete",
		Description:         "Description",
		DescriptionVeryBad:  "Very Bad Description",
		DescriptionBad:      "Bad Description",
		DescriptionOk:       "Ok Description",
		DescriptionGood:     "Good Description",
		DescriptionVeryGood: "Very Good Description",
		Weight:              10,
	}

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+testCoursePhaseID+"/competency", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusCreated, resp.Code)

	// Get list of competencies to find the created one
	req, _ = http.NewRequest("GET", "/api/course_phase/"+testCoursePhaseID+"/competency", nil)
	resp = httptest.NewRecorder()
	suite.router.ServeHTTP(resp, req)

	var competencies []competencyDTO.Competency
	err := json.Unmarshal(resp.Body.Bytes(), &competencies)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(competencies), 0, "Should have at least one competency")

	// Find the created competency by unique attributes
	var competencyID uuid.UUID
	for _, competency := range competencies {
		if competency.Name == "Competency to Delete" { // Replace with actual unique attribute
			competencyID = competency.ID
			break
		}
	}
	assert.NotZero(suite.T(), competencyID, "Created competency not found in list")

	// Now delete the competency
	req, _ = http.NewRequest("DELETE", "/api/course_phase/"+testCoursePhaseID+"/competency/"+competencyID.String(), nil)
	resp = httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *CompetencyRouterTestSuite) TestDeleteCompetencyInvalidID() {
	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+testCoursePhaseID+"/competency/invalid-uuid", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)

	var errorResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorResp["error"], "invalid UUID")
}

func TestCompetencyRouterTestSuite(t *testing.T) {
	suite.Run(t, new(CompetencyRouterTestSuite))
}
