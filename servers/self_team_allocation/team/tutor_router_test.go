package teams

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/team/teamDTO"
	"github.com/stretchr/testify/require"
)

const (
	tutorScopePhaseID = "33333333-3333-3333-3333-333333333333"
	noTutorPhaseID    = "22222222-2222-2222-2222-222222222222"
	lovelaceTeamID    = "dddddddd-0000-0000-0000-000000000001"
	turingTeamID      = "dddddddd-0000-0000-0000-000000000002"
	hamiltonTeamID    = "dddddddd-0000-0000-0000-000000000003"
	projectLeadLogin  = "pl34ead"
	projectLeadID     = "99999999-9999-9999-9999-999999999997"
)

// routerAs answers 403 unless the route admits one of the course roles tokenUser holds, as the
// SDK's authorization middleware does.
func (suite *TeamRouterTestSuite) routerAs(tokenUser keycloakTokenVerifier.TokenUser) *gin.Engine {
	router := gin.Default()
	api := router.Group("/api/course_phase/:coursePhaseID")
	RegisterRoutes(api, suite.teamsService, suite.assignmentService, func(allowedRoles ...string) gin.HandlerFunc {
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
	})
	return router
}

func tutor(login string) keycloakTokenVerifier.TokenUser {
	return keycloakTokenVerifier.TokenUser{IsEditor: true, UniversityLogin: login}
}

func (suite *TeamRouterTestSuite) get(router *gin.Engine, coursePhaseID, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+path, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func (suite *TeamRouterTestSuite) teamIDsOf(resp *httptest.ResponseRecorder) []string {
	var response struct {
		Teams []promptTypes.Team `json:"teams"`
	}
	require.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &response))
	ids := make([]string, 0, len(response.Teams))
	for _, team := range response.Teams {
		ids = append(ids, team.ID.String())
	}
	return ids
}

func (suite *TeamRouterTestSuite) postTutors(coursePhaseID uuid.UUID, tutors []teamDTO.Tutor) int {
	body, _ := json.Marshal(tutors)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+coursePhaseID.String()+"/team/tutors", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	suite.routerAs(keycloakTokenVerifier.TokenUser{IsLecturer: true}).ServeHTTP(resp, req)
	return resp.Code
}

func (suite *TeamRouterTestSuite) tutorTeams(coursePhaseID, tutorID uuid.UUID) []uuid.UUID {
	rows, err := suite.testDB.Queries.GetTutorAssignmentsByCoursePhase(suite.ctx, coursePhaseID)
	require.NoError(suite.T(), err)
	teamIDs := make([]uuid.UUID, 0)
	for _, row := range rows {
		if row.CourseParticipationID == tutorID {
			teamIDs = append(teamIDs, row.TeamID)
		}
	}
	return teamIDs
}

func (suite *TeamRouterTestSuite) TestTutorOfSeveralTeamsReadsExactlyThem() {
	router := suite.routerAs(tutor(projectLeadLogin))

	resp := suite.get(router, tutorScopePhaseID, "/team")
	require.Equal(suite.T(), http.StatusOK, resp.Code)
	require.ElementsMatch(suite.T(), []string{lovelaceTeamID, turingTeamID}, suite.teamIDsOf(resp))

	for teamID, expected := range map[string]int{lovelaceTeamID: http.StatusOK, turingTeamID: http.StatusOK, hamiltonTeamID: http.StatusForbidden} {
		require.Equal(suite.T(), expected, suite.get(router, tutorScopePhaseID, "/team/"+teamID).Code, "team %s", teamID)
	}
}

func (suite *TeamRouterTestSuite) TestEditorWithoutTutorRowForbiddenInPhaseWithTutors() {
	for _, login := range []string{"zz99zzz", ""} {
		router := suite.routerAs(tutor(login))
		for _, path := range []string{"/team", "/team/" + lovelaceTeamID} {
			require.Equal(suite.T(), http.StatusForbidden, suite.get(router, tutorScopePhaseID, path).Code, "login %q on %s", login, path)
		}
	}
}

