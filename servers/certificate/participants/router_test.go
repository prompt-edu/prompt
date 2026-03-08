package participants

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ls1intum/prompt2/servers/certificate/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ParticipantsRouterTestSuite struct {
	suite.Suite
	router          *gin.Engine
	suiteCtx        context.Context
	cleanup         func()
	mockCoreCleanup func()
}

func (s *ParticipantsRouterTestSuite) SetupSuite() {
	s.suiteCtx = context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(s.suiteCtx, "../database_dumps/certificate.sql")
	if err != nil {
		s.T().Fatalf("Failed to set up test database: %v", err)
	}
	s.cleanup = cleanup

	_, mockCoreCleanup := testutils.SetupMockCoreService()
	s.mockCoreCleanup = mockCoreCleanup

	// The mock core service sets SERVER_CORE_HOST env var
	coreURL := "http://localhost:8080"
	if val, ok := os.LookupEnv("SERVER_CORE_HOST"); ok {
		coreURL = val
	}

	ParticipantsServiceSingleton = &ParticipantsService{
		queries:    *testDB.Queries,
		httpClient: http.DefaultClient,
		coreURL:    coreURL,
	}

	gin.SetMode(gin.TestMode)
	s.router = gin.Default()
	api := s.router.Group("/api/course_phase/:coursePhaseID")
	setupParticipantsRouter(api, testutils.MockPermissionMiddleware)
}

func (s *ParticipantsRouterTestSuite) TearDownSuite() {
	if s.mockCoreCleanup != nil {
		s.mockCoreCleanup()
	}
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *ParticipantsRouterTestSuite) TestGetParticipants() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	url := fmt.Sprintf("/api/course_phase/%s/participants", coursePhaseID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer mock-token")
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusOK, resp.Code)

	var participants []ParticipantWithDownloadStatus
	err := json.Unmarshal(resp.Body.Bytes(), &participants)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), participants, 2)

	// Check that download status is enriched from DB
	for _, p := range participants {
		if p.Student.ID == uuid.MustParse("30000000-0000-0000-0000-000000000001") {
			assert.True(s.T(), p.HasDownloaded)
			assert.Equal(s.T(), int32(3), p.DownloadCount)
			assert.NotEmpty(s.T(), p.PassStatus)
		}
	}
}

func (s *ParticipantsRouterTestSuite) TestGetParticipants_InvalidID() {
	url := "/api/course_phase/not-a-uuid/participants"

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusBadRequest, resp.Code)
}

func (s *ParticipantsRouterTestSuite) TestGetStudentTeamName() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	studentID := uuid.MustParse("30000000-0000-0000-0000-000000000001")

	teamName, err := GetStudentTeamName(s.suiteCtx, "Bearer mock-token", coursePhaseID, studentID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "BMW", teamName)
}

func (s *ParticipantsRouterTestSuite) TestGetStudentTeamName_SecondStudent() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	studentID := uuid.MustParse("30000000-0000-0000-0000-000000000002")

	teamName, err := GetStudentTeamName(s.suiteCtx, "Bearer mock-token", coursePhaseID, studentID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Siemens", teamName)
}

func (s *ParticipantsRouterTestSuite) TestGetStudentTeamName_UnknownStudent() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	unknownStudentID := uuid.New()

	teamName, err := GetStudentTeamName(s.suiteCtx, "Bearer mock-token", coursePhaseID, unknownStudentID)
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), teamName)
}

func TestParticipantsRouterTestSuite(t *testing.T) {
	suite.Run(t, new(ParticipantsRouterTestSuite))
}
