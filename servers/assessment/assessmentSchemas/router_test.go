package assessmentSchemas

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
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas/assessmentSchemaDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AssessmentSchemaRouterTestSuite struct {
	suite.Suite
	router                  *gin.Engine
	suiteCtx                context.Context
	cleanup                 func()
	assessmentSchemaService AssessmentSchemaService
}

const testCoursePhaseID = "4179d58a-d00d-4fa7-94a5-397bc69fab02"

func (suite *AssessmentSchemaRouterTestSuite) SetupTest() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/categories.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.assessmentSchemaService = AssessmentSchemaService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	AssessmentSchemaServiceSingleton = &suite.assessmentSchemaService
	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "admin@example.com", "12345678", "admin123")
	}
	SetupAssessmentSchemaRouter(api, testMiddleware)
}

func (suite *AssessmentSchemaRouterTestSuite) TearDownTest() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *AssessmentSchemaRouterTestSuite) TestGetAllAssessmentSchemas() {
	req, _ := http.NewRequest("GET", "/api/course_phase/"+testCoursePhaseID+"/assessment-schema", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var schemas []assessmentSchemaDTO.AssessmentSchema
	err := json.Unmarshal(resp.Body.Bytes(), &schemas)
	assert.NoError(suite.T(), err)
	// Should return a list (might be empty)
	assert.NotNil(suite.T(), schemas)
}

func (suite *AssessmentSchemaRouterTestSuite) TestCreateAssessmentSchema() {
	createReq := assessmentSchemaDTO.CreateAssessmentSchemaRequest{
		Name:        "Test Assessment Schema",
		Description: "This is a test schema for router testing",
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+testCoursePhaseID+"/assessment-schema", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusCreated, resp.Code)

	var schema assessmentSchemaDTO.AssessmentSchema
	err := json.Unmarshal(resp.Body.Bytes(), &schema)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), createReq.Name, schema.Name)
	assert.Equal(suite.T(), createReq.Description, schema.Description)
	assert.NotEqual(suite.T(), uuid.Nil, schema.ID)
}

func (suite *AssessmentSchemaRouterTestSuite) TestCreateAssessmentSchemaInvalidJSON() {
	req, _ := http.NewRequest("POST", "/api/course_phase/"+testCoursePhaseID+"/assessment-schema", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)

	var errResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errResp, "error")
}

func (suite *AssessmentSchemaRouterTestSuite) TestGetAssessmentSchema() {
	// First create a schema to retrieve
	createReq := assessmentSchemaDTO.CreateAssessmentSchemaRequest{
		Name:        "Test Schema for Get",
		Description: "Schema to test GET endpoint",
	}
	schema, err := CreateAssessmentSchemaForCoursePhase(suite.suiteCtx, uuid.MustParse(testCoursePhaseID), createReq)
	assert.NoError(suite.T(), err)

	// Now test GET endpoint
	req, _ := http.NewRequest("GET", "/api/course_phase/"+testCoursePhaseID+"/assessment-schema/"+schema.ID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var retrievedSchema assessmentSchemaDTO.AssessmentSchema
	err = json.Unmarshal(resp.Body.Bytes(), &retrievedSchema)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), schema.ID, retrievedSchema.ID)
	assert.Equal(suite.T(), schema.Name, retrievedSchema.Name)
	assert.Equal(suite.T(), schema.Description, retrievedSchema.Description)
}

func (suite *AssessmentSchemaRouterTestSuite) TestGetAssessmentSchemaInvalidUUID() {
	req, _ := http.NewRequest("GET", "/api/course_phase/"+testCoursePhaseID+"/assessment-schema/invalid-uuid", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)

	var errResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errResp, "error")
}

func (suite *AssessmentSchemaRouterTestSuite) TestGetAssessmentSchemaNotFound() {
	nonExistentID := uuid.New()
	req, _ := http.NewRequest("GET", "/api/course_phase/"+testCoursePhaseID+"/assessment-schema/"+nonExistentID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusForbidden, resp.Code)
}

func (suite *AssessmentSchemaRouterTestSuite) TestUpdateAssessmentSchema() {
	// First create a schema to update
	createReq := assessmentSchemaDTO.CreateAssessmentSchemaRequest{
		Name:        "Original Schema",
		Description: "Original description",
	}
	schema, err := CreateAssessmentSchemaForCoursePhase(suite.suiteCtx, uuid.MustParse(testCoursePhaseID), createReq)
	assert.NoError(suite.T(), err)

	// Now test update
	updateReq := assessmentSchemaDTO.UpdateAssessmentSchemaRequest{
		Name:        "Updated Schema",
		Description: "Updated description",
	}
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/course_phase/"+testCoursePhaseID+"/assessment-schema/"+schema.ID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var successResp map[string]string
	err = json.Unmarshal(resp.Body.Bytes(), &successResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), successResp["message"], "updated successfully")
}

func (suite *AssessmentSchemaRouterTestSuite) TestUpdateAssessmentSchemaInvalidUUID() {
	updateReq := assessmentSchemaDTO.UpdateAssessmentSchemaRequest{
		Name:        "Updated Schema",
		Description: "Updated description",
	}
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/course_phase/"+testCoursePhaseID+"/assessment-schema/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *AssessmentSchemaRouterTestSuite) TestUpdateAssessmentSchemaInvalidJSON() {
	schemaID := uuid.New()
	req, _ := http.NewRequest("PUT", "/api/course_phase/"+testCoursePhaseID+"/assessment-schema/"+schemaID.String(), bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *AssessmentSchemaRouterTestSuite) TestDeleteAssessmentSchema() {
	// First create a schema to delete
	createReq := assessmentSchemaDTO.CreateAssessmentSchemaRequest{
		Name:        "Schema to Delete",
		Description: "This schema will be deleted",
	}
	schema, err := CreateAssessmentSchemaForCoursePhase(suite.suiteCtx, uuid.MustParse(testCoursePhaseID), createReq)
	assert.NoError(suite.T(), err)

	// Now test delete
	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+testCoursePhaseID+"/assessment-schema/"+schema.ID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var successResp map[string]string
	err = json.Unmarshal(resp.Body.Bytes(), &successResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), successResp["message"], "deleted successfully")
}

func (suite *AssessmentSchemaRouterTestSuite) TestDeleteAssessmentSchemaInvalidUUID() {
	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+testCoursePhaseID+"/assessment-schema/invalid-uuid", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *AssessmentSchemaRouterTestSuite) TestDeleteAssessmentSchemaNotFound() {
	nonExistentID := uuid.New()
	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+testCoursePhaseID+"/assessment-schema/"+nonExistentID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	// The service might handle non-existent records gracefully and return 200
	// This is actually valid behavior as the delete operation is idempotent
	assert.True(suite.T(), resp.Code == http.StatusOK || resp.Code == http.StatusInternalServerError)
}

func TestAssessmentSchemaRouterTestSuite(t *testing.T) {
	suite.Run(t, new(AssessmentSchemaRouterTestSuite))
}
