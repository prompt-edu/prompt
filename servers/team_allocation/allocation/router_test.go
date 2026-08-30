package allocation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/team_allocation/allocation/allocationDTO"
	"github.com/prompt-edu/prompt/servers/team_allocation/coreRequests"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	scopedTutorLogin  = "ab12cde"
	teamAlphaParticip = "99999999-9999-9999-9999-999999999991"
	teamBetaParticip  = "99999999-9999-9999-9999-999999999992"

	// The write fixtures live in their own course phase so allocation writes cannot
	// disturb the read assertions above, whatever order the suite runs in.
	writePhase         = "5179d58a-d00d-4fa7-94a5-397bc69fab03"
	teamDelta          = "dddddddd-dddd-dddd-dddd-dddddddddddd"
	teamEpsilon        = "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	tutorFreeParticip  = "99999999-9999-9999-9999-999999999995"
	deltaParticip      = "99999999-9999-9999-9999-999999999996"
	epsilonParticip    = "99999999-9999-9999-9999-999999999997"
	staffFreeParticip  = "99999999-9999-9999-9999-99999999999a"
	unknownParticipant = "99999999-9999-9999-9999-9999999999ff"
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

// lecturerAuthMiddleware mocks a course lecturer: the scoping middleware passes
// them through, so they must keep unrestricted write access.
func lecturerAuthMiddleware() func(allowedRoles ...string) gin.HandlerFunc {
	return func(allowedRoles ...string) gin.HandlerFunc {
		return func(c *gin.Context) {
			keycloakTokenVerifier.SetTokenUser(c, keycloakTokenVerifier.TokenUser{
				IsEditor:   false,
				IsLecturer: true,
			})
			c.Next()
		}
	}
}

// promptLecturerTutorAuthMiddleware mocks a tutor who also holds the global
// PROMPT_Lecturer role. The write routes do not admit that role directly, so the
// user arrives as a course editor and must stay scoped to their team.
func promptLecturerTutorAuthMiddleware(login string) func(allowedRoles ...string) gin.HandlerFunc {
	return func(allowedRoles ...string) gin.HandlerFunc {
		return func(c *gin.Context) {
			keycloakTokenVerifier.SetTokenUser(c, keycloakTokenVerifier.TokenUser{
				Roles:           map[string]bool{keycloakTokenVerifier.PromptLecturer: true},
				IsEditor:        true,
				IsLecturer:      false,
				UniversityLogin: login,
			})
			c.Next()
		}
	}
}

// anonymousAuthMiddleware leaves no token user in the context.
func anonymousAuthMiddleware() func(allowedRoles ...string) gin.HandlerFunc {
	return func(allowedRoles ...string) gin.HandlerFunc {
		return func(c *gin.Context) { c.Next() }
	}
}

// stubParticipants replaces the core lookup with a fixed membership set.
func stubParticipants(courseParticipationIDs ...string) participantResolver {
	return func(authHeader string, coursePhaseID uuid.UUID) (map[uuid.UUID]coreRequests.Participant, error) {
		participants := make(map[uuid.UUID]coreRequests.Participant, len(courseParticipationIDs))
		for _, id := range courseParticipationIDs {
			participants[uuid.MustParse(id)] = coreRequests.Participant{FirstName: "Test", LastName: "Student"}
		}
		return participants, nil
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
		queries:             *testDB.Queries,
		conn:                testDB.Conn,
		resolveParticipants: stubParticipants(tutorFreeParticip, deltaParticip, epsilonParticip, staffFreeParticip),
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

	var allocation allocationDTO.Allocation
	err := json.Unmarshal(resp.Body.Bytes(), &allocation)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), allocation.TeamAllocation)
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

