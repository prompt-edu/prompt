package competencies

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/competencies/competencyDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CompetencyServiceTestSuite struct {
	suite.Suite
	router            *gin.Engine
	ctx               context.Context
	cleanup           func()
	mockCoreCleanup   func()
	competencyService CompetencyService
}

func (suite *CompetencyServiceTestSuite) SetupTest() {
	suite.ctx = context.Background()

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/categories.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}

	suite.cleanup = cleanup

	// Set up mock core service
	_, mockCleanup := testutils.SetupMockCoreService()
	suite.mockCoreCleanup = mockCleanup

	suite.competencyService = CompetencyService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}

	CompetencyServiceSingleton = &suite.competencyService

	// Initialize other service modules needed for schema copy logic
	router := gin.New()
	group := router.Group("")
	assessmentSchemas.InitAssessmentSchemaModule(group, *testDB.Queries, testDB.Conn)
	coursePhaseConfig.InitCoursePhaseConfigModule(group, *testDB.Queries, testDB.Conn)

	suite.router = gin.Default()
}

func (suite *CompetencyServiceTestSuite) TearDownTest() {
	if suite.mockCoreCleanup != nil {
		suite.mockCoreCleanup()
	}
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *CompetencyServiceTestSuite) TestListCompetencies() {
	competencies, err := ListCompetencies(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(competencies), 0, "Expected at least one competency")

	for _, competency := range competencies {
		assert.NotEmpty(suite.T(), competency.ID, "Competency ID should not be empty")
		assert.NotEmpty(suite.T(), competency.Name, "Competency Name should not be empty")
		assert.NotEmpty(suite.T(), competency.CategoryID, "Competency CategoryID should not be empty")
		assert.Greater(suite.T(), competency.Weight, int32(0), "Competency Weight should be positive")
	}
}

func (suite *CompetencyServiceTestSuite) TestGetCompetency() {
	// Test with a known competency ID from the test data
	competencyID := uuid.MustParse("20725c05-bfd7-45a7-a981-d092e14f98d3")

	competency, err := GetCompetency(suite.ctx, competencyID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), competencyID, competency.ID, "Competency ID should match")
	assert.Equal(suite.T(), "GitLab Project Management", competency.Name, "Competency name should match")
	assert.NotEmpty(suite.T(), competency.DescriptionVeryBad, "Very Bad description should not be empty")
	assert.NotEmpty(suite.T(), competency.DescriptionBad, "Bad description should not be empty")
	assert.NotEmpty(suite.T(), competency.DescriptionOk, "Ok description should not be empty")
	assert.NotEmpty(suite.T(), competency.DescriptionGood, "Good description should not be empty")
	assert.NotEmpty(suite.T(), competency.DescriptionVeryGood, "Very Good description should not be empty")
}

func (suite *CompetencyServiceTestSuite) TestGetCompetencyNotFound() {
	// Test with a non-existent competency ID
	nonExistentID := uuid.New()

	_, err := GetCompetency(suite.ctx, nonExistentID)
	assert.Error(suite.T(), err, "Expected error for non-existent competency")
}

func (suite *CompetencyServiceTestSuite) TestListCompetenciesByCategory() {
	// Test with a known category ID from the test data
	categoryID := uuid.MustParse("25f1c984-ba31-4cf2-aa8e-5662721bf44e") // Version Control category

	competencies, err := ListCompetenciesByCategory(suite.ctx, categoryID)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(competencies), 0, "Expected at least one competency for Version Control category")

	for _, competency := range competencies {
		assert.Equal(suite.T(), categoryID, competency.CategoryID, "All competencies should belong to the specified category")
		assert.NotEmpty(suite.T(), competency.Name, "Competency name should not be empty")
	}
}

func (suite *CompetencyServiceTestSuite) TestListCompetenciesByCategoryEmpty() {
	// Test with a non-existent category ID
	nonExistentCategoryID := uuid.New()

	competencies, err := ListCompetenciesByCategory(suite.ctx, nonExistentCategoryID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, len(competencies), "Expected no competencies for non-existent category")
}

