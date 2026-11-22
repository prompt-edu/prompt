package teams

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/team/teamDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	scopedTutorLogin = "ab12cde"
	teamAlphaID      = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	teamBetaID       = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	teamGammaID      = "cccccccc-cccc-cccc-cccc-cccccccccccc"
	projectLeadLogin = "pl34ead"
	projectLeadID    = "99999999-9999-9999-9999-999999999997"
)

// tutorAuthMiddleware mocks an authenticated CourseEditor (non-lecturer) so the tutor-scoping
// middleware actually resolves a team instead of taking the lecturer/admin pass-through branch.
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

// roleAuthMiddleware sets tokenUser and answers 403 unless the route admits one of the roles the
// user holds, as the SDK's authorization middleware does for the course roles.
func roleAuthMiddleware(tokenUser keycloakTokenVerifier.TokenUser) func(allowedRoles ...string) gin.HandlerFunc {
	return func(allowedRoles ...string) gin.HandlerFunc {
		return func(c *gin.Context) {
			admits := func(role string, held bool) bool { return held && slices.Contains(allowedRoles, role) }
			if !admits(promptSDK.CourseLecturer, tokenUser.IsLecturer) &&
				!admits(promptSDK.CourseEditor, tokenUser.IsEditor) &&
				!admits(promptSDK.CourseStudent, tokenUser.IsStudentOfCourse) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied"})
				return
			}
			keycloakTokenVerifier.SetTokenUser(c, tokenUser)
			c.Next()
		}
	}
}

func (suite *TeamRouterTestSuite) routerAs(login string) *gin.Engine {
	return suite.routerWith(tutorAuthMiddleware(login))
}

func (suite *TeamRouterTestSuite) routerWith(authMiddleware func(allowedRoles ...string) gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	api := router.Group("/api/course_phase/:coursePhaseID")
	RegisterRoutes(api, suite.teamService, authMiddleware)
	return router
}

type TeamRouterTestSuite struct {
	suite.Suite
	router      *gin.Engine
	suiteCtx    context.Context
	cleanup     func()
	teamService *TeamsService
}

func (suite *TeamRouterTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.suiteCtx, "../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../database_dumps/teams.sql")
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.teamService = NewTeamsService(*testDB.Queries, testDB.Conn)
	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "lecturer@example.com", "03711111", "ab12cde")
	}
	RegisterRoutes(api, suite.teamService, testMiddleware)
}