func (suite *AllocationRouterTestSuite) putAllocation(router *gin.Engine, coursePhaseID, courseParticipationID, body string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(http.MethodPut, "/api/course_phase/"+coursePhaseID+"/allocation/"+courseParticipationID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func (suite *AllocationRouterTestSuite) deleteAllocation(router *gin.Engine, coursePhaseID, courseParticipationID string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(http.MethodDelete, "/api/course_phase/"+coursePhaseID+"/allocation/"+courseParticipationID, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func (suite *AllocationRouterTestSuite) storedTeam(courseParticipationID string) (uuid.UUID, error) {
	return GetAllocationByCourseParticipationID(suite.suiteCtx, uuid.MustParse(courseParticipationID), uuid.MustParse(writePhase))
}

func (suite *AllocationRouterTestSuite) TestAdminAssignsParticipantToAnyTeam() {
	resp := suite.putAllocation(suite.router, writePhase, staffFreeParticip, `{"teamID":"`+teamEpsilon+`"}`)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	teamID, err := suite.storedTeam(staffFreeParticip)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uuid.MustParse(teamEpsilon), teamID)
}

func (suite *AllocationRouterTestSuite) TestCourseLecturerMovesParticipantBetweenTeams() {
	router := gin.New()
	api := router.Group("/api/course_phase/:coursePhaseID")
	setupAllocationRouter(api, lecturerAuthMiddleware(), suite.allocationService.queries)

	resp := suite.putAllocation(router, writePhase, epsilonParticip, `{"teamID":"`+teamDelta+`"}`)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	teamID, err := suite.storedTeam(epsilonParticip)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uuid.MustParse(teamDelta), teamID)

	// Restore the fixture so the tutor cases below still start from a foreign team.
	restore := suite.putAllocation(router, writePhase, epsilonParticip, `{"teamID":"`+teamEpsilon+`"}`)
	assert.Equal(suite.T(), http.StatusOK, restore.Code)
}

func (suite *AllocationRouterTestSuite) TestTutorAssignsUnallocatedParticipantToOwnTeam() {
	router := suite.routerAs(scopedTutorLogin)
	resp := suite.putAllocation(router, writePhase, tutorFreeParticip, `{"teamID":"`+teamDelta+`"}`)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var allocation allocationDTO.Allocation
	assert.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &allocation))
	assert.Equal(suite.T(), uuid.MustParse(teamDelta), allocation.TeamAllocation)

	stored, err := suite.allocationService.queries.GetAllocationForStudent(suite.suiteCtx, db.GetAllocationForStudentParams{
		CourseParticipationID: uuid.MustParse(tutorFreeParticip),
		CoursePhaseID:         uuid.MustParse(writePhase),
	})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Test", stored.StudentFirstName, "the participant's name is cached on the allocation")
	assert.Equal(suite.T(), "Student", stored.StudentLastName)
}

func (suite *AllocationRouterTestSuite) TestTutorCannotAssignToForeignTeam() {
	router := suite.routerAs(scopedTutorLogin)
	resp := suite.putAllocation(router, writePhase, tutorFreeParticip, `{"teamID":"`+teamEpsilon+`"}`)
	assert.Equal(suite.T(), http.StatusForbidden, resp.Code)
}

func (suite *AllocationRouterTestSuite) TestTutorCannotPullParticipantFromForeignTeam() {
	router := suite.routerAs(scopedTutorLogin)
	resp := suite.putAllocation(router, writePhase, epsilonParticip, `{"teamID":"`+teamDelta+`"}`)
	assert.Equal(suite.T(), http.StatusForbidden, resp.Code)

	teamID, err := suite.storedTeam(epsilonParticip)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uuid.MustParse(teamEpsilon), teamID, "the foreign team must be untouched")
}

func (suite *AllocationRouterTestSuite) TestNonTutorEditorCannotWrite() {
	router := suite.routerAs("zz99zzz")
	resp := suite.putAllocation(router, writePhase, staffFreeParticip, `{"teamID":"`+teamDelta+`"}`)
	assert.Equal(suite.T(), http.StatusForbidden, resp.Code, "an editor without a tutor record must not write")

	deleteResp := suite.deleteAllocation(router, writePhase, epsilonParticip)
	assert.Equal(suite.T(), http.StatusForbidden, deleteResp.Code)
}

func (suite *AllocationRouterTestSuite) TestTutorCannotDeleteForeignTeamAllocation() {
	router := suite.routerAs(scopedTutorLogin)
	resp := suite.deleteAllocation(router, writePhase, epsilonParticip)
	assert.Equal(suite.T(), http.StatusForbidden, resp.Code)

	_, err := suite.storedTeam(epsilonParticip)
	assert.NoError(suite.T(), err, "the foreign allocation must still exist")
}

func (suite *AllocationRouterTestSuite) TestTutorDeletesOwnTeamAllocation() {
	router := suite.routerAs(scopedTutorLogin)
	resp := suite.deleteAllocation(router, writePhase, deltaParticip)
	assert.Equal(suite.T(), http.StatusNoContent, resp.Code)

	_, err := suite.storedTeam(deltaParticip)
	assert.Error(suite.T(), err, "the allocation must be gone")
}

