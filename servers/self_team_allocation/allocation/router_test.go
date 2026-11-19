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
	"github.com/prompt-edu/prompt/servers/self_team_allocation/allocation/allocationDTO"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AllocationRouterTestSuite struct {
	suite.Suite
	ctx     context.Context
	router  *gin.Engine
	service *AllocationService
	cleanup func()
}

func (suite *AllocationRouterTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()

	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.ctx, "../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../database_dumps/base.sql")
	require.NoError(suite.T(), err)
	suite.cleanup = cleanup

	allocationService := NewAllocationService(*testDB.Queries, testDB.Conn)
	suite.service = allocationService

	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	authMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.DefaultMockAuthMiddleware()
	}
	RegisterRoutes(api, allocationService, authMiddleware)
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

	require.Equal(suite.T(), http.StatusNotFound, resp.Code)
}

// tutorRouter authenticates every request as a course editor with the given login, so the tutor
// scope resolves their teams instead of passing a lecturer through.
func (suite *AllocationRouterTestSuite) tutorRouter(login string) *gin.Engine {
	router := gin.Default()
	api := router.Group("/api/course_phase/:coursePhaseID")
	RegisterRoutes(api, suite.service, func(allowedRoles ...string) gin.HandlerFunc {
		return func(c *gin.Context) {
			keycloakTokenVerifier.SetTokenUser(c, keycloakTokenVerifier.TokenUser{IsEditor: true, UniversityLogin: login})
			c.Next()
		}
	})
	return router
}

func (suite *AllocationRouterTestSuite) TestTutorOfSeveralTeamsReadsOnlyTheirAllocations() {
	const phase = "/api/course_phase/33333333-3333-3333-3333-333333333333/allocation"
	router := suite.tutorRouter("pl34ead")

	req, _ := http.NewRequest("GET", phase, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)
	var allocations []allocationDTO.AllocationWithParticipation
	require.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &allocations))
	teams := make([]string, 0, len(allocations))
	for _, allocation := range allocations {
		teams = append(teams, allocation.TeamAllocation.String())
	}
	require.ElementsMatch(suite.T(), []string{"dddddddd-0000-0000-0000-000000000001", "dddddddd-0000-0000-0000-000000000002"}, teams)

	for participant, expected := range map[string]int{
		"dddd0001-0000-0000-0000-000000000001": http.StatusOK,
		"dddd0003-0000-0000-0000-000000000003": http.StatusForbidden,
	} {
		req, _ := http.NewRequest("GET", phase+"/"+participant, nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		require.Equal(suite.T(), expected, resp.Code, participant)
	}
}

func (suite *AllocationRouterTestSuite) TestEditorWithoutTutorRowForbiddenInPhaseWithTutors() {
	req, _ := http.NewRequest("GET", "/api/course_phase/33333333-3333-3333-3333-333333333333/allocation", nil)
	resp := httptest.NewRecorder()
	suite.tutorRouter("zz99zzz").ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusForbidden, resp.Code)
}

func TestAllocationRouterTestSuite(t *testing.T) {
	suite.Run(t, new(AllocationRouterTestSuite))
}
