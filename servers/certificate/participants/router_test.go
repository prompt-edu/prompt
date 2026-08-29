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
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
	"github.com/prompt-edu/prompt/servers/certificate/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ParticipantsRouterTestSuite struct {
	suite.Suite
	router          *gin.Engine
	suiteCtx        context.Context
	cleanup         func()
	mockCoreCleanup func()
	service         *ParticipantsService
}

func (s *ParticipantsRouterTestSuite) SetupSuite() {
	s.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(s.suiteCtx, "../database_dumps/certificate.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
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

	s.service = NewParticipantsService(*testDB.Queries, coreURL)

	gin.SetMode(gin.TestMode)
	s.router = gin.Default()
	api := s.router.Group("/api/course_phase/:coursePhaseID")
	RegisterRoutes(api, s.service, sdkTestUtils.MockPermissionMiddleware)
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

	targetStudentID := uuid.MustParse("30000000-0000-0000-0000-000000000001")
	foundTargetStudent := false

	// Check that download status is enriched from DB
	for _, p := range participants {
		if p.Student.ID == targetStudentID {
			foundTargetStudent = true
			assert.True(s.T(), p.HasDownloaded)
			assert.Equal(s.T(), int32(3), p.DownloadCount)
			assert.NotEmpty(s.T(), p.PassStatus)
			assert.Equal(s.T(), "03012345", p.Student.MatriculationNumber)
			assert.Equal(s.T(), "ge12abc", p.Student.UniversityLogin)
			assert.True(s.T(), p.Student.HasUniversityAccount)
			assert.Equal(s.T(), "male", string(p.Student.Gender))
		}
	}
	assert.True(s.T(), foundTargetStudent)
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

	teamName, err := s.service.GetStudentTeamName(s.suiteCtx, "Bearer mock-token", coursePhaseID, studentID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "BMW", teamName)
}

func (s *ParticipantsRouterTestSuite) TestGetStudentTeamName_SecondStudent() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	studentID := uuid.MustParse("30000000-0000-0000-0000-000000000002")

	teamName, err := s.service.GetStudentTeamName(s.suiteCtx, "Bearer mock-token", coursePhaseID, studentID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Siemens", teamName)
}

func (s *ParticipantsRouterTestSuite) TestGetStudentTeamName_UnknownStudent() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	unknownStudentID := uuid.New()

	teamName, err := s.service.GetStudentTeamName(s.suiteCtx, "Bearer mock-token", coursePhaseID, unknownStudentID)
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), teamName)
}

func TestParticipantsRouterTestSuite(t *testing.T) {
	suite.Run(t, new(ParticipantsRouterTestSuite))
}
