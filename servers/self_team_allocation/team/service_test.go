package teams

import (
	"context"
	"fmt"
	"strings"
	"sync"
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
	ctx               context.Context
	testDB            *sdkTestUtils.TestDB[*db.Queries]
	cleanup           func()
	activePhaseID     uuid.UUID
	futurePhaseID     uuid.UUID
	teamsService      *TeamsService
	assignmentService *AssignmentService
}

func (suite *TeamsServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.ctx, "../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../database_dumps/base.sql")
	require.NoError(suite.T(), err)
	suite.testDB = testDB
	suite.cleanup = cleanup
	suite.activePhaseID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	suite.futurePhaseID = uuid.MustParse("22222222-2222-2222-2222-222222222222")

	timeframeService := timeframe.NewTimeframeService(*testDB.Queries)
	suite.teamsService = NewTeamsService(*testDB.Queries, testDB.Conn, timeframeService)
	suite.assignmentService = NewAssignmentService(*testDB.Queries, testDB.Conn)
}

func (suite *TeamsServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *TeamsServiceTestSuite) TestGetAllTeamsReturnsSeedData() {
	teams, err := suite.teamsService.GetAllTeams(suite.ctx, suite.activePhaseID)

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

	team, err := suite.teamsService.GetTeamByID(suite.ctx, suite.activePhaseID, teamID)

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), teamID, team.ID)
}

func (suite *TeamsServiceTestSuite) TestCreateNewTeams() {
	coursePhaseID := uuid.New()
	names := []string{"Team-" + uuid.NewString()[0:8], "Team-" + uuid.NewString()[0:8]}

	err := suite.teamsService.CreateNewTeams(suite.ctx, names, coursePhaseID, uuid.Nil)

	require.NoError(suite.T(), err)

	dbTeams, err := suite.testDB.Queries.GetTeamsByCoursePhase(suite.ctx, coursePhaseID)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), dbTeams, len(names))
}

func (suite *TeamsServiceTestSuite) TestCreateNewTeamsRejectsEmptyName() {
	err := suite.teamsService.CreateNewTeams(suite.ctx, []string{"   ", "Valid"}, uuid.New(), uuid.Nil)

	require.ErrorIs(suite.T(), err, ErrInvalidTeamName)
}

func (suite *TeamsServiceTestSuite) TestCreateNewTeamsRejectsInvalidNames() {
	for _, name := range []string{strings.Repeat("a", MaxTeamNameLength+1), "Line\nBreak"} {
		err := suite.teamsService.CreateNewTeams(suite.ctx, []string{name}, uuid.New(), uuid.Nil)
		require.ErrorIs(suite.T(), err, ErrInvalidTeamName, name)
	}
}

func (suite *TeamsServiceTestSuite) TestCreateNewTeamsTrimsNameAndRecordsCreator() {
	coursePhaseID := uuid.New()
	creatorID := uuid.New()

	err := suite.teamsService.CreateNewTeams(suite.ctx, []string{"  Padded  "}, coursePhaseID, creatorID)
	require.NoError(suite.T(), err)

	dbTeams, err := suite.testDB.Queries.GetTeamsByCoursePhase(suite.ctx, coursePhaseID)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), dbTeams, 1)
	require.Equal(suite.T(), "Padded", dbTeams[0].Name)
	require.Equal(suite.T(), pgtype.UUID{Bytes: creatorID, Valid: true}, dbTeams[0].CreatedByCourseParticipationID)
}

func (suite *TeamsServiceTestSuite) TestCreateNewTeamsRejectsDuplicateName() {
	coursePhaseID := uuid.New()
	suite.createTeam(coursePhaseID, "Taken")

	err := suite.teamsService.CreateNewTeams(suite.ctx, []string{"Taken"}, coursePhaseID, uuid.Nil)
	require.ErrorIs(suite.T(), err, ErrDuplicateTeamName)

	err = suite.teamsService.CreateNewTeams(suite.ctx, []string{"Twice", "Twice"}, coursePhaseID, uuid.Nil)
	require.ErrorIs(suite.T(), err, ErrDuplicateTeamName)
}

// Regression test for GHSA-6j73-qvgm-8p23: a student cannot create more than
// MaxTeamsCreatedPerStudent teams, even with concurrent requests.
func (suite *TeamsServiceTestSuite) TestCreateNewTeamsLimitsTeamsPerStudent() {
	coursePhaseID := uuid.New()
	creatorID := uuid.New()

	const attempts = 10
	var wg sync.WaitGroup
	errs := make(chan error, attempts)
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- suite.teamsService.CreateNewTeams(suite.ctx, []string{fmt.Sprintf("Spam %d", i)}, coursePhaseID, creatorID)
		}()
	}
	wg.Wait()
	close(errs)

	succeeded := 0
	for err := range errs {
		if err == nil {
			succeeded++
			continue
		}
		require.ErrorIs(suite.T(), err, ErrTeamLimitReached)
	}
	require.Equal(suite.T(), MaxTeamsCreatedPerStudent, succeeded)

	dbTeams, err := suite.testDB.Queries.GetTeamsByCoursePhase(suite.ctx, coursePhaseID)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), dbTeams, MaxTeamsCreatedPerStudent)

	// Lecturers are not limited.
	err = suite.teamsService.CreateNewTeams(suite.ctx, []string{"Lecturer 1", "Lecturer 2", "Lecturer 3"}, coursePhaseID, uuid.Nil)
	require.NoError(suite.T(), err)
}

