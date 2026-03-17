package infrastructureSetup

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type InfrastructureRouterTestSuite struct {
	suite.Suite
	ctx           context.Context
	router        *gin.Engine
	cleanup       func()
	coursePhaseID uuid.UUID
}

func (suite *InfrastructureRouterTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/intro_course.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.coursePhaseID = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")

	service := InfrastructureService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	InfrastructureServiceSingleton = &service

	suite.router = gin.Default()
	api := suite.router.Group("/intro-course/api/course_phase/:coursePhaseID")
	authMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddleware(allowedRoles)
	}
	setupInfrastructureRouter(api, authMiddleware)
}

func (suite *InfrastructureRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func TestInfrastructureRouterTestSuite(t *testing.T) {
	suite.Run(t, new(InfrastructureRouterTestSuite))
}

func (suite *InfrastructureRouterTestSuite) TestGetAllStudentGitlabStatus() {
	req, _ := http.NewRequest("GET", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/infrastructure/gitlab/student-setup", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	var statuses []map[string]interface{}
	assert.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &statuses))
	assert.True(suite.T(), len(statuses) >= 1)
}

func (suite *InfrastructureRouterTestSuite) TestManuallyOverwriteStudentGitlabStatus() {
	courseParticipationID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	req, _ := http.NewRequest("PUT", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/infrastructure/gitlab/student-setup/"+courseParticipationID.String()+"/manual", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}
