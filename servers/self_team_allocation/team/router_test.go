package teams

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/team/teamDTO"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/timeframe"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TeamRouterTestSuite struct {
	suite.Suite
	ctx               context.Context
	testDB            *sdkTestUtils.TestDB[*db.Queries]
	cleanup           func()
	teamsService      *TeamsService
	assignmentService *AssignmentService
}

func (suite *TeamRouterTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()

	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.ctx, "../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../database_dumps/base.sql")
	require.NoError(suite.T(), err)
	suite.testDB = testDB
	suite.cleanup = cleanup

	timeframeService := timeframe.NewTimeframeService(*testDB.Queries)
	suite.teamsService = NewTeamsService(*testDB.Queries, testDB.Conn, timeframeService)
	suite.assignmentService = NewAssignmentService(*testDB.Queries, testDB.Conn)
}

func (suite *TeamRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *TeamRouterTestSuite) newRouter(courseParticipationID uuid.UUID) *gin.Engine {
	router := gin.Default()
	api := router.Group("/api/course_phase/:coursePhaseID")
	RegisterRoutes(api, suite.teamsService, suite.assignmentService, func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithParticipation(allowedRoles, courseParticipationID)
	})
	return router
}

// newStudentRouter authenticates every request as a plain student of the course, without
// the admin roles the shared mock grants.
func (suite *TeamRouterTestSuite) newStudentRouter(courseParticipationID uuid.UUID) *gin.Engine {
	router := gin.Default()
	api := router.Group("/api/course_phase/:coursePhaseID")
	RegisterRoutes(api, suite.teamsService, suite.assignmentService, func(allowedRoles ...string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("courseParticipationID", courseParticipationID)
			c.Set("firstName", "Student")
			c.Set("lastName", "User")
			keycloakTokenVerifier.SetTokenUser(c, keycloakTokenVerifier.TokenUser{
				Roles:                 map[string]bool{promptSDK.CourseStudent: true},
				IsStudentOfCourse:     true,
				CourseParticipationID: courseParticipationID,
			})
			c.Next()
		}
	})
	return router
}

func (suite *TeamRouterTestSuite) postTeams(router *gin.Engine, coursePhaseID uuid.UUID, names ...string) *httptest.ResponseRecorder {
	payload, _ := json.Marshal(teamDTO.CreateTeamsRequest{TeamNames: names})
	req, _ := http.NewRequest("POST", "/api/course_phase/"+coursePhaseID.String()+"/team", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

// Regression test for GHSA-6j73-qvgm-8p23: outside the window the request is rejected with
// 400 instead of a 500 with a nil error.
func (suite *TeamRouterTestSuite) TestCreateTeamsRouteOutsideTimeframe() {
	coursePhaseID := uuid.New()
	err := suite.testDB.Queries.SetTimeframe(suite.ctx, db.SetTimeframeParams{
		CoursePhaseID: coursePhaseID,
		Starttime:     pgtype.Timestamptz{Time: time.Now().Add(-2 * time.Hour), Valid: true},
		Endtime:       pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true},
	})
	require.NoError(suite.T(), err)

	resp := suite.postTeams(suite.newStudentRouter(uuid.New()), coursePhaseID, "Late Team")

	require.Equal(suite.T(), http.StatusBadRequest, resp.Code)
	require.Contains(suite.T(), resp.Body.String(), ErrOutsideTimeframe.Error())
}

// Regression test for GHSA-6j73-qvgm-8p23: a student hits the per-student limit with 409,
// invalid names are rejected with 400, and the creator is recorded.
func (suite *TeamRouterTestSuite) TestCreateTeamsRouteStudentLimits() {
	coursePhaseID := suite.ensureTimeframePhase()
	studentID := uuid.New()
	router := suite.newStudentRouter(studentID)

	require.Equal(suite.T(), http.StatusBadRequest, suite.postTeams(router, coursePhaseID, strings.Repeat("x", MaxTeamNameLength+1)).Code)

	for i := range MaxTeamsCreatedPerStudent {
		require.Equal(suite.T(), http.StatusCreated, suite.postTeams(router, coursePhaseID, fmt.Sprintf("Student Team %d", i)).Code)
	}
	require.Equal(suite.T(), http.StatusConflict, suite.postTeams(router, coursePhaseID, "Student Team extra").Code)

	created, err := suite.testDB.Queries.CountTeamsCreatedByParticipation(suite.ctx, db.CountTeamsCreatedByParticipationParams{
		CoursePhaseID:                  coursePhaseID,
		CreatedByCourseParticipationID: pgtype.UUID{Bytes: studentID, Valid: true},
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), int64(MaxTeamsCreatedPerStudent), created)
}

func (suite *TeamRouterTestSuite) TestGetAllTeamsRoute() {
	router := suite.newRouter(uuid.UUID{})
	req, _ := http.NewRequest("GET", "/api/course_phase/11111111-1111-1111-1111-111111111111/team", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)

	var teams []promptTypes.Team
	err := json.Unmarshal(resp.Body.Bytes(), &teams)
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), teams)
}

