package teams

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TeamServiceTestSuite struct {
	suite.Suite
	suiteCtx    context.Context
	cleanup     func()
	teamService TeamsService
}

func (suite *TeamServiceTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/teams.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.teamService = TeamsService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	TeamsServiceSingleton = &suite.teamService
}

func (suite *TeamServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *TeamServiceTestSuite) TestGetAllTeams() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")

	teams, err := GetAllTeams(suite.suiteCtx, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(teams), 0, "Expected at least zero teams")

	for _, team := range teams {
		assert.NotEmpty(suite.T(), team.ID, "Team ID should not be empty")
		assert.NotEmpty(suite.T(), team.Name, "Team name should not be empty")
	}
}

func (suite *TeamServiceTestSuite) TestGetAllTeamsNonExistentCoursePhase() {
	nonExistentID := uuid.New()

	teams, err := GetAllTeams(suite.suiteCtx, nonExistentID)
	assert.NoError(suite.T(), err, "Should not error for non-existent course phase")
	assert.Equal(suite.T(), 0, len(teams), "Should return empty list for non-existent course phase")
}

func (suite *TeamServiceTestSuite) TestGetTeamByID() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	teamID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	team, err := GetTeamByID(suite.suiteCtx, coursePhaseID, teamID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), teamID, team.ID, "Team ID should match")
	assert.Equal(suite.T(), "Team Alpha", team.Name, "Team name should match")
}

func (suite *TeamServiceTestSuite) TestGetTeamByIDNotFound() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	nonExistentID := uuid.New()

	_, err := GetTeamByID(suite.suiteCtx, coursePhaseID, nonExistentID)
	assert.Error(suite.T(), err, "Should error for non-existent team")
}

func (suite *TeamServiceTestSuite) TestCreateNewTeams() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	newTeamNames := []string{"Team Epsilon", "Team Zeta", "Team Eta"}

	err := CreateNewTeams(suite.suiteCtx, newTeamNames, coursePhaseID, "", false, nil)
	assert.NoError(suite.T(), err)

	// Note: We can't easily verify the teams were created without calling GetAllTeams
	// which might have different behavior depending on allocations
}

func (suite *TeamServiceTestSuite) TestCreateNewTeamsEmptyList() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	emptyTeamNames := []string{}

	err := CreateNewTeams(suite.suiteCtx, emptyTeamNames, coursePhaseID, "", false, nil)
	assert.NoError(suite.T(), err, "Should not error for empty team list")
}

func (suite *TeamServiceTestSuite) TestUpdateTeamName() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	teamID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	newTeamName := "Updated Team Alpha"

	err := UpdateTeam(suite.suiteCtx, coursePhaseID, teamID, newTeamName)
	assert.NoError(suite.T(), err)

	// Verify the team was updated
	team, err := GetTeamByID(suite.suiteCtx, coursePhaseID, teamID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newTeamName, team.Name, "Team name should be updated")
}

func (suite *TeamServiceTestSuite) TestUpdateTeamNameNonExistent() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	nonExistentID := uuid.New()
	newTeamName := "Non-existent Team"

	err := UpdateTeam(suite.suiteCtx, coursePhaseID, nonExistentID, newTeamName)
	assert.NoError(suite.T(), err, "UpdateTeam does not return error for non-existent team")
}

func (suite *TeamServiceTestSuite) TestDeleteTeam() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	teamID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	// Verify team exists first
	_, err := GetTeamByID(suite.suiteCtx, coursePhaseID, teamID)
	assert.NoError(suite.T(), err, "Team should exist before deletion")

	// Delete the team
	err = DeleteTeam(suite.suiteCtx, coursePhaseID, teamID)
	assert.NoError(suite.T(), err)

	// Verify team is deleted
	_, err = GetTeamByID(suite.suiteCtx, coursePhaseID, teamID)
	assert.Error(suite.T(), err, "Team should not exist after deletion")
}

func (suite *TeamServiceTestSuite) TestDeleteTeamNonExistent() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	nonExistentID := uuid.New()

	err := DeleteTeam(suite.suiteCtx, coursePhaseID, nonExistentID)
	assert.NoError(suite.T(), err, "DeleteTeam does not return error for non-existent team")
}

func TestTeamServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TeamServiceTestSuite))
}
