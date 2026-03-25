package categories

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
	"github.com/prompt-edu/prompt/servers/assessment/categories/categoryDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CategoryRouterTestSuite struct {
	suite.Suite
	router          *gin.Engine
	suiteCtx        context.Context
	cleanup         func()
	mockCoreCleanup func()
	categoryService CategoryService
}

func (suite *CategoryRouterTestSuite) SetupTest() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/categories.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup

	// Set up mock core service
	_, mockCleanup := testutils.SetupMockCoreService()
	suite.mockCoreCleanup = mockCleanup

	suite.categoryService = CategoryService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	CategoryServiceSingleton = &suite.categoryService

	// Initialize service modules needed for schema copy logic
	group := gin.New().Group("")
	assessmentSchemas.InitAssessmentSchemaModule(group, *testDB.Queries, testDB.Conn)
	coursePhaseConfig.InitCoursePhaseConfigModule(group, *testDB.Queries, testDB.Conn)

	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	testMiddleWare := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "existingstudent@example.com", "03711111", "ab12cde")
	}
	setupCategoryRouter(api, testMiddleWare)
}

func (suite *CategoryRouterTestSuite) TearDownTest() {
	if suite.mockCoreCleanup != nil {
		suite.mockCoreCleanup()
	}
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *CategoryRouterTestSuite) TestGetAllCategories() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02" // Dev Application phase from test data
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/category", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var cats []categoryDTO.Category
	err := json.Unmarshal(resp.Body.Bytes(), &cats)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(cats), 0, "Should return a list of categories")
}

func (suite *CategoryRouterTestSuite) TestCreateCategory() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02" // Dev Application phase from test data

	createReq := categoryDTO.CreateCategoryRequest{
		Name:               "Router Test Category",
		ShortName:          "RTC",
		Description:        "Testing create via router",
		Weight:             3,
		AssessmentSchemaID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), // From test data
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+coursePhaseID+"/category", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
}

func (suite *CategoryRouterTestSuite) TestCreateCategoryInvalidJSON() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02" // Dev Application phase from test data
	req, _ := http.NewRequest("POST", "/api/course_phase/"+coursePhaseID+"/category", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)

	var errResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errResp, "error")
}

func (suite *CategoryRouterTestSuite) TestUpdateCategory() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02" // Dev Application phase from test data
	id := uuid.MustParse("25f1c984-ba31-4cf2-aa8e-5662721bf44e")

	updateReq := categoryDTO.UpdateCategoryRequest{
		Name:               "Router Updated",
		ShortName:          "RU",
		Description:        "Router update description",
		Weight:             2,
		AssessmentSchemaID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), // From test data
	}
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/course_phase/"+coursePhaseID+"/category/"+id.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *CategoryRouterTestSuite) TestDeleteCategory() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02") // Dev Application phase from test data

	// create category to delete via service
	createReq := categoryDTO.CreateCategoryRequest{
		Name:               "RouterDelete",
		ShortName:          "RD",
		Description:        "To delete via router",
		Weight:             1,
		AssessmentSchemaID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), // From test data
	}
	err := CreateCategory(suite.suiteCtx, coursePhaseID, createReq)
	assert.NoError(suite.T(), err)

	// find created
	reqList, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID.String()+"/category", nil)
	respList := httptest.NewRecorder()
	suite.router.ServeHTTP(respList, reqList)
	var cats []categoryDTO.Category
	_ = json.Unmarshal(respList.Body.Bytes(), &cats)
	var delID string
	for _, c := range cats {
		if c.Name == createReq.Name {
			delID = c.ID.String()
			break
		}
	}

	reqDel, _ := http.NewRequest("DELETE", "/api/course_phase/"+coursePhaseID.String()+"/category/"+delID, nil)
	respDel := httptest.NewRecorder()
	suite.router.ServeHTTP(respDel, reqDel)
	assert.Equal(suite.T(), http.StatusOK, respDel.Code)
}

func (suite *CategoryRouterTestSuite) TestDeleteCategoryInvalidID() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02" // Dev Application phase from test data
	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+coursePhaseID+"/category/invalid-uuid", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *CategoryRouterTestSuite) TestGetCategoriesWithCompetencies() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02" // Dev Application phase from test data
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/category/assessment/with-competencies", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	var respBody []categoryDTO.CategoryWithCompetencies
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	assert.NoError(suite.T(), err, "Response should unmarshal properly")
}

func TestCategoryRouterTestSuite(t *testing.T) {
	suite.Run(t, new(CategoryRouterTestSuite))
}
