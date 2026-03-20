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
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CopyRouterTestSuite struct {
	suite.Suite
	ctx    context.Context
	router *gin.Engine
}

func (suite *CopyRouterTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()

	CopyServiceSingleton = &CopyService{}

	suite.router = gin.Default()
	api := suite.router.Group("/self-team-allocation/api")
	authMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.DefaultMockAuthMiddleware()
	}
	setupCopyRouter(api, authMiddleware)
}

func (suite *CopyRouterTestSuite) TestCopyEndpointSuccess() {
	body := promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: uuid.New(),
		TargetCoursePhaseID: uuid.New(),
	}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/self-team-allocation/api/copy", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *CopyRouterTestSuite) TestCopyEndpointInvalidPayload() {
	req, _ := http.NewRequest("POST", "/self-team-allocation/api/copy", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func TestCopyRouterTestSuite(t *testing.T) {
	suite.Run(t, new(CopyRouterTestSuite))
}