func (suite *TeamRouterTestSuite) TestEditorSeesAllTeamsInPhaseWithoutTutors() {
	resp := suite.get(suite.routerAs(tutor("zz99zzz")), noTutorPhaseID, "/team")

	require.Equal(suite.T(), http.StatusOK, resp.Code)
	require.Equal(suite.T(), []string{"cccccccc-cccc-cccc-cccc-cccccccccccc"}, suite.teamIDsOf(resp))
}

func (suite *TeamRouterTestSuite) TestImportTutorsSetsExactlyTheListedTeams() {
	coursePhaseID := uuid.New()
	alpha := suite.createTeam(coursePhaseID, "Alpha")
	gamma := suite.createTeam(coursePhaseID, "Gamma")
	firstID := uuid.New()
	imported := teamDTO.Tutor{
		CourseParticipationID: firstID,
		FirstName:             "Ida",
		LastName:              "Import",
		UniversityLogin:       " IM01PRT ",
		TeamIDs:               []uuid.UUID{alpha, gamma, gamma},
	}

	require.Equal(suite.T(), http.StatusCreated, suite.postTutors(coursePhaseID, []teamDTO.Tutor{imported}))
	require.ElementsMatch(suite.T(), []uuid.UUID{alpha, gamma}, suite.tutorTeams(coursePhaseID, firstID))

	// A re-import by the same login replaces the teams and keeps the id the tutor already has.
	imported.CourseParticipationID = uuid.New()
	imported.TeamIDs = []uuid.UUID{gamma}
	require.Equal(suite.T(), http.StatusCreated, suite.postTutors(coursePhaseID, []teamDTO.Tutor{imported}))
	require.Equal(suite.T(), []uuid.UUID{gamma}, suite.tutorTeams(coursePhaseID, firstID))
	require.Empty(suite.T(), suite.tutorTeams(coursePhaseID, imported.CourseParticipationID))

	stored, err := suite.testDB.Queries.GetTutorByUniversityLogin(suite.ctx, db.GetTutorByUniversityLoginParams{
		CoursePhaseID:   coursePhaseID,
		UniversityLogin: teamDTO.UniversityLoginParam("im01prt"),
	})
	require.NoError(suite.T(), err, "the login is stored normalized")
	require.Equal(suite.T(), firstID, stored.CourseParticipationID)
}

func (suite *TeamRouterTestSuite) TestImportTutorsReplacesTheRowsOfATutorStoredWithoutLogin() {
	coursePhaseID := uuid.New()
	alpha := suite.createTeam(coursePhaseID, "Alpha")
	gamma := suite.createTeam(coursePhaseID, "Gamma")
	tutorID := uuid.New()
	require.NoError(suite.T(), suite.testDB.Queries.CreateTutor(suite.ctx, db.CreateTutorParams{
		CoursePhaseID:         coursePhaseID,
		CourseParticipationID: tutorID,
		FirstName:             "Nola",
		LastName:              "Login",
		TeamID:                alpha,
	}))
	imported := teamDTO.Tutor{CourseParticipationID: tutorID, FirstName: "Nola", LastName: "Login", UniversityLogin: "nl01ogn", TeamIDs: []uuid.UUID{gamma}}

	require.Equal(suite.T(), http.StatusCreated, suite.postTutors(coursePhaseID, []teamDTO.Tutor{imported}))
	require.Equal(suite.T(), []uuid.UUID{gamma}, suite.tutorTeams(coursePhaseID, tutorID), "the login-less row is replaced")

	imported.TeamIDs = []uuid.UUID{alpha}
	require.Equal(suite.T(), http.StatusCreated, suite.postTutors(coursePhaseID, []teamDTO.Tutor{imported}))
	require.Equal(suite.T(), []uuid.UUID{alpha}, suite.tutorTeams(coursePhaseID, tutorID))
}

