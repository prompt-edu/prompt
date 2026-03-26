package allocation

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
	"github.com/prompt-edu/prompt/servers/self_team_allocation/allocation/allocationDTO"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AllocationRouterTestSuite struct {
	suite.Suite
	ctx     context.Context
	router  *gin.Engine
	cleanup func()
}

func (suite *AllocationRouterTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	require.NoError(suite.T(), err)
	suite.cleanup = cleanup

	AllocationServiceSingleton = &AllocationService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}

	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	authMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.DefaultMockAuthMiddleware()
	}
	setupAllocationRouter(api, authMiddleware)
}

func (suite *AllocationRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *AllocationRouterTestSuite) TestGetAllAllocations() {
	req, _ := http.NewRequest("GET", "/api/course_phase/11111111-1111-1111-1111-111111111111/allocation", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)

	var allocations []allocationDTO.AllocationWithParticipation
	err := json.Unmarshal(resp.Body.Bytes(), &allocations)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), allocations, 3)
}

func (suite *AllocationRouterTestSuite) TestGetAllAllocationsInvalidCoursePhase() {
	req, _ := http.NewRequest("GET", "/api/course_phase/not-a-uuid/allocation", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *AllocationRouterTestSuite) TestGetAllocationByCourseParticipationID() {
	url := "/api/course_phase/11111111-1111-1111-1111-111111111111/allocation/aaaa1111-1111-1111-1111-111111111111"
	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)

	var teamID string
	err := json.Unmarshal(resp.Body.Bytes(), &teamID)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa").String(), teamID)
}

func (suite *AllocationRouterTestSuite) TestGetAllocationByCourseParticipationIDInvalidUUID() {
	url := "/api/course_phase/11111111-1111-1111-1111-111111111111/allocation/not-a-uuid"
	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *AllocationRouterTestSuite) TestGetAllocationByCourseParticipationIDNotFound() {
	url := "/api/course_phase/11111111-1111-1111-1111-111111111111/allocation/ffffffff-ffff-ffff-ffff-ffffffffffff"
	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusInternalServerError, resp.Code)
}

func TestAllocationRouterTestSuite(t *testing.T) {
	suite.Run(t, new(AllocationRouterTestSuite))
}
