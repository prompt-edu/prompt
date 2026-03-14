package copy

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/course/copy/courseCopyDTO"
	"github.com/prompt-edu/prompt/servers/core/course/courseDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CourseCopyRouterTestSuite struct {
	suite.Suite
	router            *gin.Engine
	ctx               context.Context
	cleanup           func()
	courseCopyService CourseCopyService
}

func (suite *CourseCopyRouterTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Set up PostgreSQL container
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../../database_dumps/copy_course_test.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}

	mockCreateGroupsAndRoles := func(ctx context.Context, courseName, iterationName, userID string) error {
		// No-op or add assertions for test
		return nil
	}

	suite.cleanup = cleanup
	suite.courseCopyService = CourseCopyService{
		queries:                    *testDB.Queries,
		conn:                       testDB.Conn,
		createCourseGroupsAndRoles: mockCreateGroupsAndRoles,
	}

	CourseCopyServiceSingleton = &suite.courseCopyService

	// Init the permissionValidation service
	permissionValidation.InitValidationService(*testDB.Queries, testDB.Conn)

	// Initialize router
	suite.router = gin.Default()
	api := suite.router.Group("/api")
	setupCourseCopyRouter(api, func() gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddleware([]string{"PROMPT_Admin", "iPraktikum-ios24245-Lecturer"})
	}, sdkTestUtils.MockPermissionMiddleware, sdkTestUtils.MockPermissionMiddleware)
	coursePhase.InitCoursePhaseModule(api, *testDB.Queries, testDB.Conn)
}

func (suite *CourseCopyRouterTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *CourseCopyRouterTestSuite) TestCopyCourse() {
	courseID := "c1f8060d-7381-4b64-a6ea-5ba8e8ac88dd"

	copyCourseRequest := courseCopyDTO.CopyCourseRequest{
		Name:        "Copied Course",
		SemesterTag: pgtype.Text{String: "ws2425", Valid: true},
		StartDate:   pgtype.Date{Valid: true, Time: time.Now()},
		EndDate:     pgtype.Date{Valid: true, Time: time.Now().Add(24 * time.Hour)},
		Template:    false,
	}

	body, _ := json.Marshal(copyCourseRequest)
	req, _ := http.NewRequest("POST", "/api/courses/"+courseID+"/copy", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusCreated, resp.Code)

	var copiedCourse courseDTO.Course
	err := json.Unmarshal(resp.Body.Bytes(), &copiedCourse)

	assert.NoError(suite.T(), err, "Unmarshalling the copied course response should not produce an error")
	assert.Equal(suite.T(), copyCourseRequest.Name, copiedCourse.Name, "Copied course name should match")
	assert.Equal(suite.T(), copyCourseRequest.SemesterTag.String, copiedCourse.SemesterTag.String, "Copied course semester tag should match")
	assert.Equal(suite.T(), copyCourseRequest.StartDate.Time.Format("2006-01-02"), copiedCourse.StartDate.Time.Format("2006-01-02"), "Copied course start date should match")
	assert.Equal(suite.T(), copyCourseRequest.EndDate.Time.Format("2006-01-02"), copiedCourse.EndDate.Time.Format("2006-01-02"), "Copied course end date should match")
	assert.NotEqual(suite.T(), uuid.MustParse(courseID), copiedCourse.ID, "Copied course ID should be different from original course ID")
}

func TestCourseCopyRouterTestSuite(t *testing.T) {
	suite.Run(t, new(CourseCopyRouterTestSuite))
}
