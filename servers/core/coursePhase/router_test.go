package coursePhase

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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RouterTestSuite struct {
	suite.Suite
	router             *gin.Engine
	ctx                context.Context
	cleanup            func()
	coursePhaseService CoursePhaseService
}

func (suite *RouterTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Set up PostgreSQL container
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/course_phase_test.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}

	suite.cleanup = cleanup
	suite.coursePhaseService = CoursePhaseService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	CoursePhaseServiceSingleton = &suite.coursePhaseService
	resolution.InitResolutionModule("localhost:8080")

	suite.router = setupRouter()
}

func (suite *RouterTestSuite) TearDownSuite() {
	suite.cleanup()
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	authMiddleware := func() gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddleware([]string{"PROMPT_Admin", "iPraktikum-ios24245-Lecturer"})
	}
	permissionIDMiddleware := sdkTestUtils.MockPermissionMiddleware
	setupCoursePhaseRouter(api, authMiddleware, permissionIDMiddleware, permissionIDMiddleware)
	return router
}

func (suite *RouterTestSuite) TestGetCoursePhaseByID() {
	req := httptest.NewRequest(http.MethodGet, "/api/course_phases/3d1f3b00-87f3-433b-a713-178c4050411b", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var coursePhase coursePhaseDTO.CoursePhase
	err := json.Unmarshal(w.Body.Bytes(), &coursePhase)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Test", coursePhase.Name, "Expected course phase name to match")
	assert.False(suite.T(), coursePhase.IsInitialPhase, "Expected course phase to not be an initial phase")
	assert.Equal(suite.T(), uuid.MustParse("3f42d322-e5bf-4faa-b576-51f2cab14c2e"), coursePhase.CourseID, "Expected CourseID to match")
	assert.Equal(suite.T(), uuid.MustParse("7dc1c4e8-4255-4874-80a0-0c12b958744b"), coursePhase.CoursePhaseTypeID, "Expected CoursePhaseTypeID to match")
	assert.Equal(suite.T(), "test-value", coursePhase.RestrictedData["test-key"], "Expected MetaData to match")
}

func (suite *RouterTestSuite) TestCreateCoursePhase() {
	jsonData := `{"new_key": "new_value"}`
	var data meta.MetaData
	err := json.Unmarshal([]byte(jsonData), &data)
	assert.NoError(suite.T(), err)

	jsonData = `{"new_key2": "new_value"}`
	var studentData meta.MetaData
	err = json.Unmarshal([]byte(jsonData), &studentData)
	assert.NoError(suite.T(), err)

	newCoursePhase := coursePhaseDTO.CreateCoursePhase{
		CourseID:            uuid.MustParse("3f42d322-e5bf-4faa-b576-51f2cab14c2e"),
		Name:                "New Phase",
		IsInitialPhase:      false,
		RestrictedData:      data,
		StudentReadableData: studentData,
		CoursePhaseTypeID:   uuid.MustParse("7dc1c4e8-4255-4874-80a0-0c12b958744c"),
	}

	body, _ := json.Marshal(newCoursePhase)
	req := httptest.NewRequest(http.MethodPost, "/api/course_phases/course/3f42d322-e5bf-4faa-b576-51f2cab14c2e", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createdCoursePhase coursePhaseDTO.CoursePhase
	err = json.Unmarshal(w.Body.Bytes(), &createdCoursePhase)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "New Phase", createdCoursePhase.Name, "Expected course phase name to match")
	assert.False(suite.T(), createdCoursePhase.IsInitialPhase, "Expected course phase to not be an initial phase")
	assert.Equal(suite.T(), newCoursePhase.CourseID, createdCoursePhase.CourseID, "Expected CourseID to match")
	assert.Equal(suite.T(), newCoursePhase.RestrictedData, createdCoursePhase.RestrictedData, "Expected MetaData to match")
	assert.Equal(suite.T(), newCoursePhase.StudentReadableData, createdCoursePhase.StudentReadableData, "Expected MetaData to match")
	assert.Equal(suite.T(), newCoursePhase.CoursePhaseTypeID, createdCoursePhase.CoursePhaseTypeID, "Expected CoursePhaseTypeID to match")
}

func (suite *RouterTestSuite) TestUpdateCoursePhase() {
	jsonData := `{"updated_key": "updated_value"}`
	var data meta.MetaData
	err := json.Unmarshal([]byte(jsonData), &data)
	assert.NoError(suite.T(), err)

	jsonData = `{"new_key2": "new_value"}`
	var studentData meta.MetaData
	err = json.Unmarshal([]byte(jsonData), &studentData)
	assert.NoError(suite.T(), err)

	updatedCoursePhase := coursePhaseDTO.UpdateCoursePhase{
		ID:                  uuid.MustParse("3d1f3b00-87f3-433b-a713-178c4050411b"),
		Name:                pgtype.Text{Valid: true, String: "Updated Phase"},
		RestrictedData:      data,
		StudentReadableData: studentData,
	}

	body, _ := json.Marshal(updatedCoursePhase)
	req := httptest.NewRequest(http.MethodPut, "/api/course_phases/3d1f3b00-87f3-433b-a713-178c4050411b", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify the update by fetching the updated course phase
	fetchReq := httptest.NewRequest(http.MethodGet, "/api/course_phases/3d1f3b00-87f3-433b-a713-178c4050411b", nil)
	fetchRes := httptest.NewRecorder()
	suite.router.ServeHTTP(fetchRes, fetchReq)

	assert.Equal(suite.T(), http.StatusOK, fetchRes.Code)

	var fetchedCoursePhase coursePhaseDTO.CoursePhase
	err = json.Unmarshal(fetchRes.Body.Bytes(), &fetchedCoursePhase)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Phase", fetchedCoursePhase.Name, "Expected updated course phase name to match")
	assert.Equal(suite.T(), updatedCoursePhase.RestrictedData["updated_key"], fetchedCoursePhase.RestrictedData["updated_key"], "Expected updated metadata to match")
	assert.Equal(suite.T(), "test-value", fetchedCoursePhase.RestrictedData["test-key"], "Expected existing metadata to match")
	assert.Equal(suite.T(), updatedCoursePhase.StudentReadableData["new_key2"], fetchedCoursePhase.StudentReadableData["new_key2"], "Expected updated metadata to match")

}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}