func (suite *CompetencyServiceTestSuite) TestCreateCompetency() {
	// Use a known category ID from test data
	categoryID := uuid.MustParse("815b159b-cab3-49b4-8060-c4722d59241d")    // User Interface category
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02") // Dev Application phase from test data

	newCompetency := competencyDTO.CreateCompetencyRequest{
		CategoryID:          categoryID,
		Name:                "Test Competency",
		Description:         "A test competency for unit testing",
		DescriptionVeryBad:  "Very Bad description for testing",
		DescriptionBad:      "Bad description for testing",
		DescriptionOk:       "Ok description for testing",
		DescriptionGood:     "Good description for testing",
		DescriptionVeryGood: "Very Good description for testing",
		Weight:              5,
	}

	err := CreateCompetency(suite.ctx, coursePhaseID, newCompetency)
	assert.NoError(suite.T(), err, "Creating competency should not produce an error")

	// Verify the competency was created by listing all competencies and checking if our new one exists
	competencies, err := ListCompetencies(suite.ctx)
	assert.NoError(suite.T(), err)

	found := false
	for _, competency := range competencies {
		if competency.Name == newCompetency.Name {
			found = true
			assert.Equal(suite.T(), newCompetency.Description, competency.Description.String)
			assert.Equal(suite.T(), newCompetency.DescriptionVeryBad, competency.DescriptionVeryBad)
			assert.Equal(suite.T(), newCompetency.DescriptionBad, competency.DescriptionBad)
			assert.Equal(suite.T(), newCompetency.DescriptionOk, competency.DescriptionOk)
			assert.Equal(suite.T(), newCompetency.DescriptionGood, competency.DescriptionGood)
			assert.Equal(suite.T(), newCompetency.DescriptionVeryGood, competency.DescriptionVeryGood)
			assert.Equal(suite.T(), newCompetency.Weight, competency.Weight)
			break
		}
	}
	assert.True(suite.T(), found, "Created competency should be found in the list")
}

func (suite *CompetencyServiceTestSuite) TestUpdateCompetency() {
	// First, create a competency to update
	categoryID := uuid.MustParse("9107c0aa-15b7-4967-bf62-6fa131f08bee")    // Fundamentals in Software Engineering category
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02") // Dev Application phase from test data

	createRequest := competencyDTO.CreateCompetencyRequest{
		CategoryID:          categoryID,
		Name:                "Competency to Update",
		Description:         "Original description",
		DescriptionVeryBad:  "Original very bad description",
		DescriptionBad:      "Original bad description",
		DescriptionOk:       "Original ok description",
		DescriptionGood:     "Original good description",
		DescriptionVeryGood: "Original very good description",
		Weight:              3,
	}

	err := CreateCompetency(suite.ctx, coursePhaseID, createRequest)
	assert.NoError(suite.T(), err)

	// Find the created competency
	competencies, err := ListCompetencies(suite.ctx)
	assert.NoError(suite.T(), err)

	var createdCompetency db.Competency
	found := false
	for _, competency := range competencies {
		if competency.Name == createRequest.Name {
			createdCompetency = competency
			found = true
			break
		}
	}
	assert.True(suite.T(), found, "Created competency should be found")

	// Update the competency
	updateRequest := competencyDTO.UpdateCompetencyRequest{
		Name:                "Updated Competency Name",
		CategoryID:          categoryID,
		Description:         "Updated description",
		DescriptionVeryBad:  "Updated very bad description",
		DescriptionBad:      "Updated bad description",
		DescriptionOk:       "Updated ok description",
		DescriptionGood:     "Updated good description",
		DescriptionVeryGood: "Updated very good description",
		Weight:              8,
	}

	// Use the same course phase ID for updating
	err = UpdateCompetency(suite.ctx, createdCompetency.ID, coursePhaseID, updateRequest)
	assert.NoError(suite.T(), err, "Updating competency should not produce an error")

	// Verify the update
	var updatedCompetency db.Competency
	updatedCompetency, err = GetCompetency(suite.ctx, createdCompetency.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), updateRequest.Name, updatedCompetency.Name)
	assert.Equal(suite.T(), updateRequest.Description, updatedCompetency.Description.String)
	assert.Equal(suite.T(), updateRequest.DescriptionVeryBad, updatedCompetency.DescriptionVeryBad)
	assert.Equal(suite.T(), updateRequest.DescriptionBad, updatedCompetency.DescriptionBad)
	assert.Equal(suite.T(), updateRequest.DescriptionOk, updatedCompetency.DescriptionOk)
	assert.Equal(suite.T(), updateRequest.DescriptionGood, updatedCompetency.DescriptionGood)
	assert.Equal(suite.T(), updateRequest.DescriptionVeryGood, updatedCompetency.DescriptionVeryGood)
	assert.Equal(suite.T(), updateRequest.Weight, updatedCompetency.Weight)
}

