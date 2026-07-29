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
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	scopedTutorLogin  = "ab12cde"
	teamAlphaParticip = "99999999-9999-9999-9999-999999999991"
	teamBetaParticip  = "99999999-9999-9999-9999-999999999992"
)

// tutorAuthMiddleware mocks an authenticated CourseEditor (non-lecturer) so the tutor-scoping
// middleware resolves a team instead of taking the lecturer/admin pass-through branch.
func tutorAuthMiddleware(login string) func(allowedRoles ...string) gin.HandlerFunc {
	return func(allowedRoles ...string) gin.HandlerFunc {
		return func(c *gin.Context) {
			keycloakTokenVerifier.SetTokenUser(c, keycloakTokenVerifier.TokenUser{
				IsEditor:        true,
				IsLecturer:      false,
				UniversityLogin: login,
			})
			c.Next()
		}
	}
}

func (suite *AllocationRouterTestSuite) routerAs(login string) *gin.Engine {
	router := gin.New()
	api := router.Group("/api/course_phase/:coursePhaseID")
	setupAllocationRouter(api, tutorAuthMiddleware(login), suite.allocationService.queries)
	return router
}

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
	setupAllocationRouter(api, testMiddleware, *testDB.Queries)
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

	assert.Equal(suite.T(), http.StatusNotFound, resp.Code)
}

func (suite *AllocationRouterTestSuite) TestTutorSeesOnlyOwnTeamAllocations() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	router := suite.routerAs(scopedTutorLogin)
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/allocation", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var allocations []map[string]interface{}
	assert.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &allocations))
	assert.Len(suite.T(), allocations, 1, "Tutor should only see allocations for their team")
	assert.Equal(suite.T(), teamAlphaParticip, allocations[0]["courseParticipationID"])
}

func (suite *AllocationRouterTestSuite) TestTutorAllowedOnOwnAllocation() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	router := suite.routerAs(scopedTutorLogin)
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/allocation/"+teamAlphaParticip, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *AllocationRouterTestSuite) TestTutorForbiddenOnOtherAllocation() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	router := suite.routerAs(scopedTutorLogin)
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/allocation/"+teamBetaParticip, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusForbidden, resp.Code)
}

func (suite *AllocationRouterTestSuite) TestNonTutorEditorSeesAllAllocations() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	router := suite.routerAs("zz99zzz")
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/allocation", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var allocations []map[string]interface{}
	assert.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &allocations))
	assert.Greater(suite.T(), len(allocations), 1, "An editor with no tutor row should see all allocations")
}

func TestAllocationRouterTestSuite(t *testing.T) {
	suite.Run(t, new(AllocationRouterTestSuite))
}