func (suite *TeamRouterTestSuite) TestImportTutorsRejectsInvalidRequests() {
	coursePhaseID := uuid.New()
	alpha := suite.createTeam(coursePhaseID, "Alpha")
	foreignTeam := suite.createTeam(uuid.New(), "Foreign")
	valid := func(login string) teamDTO.Tutor {
		return teamDTO.Tutor{CourseParticipationID: uuid.New(), FirstName: "Val", LastName: "Id", UniversityLogin: login, TeamIDs: []uuid.UUID{alpha}}
	}

	noTeams := valid("nt01eam")
	noTeams.TeamIDs = nil
	noID := valid("ni01did")
	noID.CourseParticipationID = uuid.Nil
	ofForeignTeam := valid("ft01eam")
	ofForeignTeam.TeamIDs = []uuid.UUID{foreignTeam}
	sharedID := valid("si02dup")
	sharedID.CourseParticipationID = alpha

	for name, tutors := range map[string][]teamDTO.Tutor{
		"missing login":         {valid("  ")},
		"no team":               {noTeams},
		"missing id":            {noID},
		"team of another phase": {ofForeignTeam},
		"login listed twice":    {valid("dp01lic"), valid("DP01LIC")},
		"id of two logins":      {sharedID, {CourseParticipationID: alpha, UniversityLogin: "si03dup", TeamIDs: []uuid.UUID{alpha}}},
	} {
		require.Equal(suite.T(), http.StatusBadRequest, suite.postTutors(coursePhaseID, tutors), name)
	}

	rows, err := suite.testDB.Queries.GetTutorAssignmentsByCoursePhase(suite.ctx, coursePhaseID)
	require.NoError(suite.T(), err)
	require.Empty(suite.T(), rows, "a refused import writes nothing")
}

func (suite *TeamRouterTestSuite) TestGetTutorAssignmentsListsEachTeamWithTheTutorLogin() {
	resp := suite.get(suite.routerAs(keycloakTokenVerifier.TokenUser{IsLecturer: true}), tutorScopePhaseID, "/team/tutors")
	require.Equal(suite.T(), http.StatusOK, resp.Code)

	var assignments []map[string]string
	require.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &assignments))
	require.ElementsMatch(suite.T(), []map[string]string{
		{"id": projectLeadID, "universityLogin": projectLeadLogin, "teamID": lovelaceTeamID},
		{"id": projectLeadID, "universityLogin": projectLeadLogin, "teamID": turingTeamID},
	}, assignments)
}

func (suite *TeamRouterTestSuite) TestGetTutorAssignmentsIsForLecturersOnly() {
	for name, tokenUser := range map[string]keycloakTokenVerifier.TokenUser{
		"tutor":   tutor(projectLeadLogin),
		"student": {IsStudentOfCourse: true, UniversityLogin: "st01ude"},
	} {
		resp := suite.get(suite.routerAs(tokenUser), tutorScopePhaseID, "/team/tutors")
		require.Equal(suite.T(), http.StatusForbidden, resp.Code, name)
		require.NotContains(suite.T(), resp.Body.String(), projectLeadLogin, name)
	}
}

func (suite *TeamRouterTestSuite) TestGetMyTutorMatchesUniversityLogin() {
	for _, login := range []string{projectLeadLogin, " PL34EAD "} {
		resp := suite.get(suite.routerAs(tutor(login)), tutorScopePhaseID, "/team/tutors/me")

		require.Equal(suite.T(), http.StatusOK, resp.Code, "login %q", login)
		var me promptTypes.Person
		require.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &me))
		require.Equal(suite.T(), projectLeadID, me.ID.String())
		require.Equal(suite.T(), "Petra", me.FirstName)
	}
}

func (suite *TeamRouterTestSuite) TestGetMyTutorNotFoundForNonTutor() {
	for _, login := range []string{"zz99zzz", ""} {
		resp := suite.get(suite.routerAs(tutor(login)), tutorScopePhaseID, "/team/tutors/me")
		require.Equal(suite.T(), http.StatusNotFound, resp.Code, "login %q", login)
	}
}