func (suite *CompetencyServiceTestSuite) TestUpdateNonExistentCompetency() {
	nonExistentID := uuid.New()
	categoryID := uuid.MustParse("815b159b-cab3-49b4-8060-c4722d59241d")

	updateRequest := competencyDTO.UpdateCompetencyRequest{
		Name:                "Non-existent Competency",
		CategoryID:          categoryID,
		Description:         "This should not fail but affect 0 rows",
		DescriptionVeryBad:  "Non-existent very bad description",
		DescriptionBad:      "Non-existent bad description",
		DescriptionOk:       "Non-existent ok description",
		DescriptionGood:     "Non-existent good description",
		DescriptionVeryGood: "Non-existent very good description",
		Weight:              1,
	}

	// Use a mock course phase ID for testing
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	err := UpdateCompetency(suite.ctx, nonExistentID, coursePhaseID, updateRequest)
	assert.Error(suite.T(), err, "Updating non-existent competency should produce an error")
}

func (suite *CompetencyServiceTestSuite) TestDeleteCompetency() {
	// First, create a competency to delete
	categoryID := uuid.MustParse("815b159b-cab3-49b4-8060-c4722d59241d")    // User Interface category
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02") // Dev Application phase from test data

	createRequest := competencyDTO.CreateCompetencyRequest{
		CategoryID:          categoryID,
		Name:                "Competency to Delete",
		Description:         "This competency will be deleted",
		DescriptionVeryBad:  "Very Bad description for deletion",
		DescriptionBad:      "Bad description for deletion",
		DescriptionOk:       "Ok description for deletion",
		DescriptionGood:     "Good description for deletion",
		DescriptionVeryGood: "Very Good description for deletion",
		Weight:              2,
	}

	err := CreateCompetency(suite.ctx, coursePhaseID, createRequest)
	assert.NoError(suite.T(), err)

	// Find the created competency
	competencies, err := ListCompetencies(suite.ctx)
	assert.NoError(suite.T(), err)

	var competencyToDelete db.Competency
	found := false
	for _, competency := range competencies {
		if competency.Name == createRequest.Name {
			competencyToDelete = competency
			found = true
			break
		}
	}
	assert.True(suite.T(), found, "Created competency should be found")

	// Delete the competency
	err = DeleteCompetency(suite.ctx, competencyToDelete.ID, coursePhaseID)
	assert.NoError(suite.T(), err, "Deleting competency should not produce an error")

	// Verify the competency was deleted
	_, err = GetCompetency(suite.ctx, competencyToDelete.ID)
	assert.Error(suite.T(), err, "Getting deleted competency should produce an error")
}

func (suite *CompetencyServiceTestSuite) TestDeleteNonExistentCompetency() {
	nonExistentID := uuid.New()

	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	err := DeleteCompetency(suite.ctx, nonExistentID, coursePhaseID)
	assert.NoError(suite.T(), err, "Deleting non-existent competency should not produce an error (affects 0 rows)")
}

func (suite *CompetencyServiceTestSuite) TestCreateCompetencyWithInvalidCategory() {
	nonExistentCategoryID := uuid.New()

	invalidCompetency := competencyDTO.CreateCompetencyRequest{
		CategoryID:          nonExistentCategoryID,
		Name:                "Invalid Competency",
		Description:         "This should fail due to invalid category",
		DescriptionVeryBad:  "Very Bad description for invalid category",
		DescriptionBad:      "Bad description for invalid category",
		DescriptionOk:       "Ok description for invalid category",
		DescriptionGood:     "Good description for invalid category",
		DescriptionVeryGood: "Very Good description for invalid category",
		Weight:              1,
	}

	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	err := CreateCompetency(suite.ctx, coursePhaseID, invalidCompetency)
	assert.Error(suite.T(), err, "Creating competency with invalid category should produce an error")
}

func TestCompetencyServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CompetencyServiceTestSuite))
}