func (suite *AllocationRouterTestSuite) TestDeleteUnknownAllocationReturnsNotFound() {
	resp := suite.deleteAllocation(suite.router, writePhase, unknownParticipant)
	assert.Equal(suite.T(), http.StatusNotFound, resp.Code)
}

func (suite *AllocationRouterTestSuite) TestUpdateAllocationRejectsTeamFromAnotherPhase() {
	resp := suite.putAllocation(suite.router, writePhase, staffFreeParticip, `{"teamID":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}`)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *AllocationRouterTestSuite) TestUpdateAllocationRejectsNonParticipant() {
	resp := suite.putAllocation(suite.router, writePhase, unknownParticipant, `{"teamID":"`+teamDelta+`"}`)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)

	_, err := suite.storedTeam(unknownParticipant)
	assert.Error(suite.T(), err, "a rejected write must not create an allocation")
}

func (suite *AllocationRouterTestSuite) TestUpdateAllocationFailsClosedWhenCoreIsUnavailable() {
	original := suite.allocationService.resolveParticipants
	suite.allocationService.resolveParticipants = func(authHeader string, coursePhaseID uuid.UUID) (map[uuid.UUID]coreRequests.Participant, error) {
		return nil, errors.New("core unreachable")
	}
	defer func() { suite.allocationService.resolveParticipants = original }()

	resp := suite.putAllocation(suite.router, writePhase, unknownParticipant, `{"teamID":"`+teamDelta+`"}`)
	assert.Equal(suite.T(), http.StatusBadGateway, resp.Code)

	_, err := suite.storedTeam(unknownParticipant)
	assert.Error(suite.T(), err, "an unverified write must not reach the database")
}

func (suite *AllocationRouterTestSuite) TestUpdateAllocationRejectsOversizedBody() {
	body := `{"teamID":"` + teamDelta + `","padding":"` + strings.Repeat("x", maxAllocationBodyBytes) + `"}`
	resp := suite.putAllocation(suite.router, writePhase, staffFreeParticip, body)
	assert.Equal(suite.T(), http.StatusRequestEntityTooLarge, resp.Code)
}

func (suite *AllocationRouterTestSuite) TestUpdateAllocationRejectsMissingTeamID() {
	resp := suite.putAllocation(suite.router, writePhase, staffFreeParticip, `{}`)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *AllocationRouterTestSuite) TestUpdateAllocationRejectsInvalidIDs() {
	resp := suite.putAllocation(suite.router, "invalid-uuid", staffFreeParticip, `{"teamID":"`+teamDelta+`"}`)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)

	nilParticipation := suite.putAllocation(suite.router, writePhase, uuid.Nil.String(), `{"teamID":"`+teamDelta+`"}`)
	assert.Equal(suite.T(), http.StatusBadRequest, nilParticipation.Code)
}

func (suite *AllocationRouterTestSuite) routerWith(authMiddleware func(allowedRoles ...string) gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	api := router.Group("/api/course_phase/:coursePhaseID")
	setupAllocationRouter(api, authMiddleware, suite.allocationService.queries)
	return router
}

func (suite *AllocationRouterTestSuite) TestPromptLecturerTutorStaysScoped() {
	router := suite.routerWith(promptLecturerTutorAuthMiddleware(scopedTutorLogin))

	resp := suite.putAllocation(router, writePhase, epsilonParticip, `{"teamID":"`+teamDelta+`"}`)
	assert.Equal(suite.T(), http.StatusForbidden, resp.Code, "the global lecturer role must not unscope a tutor's writes")

	teamID, err := suite.storedTeam(epsilonParticip)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uuid.MustParse(teamEpsilon), teamID)
}

func (suite *AllocationRouterTestSuite) TestWriteWithoutTokenUserIsUnauthorized() {
	router := suite.routerWith(anonymousAuthMiddleware())

	resp := suite.putAllocation(router, writePhase, staffFreeParticip, `{"teamID":"`+teamDelta+`"}`)
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.Code)

	deleteResp := suite.deleteAllocation(router, writePhase, epsilonParticip)
	assert.Equal(suite.T(), http.StatusUnauthorized, deleteResp.Code)
}

func TestAllocationRouterTestSuite(t *testing.T) {
	suite.Run(t, new(AllocationRouterTestSuite))
}
