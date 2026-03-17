package skills

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

type SkillServiceTestSuite struct {
	suite.Suite
	suiteCtx     context.Context
	cleanup      func()
	skillService SkillsService
}

func (suite *SkillServiceTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/skills.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.skillService = SkillsService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	SkillsServiceSingleton = &suite.skillService
}

func (suite *SkillServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *SkillServiceTestSuite) TestGetAllSkills() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")

	skills, err := GetAllSkills(suite.suiteCtx, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(skills), 0, "Expected at least one skill")

	for _, skill := range skills {
		assert.NotEmpty(suite.T(), skill.ID, "Skill ID should not be empty")
		assert.NotEmpty(suite.T(), skill.Name, "Skill name should not be empty")
	}
}

func (suite *SkillServiceTestSuite) TestGetAllSkillsNonExistentCoursePhase() {
	nonExistentID := uuid.New()

	skills, err := GetAllSkills(suite.suiteCtx, nonExistentID)
	assert.NoError(suite.T(), err, "Should not error for non-existent course phase")
	assert.Equal(suite.T(), 0, len(skills), "Should return empty list for non-existent course phase")
}

func (suite *SkillServiceTestSuite) TestCreateNewSkills() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	newSkillNames := []string{"React", "Vue.js", "Angular"}

	// Get initial count
	initialSkills, err := GetAllSkills(suite.suiteCtx, coursePhaseID)
	assert.NoError(suite.T(), err)
	initialCount := len(initialSkills)

	// Create new skills
	err = CreateNewSkills(suite.suiteCtx, newSkillNames, coursePhaseID)
	assert.NoError(suite.T(), err)

	// Verify skills were created
	updatedSkills, err := GetAllSkills(suite.suiteCtx, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), initialCount+len(newSkillNames), len(updatedSkills), "Should have created new skills")

	// Verify the new skills exist
	skillNames := make(map[string]bool)
	for _, skill := range updatedSkills {
		skillNames[skill.Name] = true
	}

	for _, skillName := range newSkillNames {
		assert.True(suite.T(), skillNames[skillName], "Skill %s should exist", skillName)
	}
}

func (suite *SkillServiceTestSuite) TestCreateNewSkillsEmptyList() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	emptySkillNames := []string{}

	err := CreateNewSkills(suite.suiteCtx, emptySkillNames, coursePhaseID)
	assert.NoError(suite.T(), err, "Should not error for empty skill list")
}

func (suite *SkillServiceTestSuite) TestUpdateSkillName() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	skillID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	newSkillName := "Advanced Java Programming"

	err := UpdateSkill(suite.suiteCtx, coursePhaseID, skillID, newSkillName)
	assert.NoError(suite.T(), err)

	// Verify the skill was updated
	skills, err := GetAllSkills(suite.suiteCtx, coursePhaseID)
	assert.NoError(suite.T(), err)

	found := false
	for _, skill := range skills {
		if skill.ID == skillID {
			assert.Equal(suite.T(), newSkillName, skill.Name, "Skill name should be updated")
			found = true
			break
		}
	}
	assert.True(suite.T(), found, "Updated skill should be found")
}

func (suite *SkillServiceTestSuite) TestUpdateSkillNameNonExistent() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	nonExistentID := uuid.New()
	newSkillName := "Non-existent Skill"

	err := UpdateSkill(suite.suiteCtx, coursePhaseID, nonExistentID, newSkillName)
	assert.NoError(suite.T(), err, "Should not error for non-existent skill (UPDATE does not fail)")
}

func (suite *SkillServiceTestSuite) TestDeleteSkill() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	skillID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	// Get initial count
	initialSkills, err := GetAllSkills(suite.suiteCtx, coursePhaseID)
	assert.NoError(suite.T(), err)
	initialCount := len(initialSkills)

	// Delete the skill
	err = DeleteSkill(suite.suiteCtx, coursePhaseID, skillID)
	assert.NoError(suite.T(), err)

	// Verify skill was deleted
	updatedSkills, err := GetAllSkills(suite.suiteCtx, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), initialCount-1, len(updatedSkills), "Should have deleted one skill")

	// Verify the specific skill is gone
	for _, skill := range updatedSkills {
		assert.NotEqual(suite.T(), skillID, skill.ID, "Deleted skill should not be found")
	}
}

func (suite *SkillServiceTestSuite) TestDeleteSkillNonExistent() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	nonExistentID := uuid.New()

	err := DeleteSkill(suite.suiteCtx, coursePhaseID, nonExistentID)
	assert.NoError(suite.T(), err, "Should not error for non-existent skill (DELETE does not fail)")
}

func TestSkillServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SkillServiceTestSuite))
}