func (suite *TeamRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *TeamRouterTestSuite) TestGetAllTeams() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/team", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var response struct {
		Teams []promptTypes.Team `json:"teams"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(response.Teams), 0, "Should return a list of teams")
}

func (suite *TeamRouterTestSuite) TestGetAllTeamsInvalidCoursePhaseID() {
	req, _ := http.NewRequest("GET", "/api/course_phase/invalid-uuid/team", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *TeamRouterTestSuite) TestGetTeamByID() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	teamID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/team/"+teamID, nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var team promptTypes.Team
	err := json.Unmarshal(resp.Body.Bytes(), &team)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), team.ID, "Team ID should not be empty")
	assert.NotEmpty(suite.T(), team.Name, "Team name should not be empty")
}

func (suite *TeamRouterTestSuite) TestGetTeamByIDInvalidID() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/team/invalid-uuid", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *TeamRouterTestSuite) TestCreateTeams() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"

	createReq := teamDTO.CreateTeamsRequest{
		TeamNames: []string{"Team Epsilon", "Team Zeta"},
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+coursePhaseID+"/team", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
}

func (suite *TeamRouterTestSuite) TestCreateTeamsInvalidJSON() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	req, _ := http.NewRequest("POST", "/api/course_phase/"+coursePhaseID+"/team", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *TeamRouterTestSuite) TestUpdateTeamName() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	teamID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	updateReq := teamDTO.UpdateTeamRequest{
		NewTeamName: "Updated Team Alpha",
	}
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/course_phase/"+coursePhaseID+"/team/"+teamID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *TeamRouterTestSuite) TestUpdateTeamNameInvalidID() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"

	updateReq := teamDTO.UpdateTeamRequest{
		NewTeamName: "Updated Team",
	}
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/course_phase/"+coursePhaseID+"/team/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *TeamRouterTestSuite) TestDeleteTeam() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	teamID := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+coursePhaseID+"/team/"+teamID, nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *TeamRouterTestSuite) TestDeleteTeamInvalidID() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"

	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+coursePhaseID+"/team/invalid-uuid", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *TeamRouterTestSuite) TestTutorSeesOnlyOwnTeam() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	router := suite.routerAs(scopedTutorLogin)
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/team", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var response struct {
		Teams []promptTypes.Team `json:"teams"`
	}
	assert.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &response))
	assert.Len(suite.T(), response.Teams, 1, "Tutor should only see their assigned team")
	assert.Equal(suite.T(), teamAlphaID, response.Teams[0].ID.String())
}

func (suite *TeamRouterTestSuite) TestTutorAllowedOnOwnTeam() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	router := suite.routerAs(scopedTutorLogin)
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/team/"+teamAlphaID, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *TeamRouterTestSuite) TestTutorForbiddenOnOtherTeam() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	router := suite.routerAs(scopedTutorLogin)
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/team/"+teamBetaID, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusForbidden, resp.Code)
}

func (suite *TeamRouterTestSuite) TestEditorWithoutTutorRowForbiddenInPhaseWithTutors() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	for _, login := range []string{"zz99zzz", ""} {
		router := suite.routerAs(login)
		for _, path := range []string{"/team", "/team/" + teamAlphaID} {
			req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+path, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(suite.T(), http.StatusForbidden, resp.Code, "login %q on %s", login, path)
		}
	}
}

func (suite *TeamRouterTestSuite) TestEditorSeesAllTeamsInPhaseWithoutTutors() {
	coursePhaseID := "5179d58a-d00d-4fa7-94a5-397bc69fab03"
	router := suite.routerAs("zz99zzz")
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/team", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var response struct {
		Teams []promptTypes.Team `json:"teams"`
	}
	assert.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &response))
	assert.Len(suite.T(), response.Teams, 1, "An editor in a phase without tutors should see every team of the phase")
	assert.Equal(suite.T(), "dddddddd-dddd-dddd-dddd-dddddddddddd", response.Teams[0].ID.String())
}

func (suite *TeamRouterTestSuite) postStudentNames(pathPhaseID string, body string) int {
	req, _ := http.NewRequest("POST", "/api/course_phase/"+pathPhaseID+"/team/student-names", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	suite.router.ServeHTTP(resp, req)
	return resp.Code
}

func (suite *TeamRouterTestSuite) TestAddStudentNamesUsesPathCoursePhase() {
	allocatedPhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	participationID := uuid.MustParse("99999999-9999-9999-9999-999999999991")
	studentNames := `"studentNamesPerID": {"` + participationID.String() + `": {"firstName": "Renamed", "lastName": "Student"}}`

	code := suite.postStudentNames("5179d58a-d00d-4fa7-94a5-397bc69fab03", `{"coursePhaseID": "`+allocatedPhaseID.String()+`", `+studentNames+`}`)
	assert.Equal(suite.T(), http.StatusOK, code)

	allocation, err := suite.teamService.queries.GetAllocationForStudent(suite.suiteCtx, db.GetAllocationForStudentParams{
		CourseParticipationID: participationID,
		CoursePhaseID:         allocatedPhaseID,
	})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "John", allocation.StudentFirstName)

	code = suite.postStudentNames(allocatedPhaseID.String(), `{`+studentNames+`}`)
	assert.Equal(suite.T(), http.StatusOK, code)

	allocation, err = suite.teamService.queries.GetAllocationForStudent(suite.suiteCtx, db.GetAllocationForStudentParams{
		CourseParticipationID: participationID,
		CoursePhaseID:         allocatedPhaseID,
	})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Renamed", allocation.StudentFirstName)
}

func (suite *TeamRouterTestSuite) TestTutorOfSeveralTeamsSeesEachOfThem() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	router := suite.routerAs(projectLeadLogin)
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/team", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	var response struct {
		Teams []promptTypes.Team `json:"teams"`
	}
	assert.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &response))
	teamIDs := make([]string, 0, len(response.Teams))
	for _, team := range response.Teams {
		teamIDs = append(teamIDs, team.ID.String())
	}
	assert.ElementsMatch(suite.T(), []string{teamAlphaID, teamGammaID}, teamIDs)

	for teamID, expected := range map[string]int{teamAlphaID: http.StatusOK, teamGammaID: http.StatusOK, teamBetaID: http.StatusForbidden} {
		req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/team/"+teamID, nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(suite.T(), expected, resp.Code, "team %s", teamID)
	}
}

func (suite *TeamRouterTestSuite) TestTeamsListATutorUnderEachTeamTheyLead() {
	teams, err := suite.teamService.GetAllTeams(suite.suiteCtx, uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02"))
	assert.NoError(suite.T(), err)

	leadTeams := make([]string, 0)
	for _, team := range teams {
		for _, tutor := range team.Tutors {
			if tutor.ID == uuid.MustParse(projectLeadID) {
				leadTeams = append(leadTeams, team.ID.String())
			}
		}
	}
	assert.ElementsMatch(suite.T(), []string{teamAlphaID, teamGammaID}, leadTeams)
}

func (suite *TeamRouterTestSuite) postTutors(coursePhaseID string, tutors []teamDTO.Tutor) int {
	body, _ := json.Marshal(tutors)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+coursePhaseID+"/team/tutors", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	suite.router.ServeHTTP(resp, req)
	return resp.Code
}

func (suite *TeamRouterTestSuite) tutorTeams(coursePhaseID, tutorID string) []string {
	rows, err := suite.teamService.queries.GetTutorsByCourseParticipationID(suite.suiteCtx, db.GetTutorsByCourseParticipationIDParams{
		CourseParticipationID: uuid.MustParse(tutorID),
		CoursePhaseID:         uuid.MustParse(coursePhaseID),
	})
	assert.NoError(suite.T(), err)
	teamIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		teamIDs = append(teamIDs, row.TeamID.String())
	}
	return teamIDs
}

func (suite *TeamRouterTestSuite) TestImportTutorsSetsExactlyTheListedTeams() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	firstID := "99999999-9999-9999-9999-9999999999a1"
	tutor := teamDTO.Tutor{
		CourseParticipationID: uuid.MustParse(firstID),
		FirstName:             "Ida",
		LastName:              "Import",
		UniversityLogin:       " IM01PRT ",
		TeamIDs:               []uuid.UUID{uuid.MustParse(teamAlphaID), uuid.MustParse(teamGammaID), uuid.MustParse(teamGammaID)},
	}

	assert.Equal(suite.T(), http.StatusCreated, suite.postTutors(coursePhaseID, []teamDTO.Tutor{tutor}))
	assert.ElementsMatch(suite.T(), []string{teamAlphaID, teamGammaID}, suite.tutorTeams(coursePhaseID, firstID))

	// A re-import by the same login replaces the teams and keeps the id the tutor already has.
	tutor.CourseParticipationID = uuid.MustParse("99999999-9999-9999-9999-9999999999a2")
	tutor.TeamIDs = []uuid.UUID{uuid.MustParse(teamGammaID)}
	assert.Equal(suite.T(), http.StatusCreated, suite.postTutors(coursePhaseID, []teamDTO.Tutor{tutor}))
	assert.Equal(suite.T(), []string{teamGammaID}, suite.tutorTeams(coursePhaseID, firstID))
	assert.Empty(suite.T(), suite.tutorTeams(coursePhaseID, tutor.CourseParticipationID.String()))

	login, err := suite.teamService.queries.GetTutorsByCourseParticipationID(suite.suiteCtx, db.GetTutorsByCourseParticipationIDParams{
		CourseParticipationID: uuid.MustParse(firstID),
		CoursePhaseID:         uuid.MustParse(coursePhaseID),
	})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "im01prt", login[0].UniversityLogin.String, "the login is stored normalized")
}

func (suite *TeamRouterTestSuite) TestImportTutorsReplacesTheRowsOfATutorStoredWithoutLogin() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	// Like Bob Williams in the seed: a row stored without a login, so only its id matches it.
	tutorID := "99999999-9999-9999-9999-9999999999b1"
	assert.NoError(suite.T(), suite.teamService.queries.InsertTutor(suite.suiteCtx, db.InsertTutorParams{
		CoursePhaseID:         uuid.MustParse(coursePhaseID),
		CourseParticipationID: uuid.MustParse(tutorID),
		FirstName:             "Nola",
		LastName:              "Login",
		TeamID:                uuid.MustParse(teamAlphaID),
	}))
	tutor := teamDTO.Tutor{
		CourseParticipationID: uuid.MustParse(tutorID),
		FirstName:             "Nola",
		LastName:              "Login",
		UniversityLogin:       "nl01ogn",
		TeamIDs:               []uuid.UUID{uuid.MustParse(teamGammaID)},
	}

	assert.Equal(suite.T(), http.StatusCreated, suite.postTutors(coursePhaseID, []teamDTO.Tutor{tutor}))
	assert.Equal(suite.T(), []string{teamGammaID}, suite.tutorTeams(coursePhaseID, tutorID), "the login-less row is replaced")

	// Importing the tutor for their old team again must not collide with a leftover row.
	tutor.TeamIDs = []uuid.UUID{uuid.MustParse(teamAlphaID)}
	assert.Equal(suite.T(), http.StatusCreated, suite.postTutors(coursePhaseID, []teamDTO.Tutor{tutor}))
	assert.Equal(suite.T(), []string{teamAlphaID}, suite.tutorTeams(coursePhaseID, tutorID))
}

func (suite *TeamRouterTestSuite) TestImportTutorsKeepsTheTeamsOfAnotherLoginWithTheSameID() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	sharedID := "99999999-9999-9999-9999-9999999999c1"
	owner := teamDTO.Tutor{
		CourseParticipationID: uuid.MustParse(sharedID),
		FirstName:             "Olga",
		LastName:              "Owner",
		UniversityLogin:       "ow01ner",
		TeamIDs:               []uuid.UUID{uuid.MustParse(teamAlphaID)},
	}
	assert.Equal(suite.T(), http.StatusCreated, suite.postTutors(coursePhaseID, []teamDTO.Tutor{owner}))

	other := owner
	other.UniversityLogin = "ot01her"
	other.TeamIDs = []uuid.UUID{uuid.MustParse(teamGammaID)}
	assert.Equal(suite.T(), http.StatusCreated, suite.postTutors(coursePhaseID, []teamDTO.Tutor{other}))

	assert.ElementsMatch(suite.T(), []string{teamAlphaID, teamGammaID}, suite.tutorTeams(coursePhaseID, sharedID), "the owner keeps their team")
}

func (suite *TeamRouterTestSuite) TestImportTutorsRejectsInvalidRequests() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	valid := func(login string) teamDTO.Tutor {
		return teamDTO.Tutor{
			CourseParticipationID: uuid.New(),
			FirstName:             "Val",
			LastName:              "Id",
			UniversityLogin:       login,
			TeamIDs:               []uuid.UUID{uuid.MustParse(teamAlphaID)},
		}
	}

	noLogin := valid("  ")
	noTeams := valid("nt01eam")
	noTeams.TeamIDs = nil
	noID := valid("ni01did")
	noID.CourseParticipationID = uuid.Nil
	foreignTeam := valid("ft01eam")
	foreignTeam.TeamIDs = []uuid.UUID{uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")}
	sharedID := valid("si02dup")
	sharedID.CourseParticipationID = noTeams.CourseParticipationID
	noTeams2 := noTeams
	noTeams2.TeamIDs = []uuid.UUID{uuid.MustParse(teamAlphaID)}

	for name, tutors := range map[string][]teamDTO.Tutor{
		"missing login":         {noLogin},
		"no team":               {noTeams},
		"missing id":            {noID},
		"team of another phase": {foreignTeam},
		"login listed twice":    {valid("dp01lic"), valid("DP01LIC")},
		"id of two logins":      {noTeams2, sharedID},
	} {
		assert.Equal(suite.T(), http.StatusBadRequest, suite.postTutors(coursePhaseID, tutors), name)
	}

	assert.Empty(suite.T(), suite.tutorTeams(coursePhaseID, foreignTeam.CourseParticipationID.String()), "a refused import writes nothing")
}

func (suite *TeamRouterTestSuite) getTutorAssignments(authMiddleware func(allowedRoles ...string) gin.HandlerFunc) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", "/api/course_phase/4179d58a-d00d-4fa7-94a5-397bc69fab02/team/tutors", nil)
	resp := httptest.NewRecorder()
	suite.routerWith(authMiddleware).ServeHTTP(resp, req)
	return resp
}

func (suite *TeamRouterTestSuite) TestGetTutorAssignmentsListsEachTeamWithTheTutorLogin() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	// A row stored without a login is listed under its id with an empty login.
	loginlessID := uuid.MustParse("99999999-9999-9999-9999-9999999999c1")
	assert.NoError(suite.T(), suite.teamService.queries.InsertTutor(suite.suiteCtx, db.InsertTutorParams{
		CoursePhaseID:         uuid.MustParse(coursePhaseID),
		CourseParticipationID: loginlessID,
		FirstName:             "Lena",
		LastName:              "Less",
		TeamID:                uuid.MustParse(teamGammaID),
	}))

	resp := suite.getTutorAssignments(roleAuthMiddleware(keycloakTokenVerifier.TokenUser{IsLecturer: true}))
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var assignments []map[string]string
	assert.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &assignments))
	byTutor := map[string][]map[string]string{}
	for _, assignment := range assignments {
		assert.ElementsMatch(suite.T(), []string{"id", "universityLogin", "teamID"}, slices.Collect(maps.Keys(assignment)))
		byTutor[assignment["id"]] = append(byTutor[assignment["id"]], assignment)
	}
	assert.ElementsMatch(suite.T(), []map[string]string{
		{"id": projectLeadID, "universityLogin": projectLeadLogin, "teamID": teamAlphaID},
		{"id": projectLeadID, "universityLogin": projectLeadLogin, "teamID": teamGammaID},
	}, byTutor[projectLeadID])
	assert.Equal(suite.T(), []map[string]string{
		{"id": loginlessID.String(), "universityLogin": "", "teamID": teamGammaID},
	}, byTutor[loginlessID.String()])
}

func (suite *TeamRouterTestSuite) TestGetTutorAssignmentsIsForLecturersOnly() {
	for name, tokenUser := range map[string]keycloakTokenVerifier.TokenUser{
		"tutor":   {IsEditor: true, UniversityLogin: projectLeadLogin},
		"student": {IsStudentOfCourse: true, UniversityLogin: "st01ude"},
	} {
		resp := suite.getTutorAssignments(roleAuthMiddleware(tokenUser))
		assert.Equal(suite.T(), http.StatusForbidden, resp.Code, name)
		assert.NotContains(suite.T(), resp.Body.String(), projectLeadLogin, name)
	}
}

func TestTeamRouterTestSuite(t *testing.T) {
	suite.Run(t, new(TeamRouterTestSuite))
}

func (suite *TeamRouterTestSuite) getMyTutor(login string) *httptest.ResponseRecorder {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/team/tutors/me", nil)
	resp := httptest.NewRecorder()
	suite.routerAs(login).ServeHTTP(resp, req)
	return resp
}

func (suite *TeamRouterTestSuite) TestGetMyTutorMatchesUniversityLogin() {
	for _, login := range []string{scopedTutorLogin, " AB12CDE "} {
		resp := suite.getMyTutor(login)

		assert.Equal(suite.T(), http.StatusOK, resp.Code, "login %q", login)
		var tutor promptTypes.Person
		assert.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &tutor))
		assert.Equal(suite.T(), "99999999-9999-9999-9999-999999999993", tutor.ID.String())
		assert.Equal(suite.T(), "Alice", tutor.FirstName)
		assert.Equal(suite.T(), "Johnson", tutor.LastName)
	}
}

func (suite *TeamRouterTestSuite) TestGetMyTutorNotFoundForNonTutor() {
	for _, login := range []string{"zz99zzz", ""} {
		resp := suite.getMyTutor(login)

		assert.Equal(suite.T(), http.StatusNotFound, resp.Code, "login %q", login)
	}
}

const tutorPhaseID = "4179d58a-d00d-4fa7-94a5-397bc69fab02"

func (suite *TeamRouterTestSuite) postTeamTutor(router *gin.Engine, teamID string, tutor teamDTO.TeamTutor) *httptest.ResponseRecorder {
	body, _ := json.Marshal(tutor)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+tutorPhaseID+"/team/"+teamID+"/tutors", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func (suite *TeamRouterTestSuite) deleteTeamTutor(router *gin.Engine, teamID, tutorID string) int {
	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+tutorPhaseID+"/team/"+teamID+"/tutors/"+tutorID, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp.Code
}

func (suite *TeamRouterTestSuite) addedTutor(resp *httptest.ResponseRecorder) promptTypes.Person {
	var tutor promptTypes.Person
	assert.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &tutor))
	return tutor
}

func (suite *TeamRouterTestSuite) TestAddTutorToTeamKeepsTheIDOfAnExistingTutor() {
	tutorID := "99999999-9999-9999-9999-9999999999d1"
	tutor := teamDTO.TeamTutor{
		CourseParticipationID: uuid.MustParse(tutorID),
		FirstName:             "Tara",
		LastName:              "Team",
		UniversityLogin:       " TT01EAM ",
	}

	resp := suite.postTeamTutor(suite.router, teamAlphaID, tutor)
	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
	assert.Equal(suite.T(), tutorID, suite.addedTutor(resp).ID.String())

	// The same login under another id leads a second team under the id the tutor already has.
	tutor.CourseParticipationID = uuid.MustParse("99999999-9999-9999-9999-9999999999d2")
	resp = suite.postTeamTutor(suite.router, teamGammaID, tutor)
	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
	assert.Equal(suite.T(), tutorID, suite.addedTutor(resp).ID.String())

	assert.ElementsMatch(suite.T(), []string{teamAlphaID, teamGammaID}, suite.tutorTeams(tutorPhaseID, tutorID))
	assert.Empty(suite.T(), suite.tutorTeams(tutorPhaseID, tutor.CourseParticipationID.String()))

	stored, err := suite.teamService.GetTutorByUniversityLogin(suite.suiteCtx, uuid.MustParse(tutorPhaseID), "tt01eam")
	assert.NoError(suite.T(), err, "the login is stored normalized")
	assert.Equal(suite.T(), tutorID, stored.ID.String())
}

func (suite *TeamRouterTestSuite) TestAddTutorToTeamTwiceAddsNothing() {
	tutorID := "99999999-9999-9999-9999-9999999999d3"
	tutor := teamDTO.TeamTutor{
		CourseParticipationID: uuid.MustParse(tutorID),
		FirstName:             "Dina",
		LastName:              "Double",
		UniversityLogin:       "dd01ubl",
	}

	assert.Equal(suite.T(), http.StatusCreated, suite.postTeamTutor(suite.router, teamAlphaID, tutor).Code)
	resp := suite.postTeamTutor(suite.router, teamAlphaID, tutor)
	assert.Equal(suite.T(), http.StatusOK, resp.Code, "a tutor already on the team is not an error")
	assert.Equal(suite.T(), tutorID, suite.addedTutor(resp).ID.String())

	assert.Equal(suite.T(), []string{teamAlphaID}, suite.tutorTeams(tutorPhaseID, tutorID))
}

func (suite *TeamRouterTestSuite) TestAddTutorToTeamRejectsAnIDOnTheTeamUnderAnotherLogin() {
	// Like Bob Williams in the seed: a row stored without a login, so the new login does not match it.
	tutorID := "99999999-9999-9999-9999-9999999999d5"
	assert.NoError(suite.T(), suite.teamService.queries.InsertTutor(suite.suiteCtx, db.InsertTutorParams{
		CoursePhaseID:         uuid.MustParse(tutorPhaseID),
		CourseParticipationID: uuid.MustParse(tutorID),
		FirstName:             "Lou",
		LastName:              "Loginless",
		TeamID:                uuid.MustParse(teamGammaID),
	}))
	tutor := teamDTO.TeamTutor{
		CourseParticipationID: uuid.MustParse(tutorID),
		FirstName:             "Lou",
		LastName:              "Loginless",
		UniversityLogin:       "ll01ogn",
	}

	assert.Equal(suite.T(), http.StatusBadRequest, suite.postTeamTutor(suite.router, teamGammaID, tutor).Code,
		"the id already leads the team, so the tutor is not reported as added")
	_, err := suite.teamService.GetTutorByUniversityLogin(suite.suiteCtx, uuid.MustParse(tutorPhaseID), "ll01ogn")
	assert.ErrorIs(suite.T(), err, ErrTutorNotFound)
}

func (suite *TeamRouterTestSuite) TestAddTutorToTeamRejectsInvalidRequests() {
	valid := teamDTO.TeamTutor{
		CourseParticipationID: uuid.MustParse("99999999-9999-9999-9999-9999999999d4"),
		FirstName:             "Val",
		LastName:              "Id",
		UniversityLogin:       "vi01dtm",
	}
	noLogin := valid
	noLogin.UniversityLogin = "  "
	noID := valid
	noID.CourseParticipationID = uuid.Nil

	for name, tc := range map[string]struct {
		teamID string
		tutor  teamDTO.TeamTutor
	}{
		"missing login":         {teamAlphaID, noLogin},
		"missing id":            {teamAlphaID, noID},
		"unknown team":          {uuid.NewString(), valid},
		"team of another phase": {"dddddddd-dddd-dddd-dddd-dddddddddddd", valid},
	} {
		assert.Equal(suite.T(), http.StatusBadRequest, suite.postTeamTutor(suite.router, tc.teamID, tc.tutor).Code, name)
	}

	req, _ := http.NewRequest("POST", "/api/course_phase/"+tutorPhaseID+"/team/"+teamAlphaID+"/tutors", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code, "invalid JSON")

	assert.Empty(suite.T(), suite.tutorTeams(tutorPhaseID, valid.CourseParticipationID.String()), "a refused request writes nothing")
}

func (suite *TeamRouterTestSuite) TestRemoveTutorFromOneOfSeveralTeams() {
	tutorID := "99999999-9999-9999-9999-9999999999e1"
	tutor := teamDTO.TeamTutor{
		CourseParticipationID: uuid.MustParse(tutorID),
		FirstName:             "Sev",
		LastName:              "Eral",
		UniversityLogin:       "se01vrl",
	}
	assert.Equal(suite.T(), http.StatusCreated, suite.postTeamTutor(suite.router, teamAlphaID, tutor).Code)
	assert.Equal(suite.T(), http.StatusCreated, suite.postTeamTutor(suite.router, teamGammaID, tutor).Code)

	assert.Equal(suite.T(), http.StatusOK, suite.deleteTeamTutor(suite.router, teamAlphaID, tutorID))

	assert.Equal(suite.T(), []string{teamGammaID}, suite.tutorTeams(tutorPhaseID, tutorID))
	_, err := suite.teamService.GetTutorByUniversityLogin(suite.suiteCtx, uuid.MustParse(tutorPhaseID), "se01vrl")
	assert.NoError(suite.T(), err, "the tutor still leads a team")
}

func (suite *TeamRouterTestSuite) TestRemoveTutorFromTheirLastTeamRemovesTheTutor() {
	tutorID := "99999999-9999-9999-9999-9999999999e2"
	tutor := teamDTO.TeamTutor{
		CourseParticipationID: uuid.MustParse(tutorID),
		FirstName:             "Lana",
		LastName:              "Last",
		UniversityLogin:       "la01stt",
	}
	assert.Equal(suite.T(), http.StatusCreated, suite.postTeamTutor(suite.router, teamGammaID, tutor).Code)

	assert.Equal(suite.T(), http.StatusOK, suite.deleteTeamTutor(suite.router, teamGammaID, tutorID))

	assert.Empty(suite.T(), suite.tutorTeams(tutorPhaseID, tutorID))
	_, err := suite.teamService.GetTutorByUniversityLogin(suite.suiteCtx, uuid.MustParse(tutorPhaseID), "la01stt")
	assert.ErrorIs(suite.T(), err, ErrTutorNotFound, "a tutor without a team is no longer a tutor of the phase")

	resp := suite.getTutorAssignments(roleAuthMiddleware(keycloakTokenVerifier.TokenUser{IsLecturer: true}))
	assert.NotContains(suite.T(), resp.Body.String(), tutorID)
}

func (suite *TeamRouterTestSuite) TestRemoveTutorFromTeamRejectsUnknownTeamsAndTutors() {
	tutorID := "99999999-9999-9999-9999-9999999999e3"
	tutor := teamDTO.TeamTutor{
		CourseParticipationID: uuid.MustParse(tutorID),
		FirstName:             "Not",
		LastName:              "Here",
		UniversityLogin:       "nh01ere",
	}
	assert.Equal(suite.T(), http.StatusCreated, suite.postTeamTutor(suite.router, teamAlphaID, tutor).Code)

	assert.Equal(suite.T(), http.StatusNotFound, suite.deleteTeamTutor(suite.router, teamGammaID, tutorID), "a team the tutor does not lead")
	assert.Equal(suite.T(), http.StatusNotFound, suite.deleteTeamTutor(suite.router, teamAlphaID, uuid.NewString()), "an unknown tutor")
	assert.Equal(suite.T(), http.StatusBadRequest, suite.deleteTeamTutor(suite.router, uuid.NewString(), tutorID), "an unknown team")
	assert.Equal(suite.T(), http.StatusBadRequest, suite.deleteTeamTutor(suite.router, "dddddddd-dddd-dddd-dddd-dddddddddddd", tutorID), "a team of another phase")
	assert.Equal(suite.T(), http.StatusBadRequest, suite.deleteTeamTutor(suite.router, teamAlphaID, "not-a-uuid"), "an invalid tutor id")

	assert.Equal(suite.T(), []string{teamAlphaID}, suite.tutorTeams(tutorPhaseID, tutorID))
}

func (suite *TeamRouterTestSuite) TestTeamTutorRoutesAreForLecturersOnly() {
	tutorID := "99999999-9999-9999-9999-9999999999f1"
	tutor := teamDTO.TeamTutor{
		CourseParticipationID: uuid.MustParse(tutorID),
		FirstName:             "Rhea",
		LastName:              "Role",
		UniversityLogin:       "rr01ole",
	}

	for name, tokenUser := range map[string]keycloakTokenVerifier.TokenUser{
		"tutor":   {IsEditor: true, UniversityLogin: projectLeadLogin},
		"student": {IsStudentOfCourse: true, UniversityLogin: "st01ude"},
	} {
		router := suite.routerWith(roleAuthMiddleware(tokenUser))
		assert.Equal(suite.T(), http.StatusForbidden, suite.postTeamTutor(router, teamAlphaID, tutor).Code, name)
		assert.Equal(suite.T(), http.StatusForbidden, suite.deleteTeamTutor(router, teamAlphaID, projectLeadID), name)
	}
	assert.Empty(suite.T(), suite.tutorTeams(tutorPhaseID, tutorID), "a forbidden add writes nothing")
	assert.Contains(suite.T(), suite.tutorTeams(tutorPhaseID, projectLeadID), teamAlphaID, "a forbidden removal removes nothing")

	lecturer := suite.routerWith(roleAuthMiddleware(keycloakTokenVerifier.TokenUser{IsLecturer: true}))
	assert.Equal(suite.T(), http.StatusCreated, suite.postTeamTutor(lecturer, teamAlphaID, tutor).Code)
	assert.Equal(suite.T(), http.StatusOK, suite.deleteTeamTutor(lecturer, teamAlphaID, tutorID))
}
