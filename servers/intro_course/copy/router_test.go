package copy

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/intro_course/testutils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CopyRouterTestSuite struct {
	suite.Suite
	ctx     context.Context
	router  *gin.Engine
	cleanup func()
}

func (suite *CopyRouterTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()

	testDB, cleanup, err := testutils.SetupTestDB(suite.ctx, "../database_dumps/intro_course.sql")
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup

	CopyServiceSingleton = &CopyService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}

	suite.router = gin.Default()
	api := suite.router.Group("/intro-course/api")
	authMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return testutils.DefaultMockAuthMiddleware()
	}
	setupCopyRouter(api, authMiddleware)
}

func (suite *CopyRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *CopyRouterTestSuite) TestCopyEndpointSuccess() {
	body := promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: uuid.New(),
		TargetCoursePhaseID: uuid.New(),
	}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/intro-course/api/copy", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *CopyRouterTestSuite) TestCopyEndpointInvalidPayload() {
	req, _ := http.NewRequest("POST", "/intro-course/api/copy", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func TestCopyRouterTestSuite(t *testing.T) {
	suite.Run(t, new(CopyRouterTestSuite))
}
