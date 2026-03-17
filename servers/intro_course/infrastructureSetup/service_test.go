package infrastructureSetup

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type InfrastructureServiceTestSuite struct {
	suite.Suite
	ctx     context.Context
	cleanup func()
	service InfrastructureService
}

func (suite *InfrastructureServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/intro_course.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup

	suite.service = InfrastructureService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	InfrastructureServiceSingleton = &suite.service
}

func (suite *InfrastructureServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func TestInfrastructureServiceTestSuite(t *testing.T) {
	suite.Run(t, new(InfrastructureServiceTestSuite))
}

func (suite *InfrastructureServiceTestSuite) TestGetAllStudentGitlabStatus() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")

	statuses, err := GetAllStudentGitlabStatus(suite.ctx, coursePhaseID)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), statuses, 2)
}

func (suite *InfrastructureServiceTestSuite) TestManuallyOverwriteStudentGitlabStatus() {
	coursePhaseID := uuid.MustParse("5179d58a-d00d-4fa7-94a5-397bc69fab03")
	courseParticipationID := uuid.MustParse("55555555-5555-5555-5555-555555555555")

	err := ManuallyOverwriteStudentGitlabStatus(suite.ctx, coursePhaseID, courseParticipationID)
	assert.NoError(suite.T(), err)

	statuses, err := GetAllStudentGitlabStatus(suite.ctx, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), statuses, 1)
	assert.True(suite.T(), statuses[0].GitlabSuccess)
}
