package tutor

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/intro_course/testutils"
	"github.com/prompt-edu/prompt/servers/intro_course/tutor/tutorDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TutorRouterTestSuite struct {
	suite.Suite
	ctx           context.Context
	router        *gin.Engine
	cleanup       func()
	coursePhaseID uuid.UUID
	mockServer    *httptest.Server
	prevCoreHost  string
}

func (suite *TutorRouterTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(suite.ctx, "../database_dumps/intro_course.sql")
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.coursePhaseID = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	suite.prevCoreHost = os.Getenv("SERVER_CORE_HOST")
	_ = os.Setenv("SERVER_CORE_HOST", suite.mockServer.URL)

	service := TutorService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	TutorServiceSingleton = &service

	suite.router = gin.Default()
	api := suite.router.Group("/intro-course/api/course_phase/:coursePhaseID")
	authMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return testutils.DefaultMockAuthMiddleware()
	}
	setupTutorRouter(api, authMiddleware)
}

func (suite *TutorRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
	if suite.mockServer != nil {
		suite.mockServer.Close()
	}
	if suite.prevCoreHost != "" {
		_ = os.Setenv("SERVER_CORE_HOST", suite.prevCoreHost)
	} else {
		_ = os.Unsetenv("SERVER_CORE_HOST")
	}
}

func TestTutorRouterTestSuite(t *testing.T) {
	suite.Run(t, new(TutorRouterTestSuite))
}

func (suite *TutorRouterTestSuite) TestGetTutors() {
	req, _ := http.NewRequest("GET", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/tutor", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *TutorRouterTestSuite) TestImportTutors() {
	newTutors := []tutorDTO.Tutor{
		{
			ID:                  uuid.New(),
			FirstName:           "Router",
			LastName:            "Tutor",
			Email:               "router.tutor@example.com",
			MatriculationNumber: "300001",
			UniversityLogin:     "routertutor",
		},
	}
	body, _ := json.Marshal(newTutors)
	req, _ := http.NewRequest("POST", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/tutor/course/"+uuid.New().String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
}

func (suite *TutorRouterTestSuite) TestImportTutorsInvalidCoursePhase() {
	req, _ := http.NewRequest("POST", "/intro-course/api/course_phase/not-a-uuid/tutor/course/"+uuid.New().String(), bytes.NewBuffer([]byte("[]")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *TutorRouterTestSuite) TestUpdateGitLabUsername() {
	update := tutorDTO.UpdateTutor{GitlabUsername: "router-gitlab"}
	body, _ := json.Marshal(update)
	req, _ := http.NewRequest("PUT", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/tutor/11111111-1111-1111-1111-111111111111", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *TutorRouterTestSuite) TestUpdateGitLabUsernameInvalidTutorID() {
	update := tutorDTO.UpdateTutor{GitlabUsername: "router-gitlab"}
	body, _ := json.Marshal(update)
	req, _ := http.NewRequest("PUT", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/tutor/not-a-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}
