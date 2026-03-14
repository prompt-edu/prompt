package courseParticipation

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
	"github.com/prompt-edu/prompt/servers/core/course/courseParticipation/courseParticipationDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RouterTestSuite struct {
	suite.Suite
	router                     *gin.Engine
	ctx                        context.Context
	cleanup                    func()
	courseParticipationService CourseParticipationService
}

func (suite *RouterTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Set up PostgreSQL container
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../../database_dumps/course_participation_test.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}

	suite.cleanup = cleanup
	suite.courseParticipationService = CourseParticipationService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	CourseParticipationServiceSingleton = &suite.courseParticipationService

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
	setupCourseParticipationRouter(api, authMiddleware, permissionIDMiddleware)
	return router
}

func (suite *RouterTestSuite) TestGetCourseParticipationsForCourse() {
	courseID := "3f42d322-e5bf-4faa-b576-51f2cab14c2e"

	req := httptest.NewRequest(http.MethodGet, "/api/courses/"+courseID+"/participations", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var participations []courseParticipationDTO.GetCourseParticipation
	err := json.Unmarshal(w.Body.Bytes(), &participations)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(participations), 0, "Expected participations for the course")

	expectedStudentIDs := []uuid.UUID{
		uuid.MustParse("3d1f3b00-87f3-433b-a713-178c4050411a"),
		uuid.MustParse("7dc1c4e8-4255-4874-80a0-0c12b958744b"),
		uuid.MustParse("500db7ed-2eb2-42d0-82b3-8750e12afa8b"),
	}

	for i, participation := range participations {
		assert.Equal(suite.T(), uuid.MustParse(courseID), participation.CourseID, "Expected CourseID to match")
		assert.Equal(suite.T(), expectedStudentIDs[i], participation.StudentID, "Expected StudentID to match")
	}
}

func (suite *RouterTestSuite) TestCreateCourseParticipation() {
	courseID := "918977e1-2d27-4b55-9064-8504ff027a1a"
	newParticipation := courseParticipationDTO.CreateCourseParticipation{
		StudentID: uuid.MustParse("3d1f3b00-87f3-433b-a713-178c4050411a"),
	}

	body, _ := json.Marshal(newParticipation)
	req := httptest.NewRequest(http.MethodPost, "/api/courses/"+courseID+"/participations/enroll", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var createdParticipation courseParticipationDTO.GetCourseParticipation
	err := json.Unmarshal(w.Body.Bytes(), &createdParticipation)
	assert.NoError(suite.T(), err)

	// Validate the created participation
	assert.Equal(suite.T(), uuid.MustParse(courseID), createdParticipation.CourseID, "Expected CourseID to match")
	assert.Equal(suite.T(), newParticipation.StudentID, createdParticipation.StudentID, "Expected StudentID to match")
	assert.NotEqual(suite.T(), uuid.Nil, createdParticipation.ID, "Expected a valid UUID for the new participation")
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}
