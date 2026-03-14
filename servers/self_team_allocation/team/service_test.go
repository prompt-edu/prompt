package teams

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/team/teamDTO"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/timeframe"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TeamsServiceTestSuite struct {
	suite.Suite
	ctx           context.Context
	testDB        *sdkTestUtils.TestDB[*db.Queries]
	cleanup       func()
	activePhaseID uuid.UUID
	futurePhaseID uuid.UUID
}

func (suite *TeamsServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	require.NoError(suite.T(), err)
	suite.testDB = testDB
	suite.cleanup = cleanup
	suite.activePhaseID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	suite.futurePhaseID = uuid.MustParse("22222222-2222-2222-2222-222222222222")

	TeamsServiceSingleton = &TeamsService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	AssignmentServiceSingleton = &AssignmentService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	timeframe.TimeframeServiceSingleton = timeframe.NewTimeframeService(*testDB.Queries, testDB.Conn)
}

func (suite *TeamsServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *TeamsServiceTestSuite) TestGetAllTeamsReturnsSeedData() {
	teams, err := GetAllTeams(suite.ctx, suite.activePhaseID)

	require.NoError(suite.T(), err)
	suite.GreaterOrEqual(len(teams), 2)

	foundAlpha := false
	for _, team := range teams {
		if team.Name == "Alpha Team" {
			foundAlpha = true
		}
	}
	require.True(suite.T(), foundAlpha)
}

func (suite *TeamsServiceTestSuite) TestGetTeamByID() {
	teamID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	team, err := GetTeamByID(suite.ctx, suite.activePhaseID, teamID)

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), teamID, team.ID)
}

func (suite *TeamsServiceTestSuite) TestCreateNewTeams() {
	coursePhaseID := uuid.New()
	names := []string{"Team-" + uuid.NewString()[0:8], "Team-" + uuid.NewString()[0:8]}

	err := CreateNewTeams(suite.ctx, names, coursePhaseID)

	require.NoError(suite.T(), err)

	dbTeams, err := suite.testDB.Queries.GetTeamsByCoursePhase(suite.ctx, coursePhaseID)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), dbTeams, len(names))
}

func (suite *TeamsServiceTestSuite) TestCreateNewTeamsRejectsEmptyName() {
	err := CreateNewTeams(suite.ctx, []string{"", "Valid"}, uuid.New())

	require.Error(suite.T(), err)
}

func (suite *TeamsServiceTestSuite) TestUpdateTeam() {
	coursePhaseID := uuid.New()
	teamID := suite.createTeam(coursePhaseID, "Original")

	err := UpdateTeam(suite.ctx, coursePhaseID, teamID, "Updated")

	require.NoError(suite.T(), err)

	updated, err := suite.testDB.Queries.GetTeamWithStudentNamesByTeamID(suite.ctx, db.GetTeamWithStudentNamesByTeamIDParams{
		ID:            teamID,
		CoursePhaseID: coursePhaseID,
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "Updated", updated.Name)
}

func (suite *TeamsServiceTestSuite) TestAssignAndLeaveTeam() {
	coursePhaseID := uuid.New()
	teamID := suite.createTeam(coursePhaseID, "Assignables")
	participantID := uuid.New()

	err := AssignTeam(suite.ctx, coursePhaseID, teamID, participantID, "Sam", "Student")
	require.NoError(suite.T(), err)

	row, err := suite.testDB.Queries.GetAssignmentForStudent(suite.ctx, db.GetAssignmentForStudentParams{
		CourseParticipationID: participantID,
		CoursePhaseID:         coursePhaseID,
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), teamID, row.TeamID)

	err = LeaveTeam(suite.ctx, coursePhaseID, teamID, participantID)
	require.NoError(suite.T(), err)

	_, err = suite.testDB.Queries.GetAssignmentForStudent(suite.ctx, db.GetAssignmentForStudentParams{
		CourseParticipationID: participantID,
		CoursePhaseID:         coursePhaseID,
	})
	require.Error(suite.T(), err)
}

func (suite *TeamsServiceTestSuite) TestDeleteTeam() {
	coursePhaseID := uuid.New()
	teamID := suite.createTeam(coursePhaseID, "Disposable")

	err := DeleteTeam(suite.ctx, coursePhaseID, teamID)
	require.NoError(suite.T(), err)

	_, err = suite.testDB.Queries.GetTeamWithStudentNamesByTeamID(suite.ctx, db.GetTeamWithStudentNamesByTeamIDParams{
		ID:            teamID,
		CoursePhaseID: coursePhaseID,
	})
	require.Error(suite.T(), err)
}

func (suite *TeamsServiceTestSuite) TestValidateTimeframeInsideWindow() {
	allowed, err := ValidateTimeframe(suite.ctx, suite.activePhaseID)

	require.NoError(suite.T(), err)
	require.True(suite.T(), allowed)
}

func (suite *TeamsServiceTestSuite) TestValidateTimeframeOutsideWindow() {
	coursePhaseID := uuid.New()
	suite.setTimeframe(coursePhaseID, time.Now().Add(24*time.Hour), time.Now().Add(48*time.Hour))

	allowed, err := ValidateTimeframe(suite.ctx, coursePhaseID)

	require.Error(suite.T(), err)
	require.False(suite.T(), allowed)
}

func (suite *TeamsServiceTestSuite) TestImportTutorsAndGetTutors() {
	coursePhaseID := uuid.New()
	teamID := suite.createTeam(coursePhaseID, "TutorTeam")
	tutors := []teamDTO.Tutor{
		{
			CoursePhaseID:         coursePhaseID,
			CourseParticipationID: uuid.New(),
			FirstName:             "Ina",
			LastName:              "Instructor",
			TeamID:                teamID,
		},
	}

	err := ImportTutors(suite.ctx, coursePhaseID, tutors)
	require.NoError(suite.T(), err)

	result, err := GetTutorsByCoursePhase(suite.ctx, coursePhaseID)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), result, 1)
}

func (suite *TeamsServiceTestSuite) TestCreateAndDeleteManualTutor() {
	coursePhaseID := uuid.New()
	teamID := suite.createTeam(coursePhaseID, "ManualTutorTeam")

	err := CreateManualTutor(suite.ctx, coursePhaseID, "Manual", "Tutor", teamID)
	require.NoError(suite.T(), err)

	tutors, err := GetTutorsByCoursePhase(suite.ctx, coursePhaseID)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), tutors, 1)

	err = DeleteTutor(suite.ctx, coursePhaseID, tutors[0].CourseParticipationID)
	require.NoError(suite.T(), err)
}

func (suite *TeamsServiceTestSuite) createTeam(coursePhaseID uuid.UUID, name string) uuid.UUID {
	teamID := uuid.New()
	err := suite.testDB.Queries.CreateTeam(suite.ctx, db.CreateTeamParams{
		ID:            teamID,
		Name:          name,
		CoursePhaseID: coursePhaseID,
	})
	require.NoError(suite.T(), err)
	return teamID
}

func (suite *TeamsServiceTestSuite) setTimeframe(coursePhaseID uuid.UUID, start, end time.Time) {
	err := suite.testDB.Queries.SetTimeframe(suite.ctx, db.SetTimeframeParams{
		CoursePhaseID: coursePhaseID,
		Starttime:     pgtype.Timestamp{Time: start, Valid: true},
		Endtime:       pgtype.Timestamp{Time: end, Valid: true},
	})
	require.NoError(suite.T(), err)
}

func TestTeamsServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TeamsServiceTestSuite))
}
