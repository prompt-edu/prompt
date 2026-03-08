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
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/interview/testutils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CopyRouterTestSuite struct {
	suite.Suite
	ctx     context.Context
	testDB  *testutils.TestDB
	cleanup func()
}

func (suite *CopyRouterTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()

	testDB, cleanup, err := testutils.SetupTestDB(suite.ctx, "../database_dumps/base.sql")
	require.NoError(suite.T(), err)
	suite.testDB = testDB
	suite.cleanup = cleanup

	CopyServiceSingleton = &CopyService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
}

func (suite *CopyRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *CopyRouterTestSuite) newRouter(identity testutils.MockIdentity) *gin.Engine {
	router := gin.Default()
	api := router.Group("/api/course_phase/:coursePhaseID")
	setupCopyRouter(api, testutils.NewMockAuthMiddleware(identity))
	return router
}

func (suite *CopyRouterTestSuite) TestCopyPhaseRoute() {
	router := suite.newRouter(testutils.MockIdentity{})

	reqBody := promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		TargetCoursePhaseID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
	}
	payload, _ := json.Marshal(reqBody)

	httpReq, _ := http.NewRequest("POST", "/api/course_phase/11111111-1111-1111-1111-111111111111/copy", bytes.NewBuffer(payload))
	httpReq.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, httpReq)

	require.Equal(suite.T(), http.StatusOK, resp.Code)
}

func TestCopyRouterTestSuite(t *testing.T) {
	suite.Run(t, new(CopyRouterTestSuite))
}
