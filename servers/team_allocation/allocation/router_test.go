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
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AllocationRouterTestSuite struct {
	suite.Suite
	router            *gin.Engine
	suiteCtx          context.Context
	cleanup           func()
	allocationService AllocationService
}

func (suite *AllocationRouterTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/allocations.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.allocationService = AllocationService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	AllocationServiceSingleton = &suite.allocationService
	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "admin@example.com", "03711111", "ab12cde")
	}
	setupAllocationRouter(api, testMiddleware)
}

func (suite *AllocationRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *AllocationRouterTestSuite) TestGetAllAllocations() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/allocation", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var allocations []map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &allocations)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(allocations), 0, "Should return a list of allocations")
}

func (suite *AllocationRouterTestSuite) TestGetAllAllocationsInvalidCoursePhaseID() {
	req, _ := http.NewRequest("GET", "/api/course_phase/invalid-uuid/allocation", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *AllocationRouterTestSuite) TestGetAllocationByCourseParticipationID() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	courseParticipationID := "99999999-9999-9999-9999-999999999991"
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/allocation/"+courseParticipationID, nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var teamID string
	err := json.Unmarshal(resp.Body.Bytes(), &teamID)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), teamID, "Should return a team ID")
}

func (suite *AllocationRouterTestSuite) TestGetAllocationByCourseParticipationIDInvalidID() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/allocation/invalid-uuid", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *AllocationRouterTestSuite) TestGetAllocationByCourseParticipationIDNotFound() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	nonExistentID := uuid.New().String()
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/allocation/"+nonExistentID, nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, resp.Code)
}

func TestAllocationRouterTestSuite(t *testing.T) {
	suite.Run(t, new(AllocationRouterTestSuite))
}
