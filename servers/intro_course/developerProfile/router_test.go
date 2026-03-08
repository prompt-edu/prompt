package developerProfile

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prompt-edu/prompt/servers/intro_course/developerProfile/developerProfileDTO"
	"github.com/prompt-edu/prompt/servers/intro_course/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DeveloperProfileRouterTestSuite struct {
	suite.Suite
	ctx           context.Context
	router        *gin.Engine
	cleanup       func()
	coursePhaseID uuid.UUID
	studentID     uuid.UUID
}

func (suite *DeveloperProfileRouterTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(suite.ctx, "../database_dumps/intro_course.sql")
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.coursePhaseID = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	suite.studentID = uuid.MustParse("33333333-3333-3333-3333-333333333333")

	service := DeveloperProfileService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	DeveloperProfileServiceSingleton = &service

	suite.router = gin.Default()
	api := suite.router.Group("/intro-course/api/course_phase/:coursePhaseID")
	authMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return testutils.MockAuthMiddlewareWithParticipation(allowedRoles, suite.studentID)
	}
	setupDeveloperProfileRouter(api, authMiddleware)
}

func (suite *DeveloperProfileRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func TestDeveloperProfileRouterTestSuite(t *testing.T) {
	suite.Run(t, new(DeveloperProfileRouterTestSuite))
}

func (suite *DeveloperProfileRouterTestSuite) TestCreateDeveloperProfile() {
	request := developerProfileDTO.PostDeveloperProfile{
		AppleID:        "router@apple.com",
		GitLabUsername: "routeruser",
		HasMacBook:     true,
		IPhoneUDID:     pgtype.Text{String: "AAAABBBB-CCCCDDDDEEEEFFFF", Valid: true},
	}
	body, _ := json.Marshal(request)
	// Use a different course phase where the student doesn't have a profile yet
	newCoursePhaseID := "5179d58a-d00d-4fa7-94a5-397bc69fab03"
	req, _ := http.NewRequest("POST", "/intro-course/api/course_phase/"+newCoursePhaseID+"/developer_profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
}

func (suite *DeveloperProfileRouterTestSuite) TestCreateDeveloperProfileInvalidUDID() {
	request := developerProfileDTO.PostDeveloperProfile{
		AppleID:        "invalid@apple.com",
		GitLabUsername: "invalid",
		HasMacBook:     true,
		IPhoneUDID:     pgtype.Text{String: "INVALID", Valid: true},
	}
	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/developer_profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *DeveloperProfileRouterTestSuite) TestGetOwnDeveloperProfile() {
	req, _ := http.NewRequest("GET", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/developer_profile/self", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	var profile developerProfileDTO.DeveloperProfile
	assert.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &profile))
	assert.Equal(suite.T(), "student1git", profile.GitLabUsername)
}

func (suite *DeveloperProfileRouterTestSuite) TestGetAllDeveloperProfiles() {
	req, _ := http.NewRequest("GET", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/developer_profile", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	var profiles []developerProfileDTO.DeveloperProfile
	assert.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &profiles))
	assert.True(suite.T(), len(profiles) >= 2)
}

func (suite *DeveloperProfileRouterTestSuite) TestUpdateDeveloperProfile() {
	update := developerProfileDTO.DeveloperProfile{
		AppleID:        "updated-router@apple.com",
		GitLabUsername: "student1git",
		HasMacBook:     false,
	}
	body, _ := json.Marshal(update)
	req, _ := http.NewRequest("PUT", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/developer_profile/"+suite.studentID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *DeveloperProfileRouterTestSuite) TestUpdateDeveloperProfileInvalidID() {
	req, _ := http.NewRequest("PUT", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/developer_profile/not-a-uuid", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *DeveloperProfileRouterTestSuite) TestGetDevicesForAllParticipations() {
	req, _ := http.NewRequest("GET", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/devices", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	var devices []developerProfileDTO.DeviceWithParticipationID
	assert.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &devices))
	assert.True(suite.T(), len(devices) >= 2)
}

func (suite *DeveloperProfileRouterTestSuite) TestGetDevicesForCourseParticipation() {
	req, _ := http.NewRequest("GET", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/devices/"+suite.studentID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	var devices []string
	assert.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &devices))
	assert.NotEmpty(suite.T(), devices)
}

func (suite *DeveloperProfileRouterTestSuite) TestGetDevicesForCourseParticipationNotFound() {
	req, _ := http.NewRequest("GET", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/devices/"+uuid.New().String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusNotFound, resp.Code)
	var errResp map[string]string
	assert.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &errResp))
	assert.Contains(suite.T(), errResp["error"], "No devices found")
}