func (suite *TeamsServiceTestSuite) TestCreateNewTeamsLimitsTeamsPerPhaseForStudents() {
	coursePhaseID := uuid.New()
	for i := range MaxTeamsPerPhase {
		suite.createTeam(coursePhaseID, fmt.Sprintf("Existing %d", i))
	}

	err := suite.teamsService.CreateNewTeams(suite.ctx, []string{"One too many"}, coursePhaseID, uuid.New())
	require.ErrorIs(suite.T(), err, ErrTeamLimitReached)
}

func (suite *TeamsServiceTestSuite) TestUpdateTeamRejectsInvalidName() {
	coursePhaseID := uuid.New()
	teamID := suite.createTeam(coursePhaseID, "Before")

	err := suite.teamsService.UpdateTeam(suite.ctx, coursePhaseID, teamID, " ")
	require.ErrorIs(suite.T(), err, ErrInvalidTeamName)
}

func (suite *TeamsServiceTestSuite) TestUpdateTeam() {
	coursePhaseID := uuid.New()
	teamID := suite.createTeam(coursePhaseID, "Original")

	err := suite.teamsService.UpdateTeam(suite.ctx, coursePhaseID, teamID, "Updated")

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

	err := suite.assignmentService.AssignTeam(suite.ctx, coursePhaseID, teamID, participantID, "Sam", "Student")
	require.NoError(suite.T(), err)

	row, err := suite.testDB.Queries.GetAssignmentForStudent(suite.ctx, db.GetAssignmentForStudentParams{
		CourseParticipationID: participantID,
		CoursePhaseID:         coursePhaseID,
	})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), teamID, row.TeamID)

	err = suite.assignmentService.LeaveTeam(suite.ctx, coursePhaseID, teamID, participantID)
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

	err := suite.teamsService.DeleteTeam(suite.ctx, coursePhaseID, teamID)
	require.NoError(suite.T(), err)

	_, err = suite.testDB.Queries.GetTeamWithStudentNamesByTeamID(suite.ctx, db.GetTeamWithStudentNamesByTeamIDParams{
		ID:            teamID,
		CoursePhaseID: coursePhaseID,
	})
	require.Error(suite.T(), err)
}

func (suite *TeamsServiceTestSuite) TestValidateTimeframeInsideWindow() {
	allowed, err := suite.teamsService.ValidateTimeframe(suite.ctx, suite.activePhaseID)

	require.NoError(suite.T(), err)
	require.True(suite.T(), allowed)
}

func (suite *TeamsServiceTestSuite) TestValidateTimeframeOutsideWindow() {
	coursePhaseID := uuid.New()
	suite.setTimeframe(coursePhaseID, time.Now().Add(24*time.Hour), time.Now().Add(48*time.Hour))

	allowed, err := suite.teamsService.ValidateTimeframe(suite.ctx, coursePhaseID)

	require.ErrorIs(suite.T(), err, ErrOutsideTimeframe)
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

	err := suite.teamsService.ImportTutors(suite.ctx, coursePhaseID, tutors)
	require.NoError(suite.T(), err)

	result, err := suite.teamsService.GetTutorsByCoursePhase(suite.ctx, coursePhaseID)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), result, 1)
}

func (suite *TeamsServiceTestSuite) TestCreateAndDeleteManualTutor() {
	coursePhaseID := uuid.New()
	teamID := suite.createTeam(coursePhaseID, "ManualTutorTeam")

	err := suite.teamsService.CreateManualTutor(suite.ctx, coursePhaseID, "Manual", "Tutor", teamID)
	require.NoError(suite.T(), err)

	tutors, err := suite.teamsService.GetTutorsByCoursePhase(suite.ctx, coursePhaseID)
	require.NoError(suite.T(), err)
	require.Len(suite.T(), tutors, 1)

	err = suite.teamsService.DeleteTutor(suite.ctx, coursePhaseID, tutors[0].CourseParticipationID)
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
		Starttime:     pgtype.Timestamptz{Time: start, Valid: true},
		Endtime:       pgtype.Timestamptz{Time: end, Valid: true},
	})
	require.NoError(suite.T(), err)
}

func TestTeamsServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TeamsServiceTestSuite))
}
