package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/auth/authDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type CourseRouterTestSuite struct {
	suite.Suite
	router                 *gin.Engine
	ctx                    context.Context
	cleanup                func()
	authService CoursePhaseAuthService
}

func (suite *CourseRouterTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Set up PostgreSQL container with test dump.
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/full_db.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup

	suite.authService = CoursePhaseAuthService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	CoursePhaseAuthServiceSingleton = &suite.authService

	// Init the permissionValidation service.
	permissionValidation.InitValidationService(*testDB.Queries, testDB.Conn)

	// Initialize router.
	suite.router = gin.Default()
	api := suite.router.Group("/api")
	setupCoursePhaseAuthRouter(api,
		// Auth middleware: simulate a student with known email, matriculation number, and university login.
		func() gin.HandlerFunc {
			return sdkTestUtils.MockAuthMiddlewareWithEmail([]string{"ios2425-TestCourse-Student"}, "existingstudent@example.com", "09999999", "as45fgh")
		},
		// Permission middleware: use test helper.
		sdkTestUtils.MockPermissionMiddleware,
	)
}

func (suite *CourseRouterTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *CourseRouterTestSuite) TestGetCoursePhaseAuthRoles() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	url := "/api/auth/course_phase/" + coursePhaseID + "/roles"
	req, err := http.NewRequest("GET", url, nil)
	suite.Require().NoError(err)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	// Check that the expected keys exist.
	suite.Contains(response, "courseLecturerRole")
	suite.Contains(response, "courseEditorRole")
	suite.Contains(response, "customRolePrefix")
}

func (suite *CourseRouterTestSuite) TestGetCoursePhaseAuthRoles_InvalidUUID() {
	// Passing an invalid UUID should return a 400 Bad Request.
	invalidID := "not-a-uuid"
	url := "/api/auth/course_phase/" + invalidID + "/roles"
	req, err := http.NewRequest("GET", url, nil)
	suite.Require().NoError(err)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)
	suite.Equal("invalid course phase ID", response["error"])
}

func (suite *CourseRouterTestSuite) TestGetCoursePhaseParticipation_Success() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	url := "/api/auth/course_phase/" + coursePhaseID + "/is_student"
	req, err := http.NewRequest("GET", url, nil)
	suite.Require().NoError(err)
	// Rely on the auth middleware to set matriculationNumber and universityLogin.

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)
	var response authDTO.GetCoursePhaseParticipation
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	logrus.Info(response)

	// Assert that the student is indeed part of the course phase.
	suite.True(response.IsStudentOfCoursePhase)

	expectedParticipationID := "1378db54-4ab4-4225-ac67-68e4345f21e2"
	suite.Equal(expectedParticipationID, response.CourseParticipationID.String())
}

func (suite *CourseRouterTestSuite) TestGetCoursePhaseParticipation_NOT_PART() {
	coursePhaseID := "4e736d05-c125-48f0-8fa0-848b03ca6908"
	url := "/api/auth/course_phase/" + coursePhaseID + "/is_student"
	req, err := http.NewRequest("GET", url, nil)
	suite.Require().NoError(err)
	// Rely on the auth middleware to set matriculationNumber and universityLogin.

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)
	var response authDTO.GetCoursePhaseParticipation
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)

	// Assert that the student is indeed part of the course phase.
	suite.False(response.IsStudentOfCoursePhase)

	expectedParticipationID := "1378db54-4ab4-4225-ac67-68e4345f21e2"
	suite.Equal(expectedParticipationID, response.CourseParticipationID.String())
}

func TestCourseRouterTestSuite(t *testing.T) {
	suite.Run(t, new(CourseRouterTestSuite))
}
