package competencyMap

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/competencies/competencyMap/competencyMapDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type CompetencyMapServiceTestSuite struct {
	suite.Suite
	suiteCtx      context.Context
	cleanup       func()
	service       CompetencyMapService
	coursePhaseID uuid.UUID
}

func (suite *CompetencyMapServiceTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../../database_dumps/competencyMaps.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}

	suite.cleanup = cleanup
	suite.service = CompetencyMapService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	CompetencyMapServiceSingleton = &suite.service
	suite.coursePhaseID = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
}

func (suite *CompetencyMapServiceTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *CompetencyMapServiceTestSuite) TestCreateCompetencyMapping() {
	// Use existing competency IDs from the database dump
	fromCompetencyID := uuid.MustParse("c1234567-1234-1234-1234-123456789012") // Programming Skills
	toCompetencyID := uuid.MustParse("c4234567-1234-1234-1234-123456789012")   // Communication

	req := competencyMapDTO.CompetencyMapping{
		FromCompetencyID: fromCompetencyID,
		ToCompetencyID:   toCompetencyID,
	}

	err := CreateCompetencyMapping(suite.suiteCtx, suite.coursePhaseID, req)
	assert.NoError(suite.T(), err)

	// Verify the mapping was created
	mappings, err := GetCompetencyMappings(suite.suiteCtx, suite.coursePhaseID, fromCompetencyID)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(mappings), 1) // Should have at least the one we just created

	// Find our created mapping
	found := false
	for _, mapping := range mappings {
		if mapping.FromCompetencyID == fromCompetencyID && mapping.ToCompetencyID == toCompetencyID {
			found = true
			break
		}
	}
	assert.True(suite.T(), found, "Created mapping should be found")
}

func (suite *CompetencyMapServiceTestSuite) TestDeleteCompetencyMapping() {
	// Use existing competency IDs from the database dump
	fromCompetencyID := uuid.MustParse("c2234567-1234-1234-1234-123456789012") // System Design
	toCompetencyID := uuid.MustParse("c7234567-1234-1234-1234-123456789012")   // Mentoring

	// First create a mapping
	createReq := competencyMapDTO.CompetencyMapping{
		FromCompetencyID: fromCompetencyID,
		ToCompetencyID:   toCompetencyID,
	}
	err := CreateCompetencyMapping(suite.suiteCtx, suite.coursePhaseID, createReq)
	assert.NoError(suite.T(), err)

	// Then delete it
	deleteReq := competencyMapDTO.CompetencyMapping{
		FromCompetencyID: fromCompetencyID,
		ToCompetencyID:   toCompetencyID,
	}
	err = DeleteCompetencyMapping(suite.suiteCtx, suite.coursePhaseID, deleteReq)
	assert.NoError(suite.T(), err)

	// Verify the mapping was deleted - check that our specific mapping is gone
	mappings, err := GetCompetencyMappings(suite.suiteCtx, suite.coursePhaseID, fromCompetencyID)
	assert.NoError(suite.T(), err)

	// Ensure our specific mapping is not found
	found := false
	for _, mapping := range mappings {
		if mapping.FromCompetencyID == fromCompetencyID && mapping.ToCompetencyID == toCompetencyID {
			found = true
			break
		}
	}
	assert.False(suite.T(), found, "Deleted mapping should not be found")
}

func (suite *CompetencyMapServiceTestSuite) TestGetAllCompetencyMappings() {
	// Create a few mappings using existing competency IDs
	fromCompetencyID1 := uuid.MustParse("c3234567-1234-1234-1234-123456789012") // Testing
	toCompetencyID1 := uuid.MustParse("c5234567-1234-1234-1234-123456789012")   // Teamwork
	fromCompetencyID2 := uuid.MustParse("c4234567-1234-1234-1234-123456789012") // Communication
	toCompetencyID2 := uuid.MustParse("c6234567-1234-1234-1234-123456789012")   // Project Management

	req1 := competencyMapDTO.CompetencyMapping{
		FromCompetencyID: fromCompetencyID1,
		ToCompetencyID:   toCompetencyID1,
	}
	req2 := competencyMapDTO.CompetencyMapping{
		FromCompetencyID: fromCompetencyID2,
		ToCompetencyID:   toCompetencyID2,
	}

	err := CreateCompetencyMapping(suite.suiteCtx, suite.coursePhaseID, req1)
	assert.NoError(suite.T(), err)
	err = CreateCompetencyMapping(suite.suiteCtx, suite.coursePhaseID, req2)
	assert.NoError(suite.T(), err)

	// Get all mappings
	mappings, err := GetAllCompetencyMappings(suite.suiteCtx, suite.coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(mappings), 8) // 6 from dump + 2 we just created
}

func TestCompetencyMapServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CompetencyMapServiceTestSuite))
}