func (suite *TeamRouterTestSuite) TestGetTeamByIDRoute() {
	router := suite.newRouter(uuid.UUID{})
	url := "/api/course_phase/11111111-1111-1111-1111-111111111111/team/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *TeamRouterTestSuite) TestCreateTeamsRoute() {
	router := suite.newRouter(uuid.UUID{})
	body := teamDTO.CreateTeamsRequest{TeamNames: []string{"RouterTeam"}}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/course_phase/11111111-1111-1111-1111-111111111111/team", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusCreated, resp.Code)

	dbTeams, err := suite.testDB.Queries.GetTeamsByCoursePhase(suite.ctx, uuid.MustParse("11111111-1111-1111-1111-111111111111"))
	require.NoError(suite.T(), err)
	found := false
	for _, t := range dbTeams {
		if t.Name == "RouterTeam" {
			found = true
			break
		}
	}
	require.True(suite.T(), found)
}

func (suite *TeamRouterTestSuite) TestUpdateTeamRoute() {
	coursePhaseID := uuid.New()
	teamID := suite.createTeam(coursePhaseID, "RouterOriginal")

	router := suite.newRouter(uuid.UUID{})
	body := teamDTO.UpdateTeamRequest{NewTeamName: "RouterUpdated"}
	payload, _ := json.Marshal(body)

	url := "/api/course_phase/" + coursePhaseID.String() + "/team/" + teamID.String()
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)

	updated, err := suite.testDB.Queries.GetTeamWithStudentNamesByTeamID(suite.ctx, db.GetTeamWithStudentNamesByTeamIDParams{
		ID:            teamID,
		CoursePhaseID: coursePhaseID,
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "RouterUpdated", updated.Name)
}

func (suite *TeamRouterTestSuite) TestAssignTeamRoute() {
	coursePhaseID := suite.ensureTimeframePhase()
	teamID := suite.createTeam(coursePhaseID, "RouterAssignable")
	participationID := uuid.New()
	router := suite.newRouter(participationID)

	url := "/api/course_phase/" + coursePhaseID.String() + "/team/" + teamID.String() + "/assignment"
	req, _ := http.NewRequest("PUT", url, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)

	assignment, err := suite.testDB.Queries.GetAssignmentForStudent(suite.ctx, db.GetAssignmentForStudentParams{
		CourseParticipationID: participationID,
		CoursePhaseID:         coursePhaseID,
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), teamID, assignment.TeamID)
}

func (suite *TeamRouterTestSuite) TestAssignTeamRouteTeamFull() {
	coursePhaseID := suite.ensureTimeframePhase()
	teamID := suite.createTeam(coursePhaseID, "RouterFull")
	for i := 0; i < 3; i++ {
		err := suite.testDB.Queries.CreateOrUpdateAssignment(suite.ctx, db.CreateOrUpdateAssignmentParams{
			ID:                    uuid.New(),
			CourseParticipationID: uuid.New(),
			StudentFirstName:      "Member",
			StudentLastName:       uuid.NewString()[0:4],
			TeamID:                teamID,
			CoursePhaseID:         coursePhaseID,
		})
		require.NoError(suite.T(), err)
	}

	router := suite.newRouter(uuid.New())
	url := "/api/course_phase/" + coursePhaseID.String() + "/team/" + teamID.String() + "/assignment"
	req, _ := http.NewRequest("PUT", url, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *TeamRouterTestSuite) TestLeaveTeamRoute() {
	coursePhaseID := suite.ensureTimeframePhase()
	teamID := suite.createTeam(coursePhaseID, "RouterLeave")
	participantID := uuid.New()
	err := suite.testDB.Queries.CreateOrUpdateAssignment(suite.ctx, db.CreateOrUpdateAssignmentParams{
		ID:                    uuid.New(),
		CourseParticipationID: participantID,
		StudentFirstName:      "Leave",
		StudentLastName:       "Tester",
		TeamID:                teamID,
		CoursePhaseID:         coursePhaseID,
	})
	require.NoError(suite.T(), err)

	router := suite.newRouter(participantID)
	url := "/api/course_phase/" + coursePhaseID.String() + "/team/" + teamID.String() + "/assignment"
	req, _ := http.NewRequest("DELETE", url, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)

	_, err = suite.testDB.Queries.GetAssignmentForStudent(suite.ctx, db.GetAssignmentForStudentParams{
		CourseParticipationID: participantID,
		CoursePhaseID:         coursePhaseID,
	})
	require.Error(suite.T(), err)
}

func (suite *TeamRouterTestSuite) TestDeleteTeamRoute() {
	coursePhaseID := uuid.New()
	teamID := suite.createTeam(coursePhaseID, "RouterDelete")

	router := suite.newRouter(uuid.UUID{})
	url := "/api/course_phase/" + coursePhaseID.String() + "/team/" + teamID.String()
	req, _ := http.NewRequest("DELETE", url, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)

	_, err := suite.testDB.Queries.GetTeamWithStudentNamesByTeamID(suite.ctx, db.GetTeamWithStudentNamesByTeamIDParams{
		ID:            teamID,
		CoursePhaseID: coursePhaseID,
	})
	require.Error(suite.T(), err)
}

func (suite *TeamRouterTestSuite) TestImportTutorsRoute() {
	coursePhaseID := uuid.New()
	teamID := suite.createTeam(coursePhaseID, "RouterTutorImport")
	router := suite.newRouter(uuid.UUID{})

	tutors := []teamDTO.Tutor{
		{
			CoursePhaseID:         coursePhaseID,
			CourseParticipationID: uuid.New(),
			FirstName:             "Import",
			LastName:              "Tutor",
			TeamID:                teamID,
		},
	}
	payload, _ := json.Marshal(tutors)

	url := "/api/course_phase/" + coursePhaseID.String() + "/team/tutors"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusCreated, resp.Code)

	tutorList, err := suite.testDB.Queries.GetTutorsByCoursePhase(suite.ctx, coursePhaseID)
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), tutorList)
}

func (suite *TeamRouterTestSuite) TestCreateManualTutorRoute() {
	coursePhaseID := uuid.New()
	teamID := suite.createTeam(coursePhaseID, "RouterManualTutor")
	router := suite.newRouter(uuid.UUID{})

	body := map[string]string{"firstName": "Manual", "lastName": "Tutor"}
	payload, _ := json.Marshal(body)

	url := "/api/course_phase/" + coursePhaseID.String() + "/team/" + teamID.String() + "/tutor"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusCreated, resp.Code)

	tutors, err := suite.testDB.Queries.GetTutorsByCoursePhase(suite.ctx, coursePhaseID)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), tutors, 1)
}

func (suite *TeamRouterTestSuite) TestDeleteTutorRoute() {
	coursePhaseID := uuid.New()
	teamID := suite.createTeam(coursePhaseID, "RouterDeleteTutor")
	tutorID := uuid.New()
	err := suite.testDB.Queries.CreateTutor(suite.ctx, db.CreateTutorParams{
		CoursePhaseID:         coursePhaseID,
		CourseParticipationID: tutorID,
		FirstName:             "Delete",
		LastName:              "Tutor",
		TeamID:                teamID,
	})
	require.NoError(suite.T(), err)

	router := suite.newRouter(uuid.UUID{})
	url := "/api/course_phase/" + coursePhaseID.String() + "/team/tutor/" + tutorID.String()
	req, _ := http.NewRequest("DELETE", url, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)

	tutors, err := suite.testDB.Queries.GetTutorsByCoursePhase(suite.ctx, coursePhaseID)
	require.NoError(suite.T(), err)
	require.Empty(suite.T(), tutors)
}

func (suite *TeamRouterTestSuite) TestGetTutorsRoute() {
	router := suite.newRouter(uuid.UUID{})
	req, _ := http.NewRequest("GET", "/api/course_phase/11111111-1111-1111-1111-111111111111/team/tutors", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *TeamRouterTestSuite) createTeam(coursePhaseID uuid.UUID, name string) uuid.UUID {
	teamID := uuid.New()
	err := suite.testDB.Queries.CreateTeam(suite.ctx, db.CreateTeamParams{
		ID:            teamID,
		Name:          name,
		CoursePhaseID: coursePhaseID,
	})
	require.NoError(suite.T(), err)
	return teamID
}

func (suite *TeamRouterTestSuite) ensureTimeframePhase() uuid.UUID {
	coursePhaseID := uuid.New()
	err := suite.testDB.Queries.SetTimeframe(suite.ctx, db.SetTimeframeParams{
		CoursePhaseID: coursePhaseID,
		Starttime:     pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true},
		Endtime:       pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
	})
	require.NoError(suite.T(), err)
	return coursePhaseID
}

func TestTeamRouterTestSuite(t *testing.T) {
	suite.Run(t, new(TeamRouterTestSuite))
}
