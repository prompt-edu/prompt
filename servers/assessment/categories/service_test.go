package categories

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/categories/categoryDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/testutils"
)

type CategoryServiceTestSuite struct {
	suite.Suite
	suiteCtx        context.Context
	cleanup         func()
	mockCoreCleanup func()
	categoryService CategoryService
}

func (suite *CategoryServiceTestSuite) SetupTest() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/categories.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup

	// Set up mock core service
	_, mockCleanup := testutils.SetupMockCoreService()
	suite.mockCoreCleanup = mockCleanup

	suite.categoryService = CategoryService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	CategoryServiceSingleton = &suite.categoryService

	// Initialize other service modules needed for schema copy logic
	router := gin.New()
	group := router.Group("")
	assessmentSchemas.InitAssessmentSchemaModule(group, *testDB.Queries, testDB.Conn)
	coursePhaseConfig.InitCoursePhaseConfigModule(group, *testDB.Queries, testDB.Conn)
}

func (suite *CategoryServiceTestSuite) TearDownTest() {
	if suite.mockCoreCleanup != nil {
		suite.mockCoreCleanup()
	}
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *CategoryServiceTestSuite) TestListCategories() {
	categories, err := ListCategories(suite.suiteCtx)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(categories), 0, "Expected at least one category")
	for _, c := range categories {
		assert.NotEmpty(suite.T(), c.ID, "Category ID should not be empty")
		assert.NotEmpty(suite.T(), c.Name, "Category Name should not be empty")
		assert.Greater(suite.T(), c.Weight, int32(0), "Category Weight should be positive")
	}
}

func (suite *CategoryServiceTestSuite) TestGetCategory() {
	id := uuid.MustParse("25f1c984-ba31-4cf2-aa8e-5662721bf44e")
	c, err := GetCategory(suite.suiteCtx, id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), id, c.ID, "Category ID should match")
	assert.Equal(suite.T(), "Version Control", c.Name, "Category name should match")
}

func (suite *CategoryServiceTestSuite) TestGetCategoryNotFound() {
	nonExistent := uuid.New()
	_, err := GetCategory(suite.suiteCtx, nonExistent)
	assert.Error(suite.T(), err, "Expected error for non-existent category")
}

func (suite *CategoryServiceTestSuite) TestCreateCategory() {
	// Use the default assessment schema ID from the database dump
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02") // Dev Application phase from test data

	req := categoryDTO.CreateCategoryRequest{
		Name:               "Test Category",
		ShortName:          "TC",
		Description:        "A test category",
		Weight:             5,
		AssessmentSchemaID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), // From test data
	}
	err := CreateCategory(suite.suiteCtx, coursePhaseID, req)
	assert.NoError(suite.T(), err, "Creating category should not produce an error")

	cats, err := ListCategories(suite.suiteCtx)
	assert.NoError(suite.T(), err)
	found := false
	for _, c := range cats {
		if c.Name == req.Name {
			found = true
			assert.Equal(suite.T(), req.Description, c.Description.String)
			assert.Equal(suite.T(), req.Weight, c.Weight)
			break
		}
	}
	assert.True(suite.T(), found, "Created category should be found in the list")
}

func (suite *CategoryServiceTestSuite) TestUpdateCategory() {
	// use existing category from dump
	id := uuid.MustParse("815b159b-cab3-49b4-8060-c4722d59241d")
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02") // Dev Application phase from test data

	req := categoryDTO.UpdateCategoryRequest{
		Name:               "Updated Category",
		ShortName:          "UC",
		Description:        "Updated description",
		Weight:             2,
		AssessmentSchemaID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), // From test data
	}
	err := UpdateCategory(suite.suiteCtx, id, coursePhaseID, req)
	assert.NoError(suite.T(), err, "Updating category should not produce an error")

	categories, err := ListCategoriesForCoursePhase(suite.suiteCtx, coursePhaseID)
	assert.NoError(suite.T(), err)

	foundUpdated := false
	for _, c := range categories {
		if c.Name == req.Name {
			foundUpdated = true
			assert.Equal(suite.T(), req.Description, c.Description.String)
			assert.Equal(suite.T(), req.Weight, c.Weight)
			break
		}
	}

	assert.True(suite.T(), foundUpdated, "Updated category should be visible for the course phase")
}

func (suite *CategoryServiceTestSuite) TestUpdateNonExistentCategory() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02") // Dev Application phase from test data
	id := uuid.New()

	req := categoryDTO.UpdateCategoryRequest{
		Name:               "Non-existent",
		ShortName:          "NE",
		Description:        "Should not fail",
		Weight:             1,
		AssessmentSchemaID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), // From test data
	}
	err := UpdateCategory(suite.suiteCtx, id, coursePhaseID, req)
	assert.NoError(suite.T(), err, "Updating non-existent category should not error")
}

func (suite *CategoryServiceTestSuite) TestDeleteCategory() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02") // Dev Application phase from test data

	// create category to delete
	reqCreate := categoryDTO.CreateCategoryRequest{
		Name:               "To Delete",
		ShortName:          "TD",
		Description:        "To be deleted",
		Weight:             1,
		AssessmentSchemaID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), // From test data
	}
	err := CreateCategory(suite.suiteCtx, coursePhaseID, reqCreate)
	assert.NoError(suite.T(), err)

	cats, err := ListCategories(suite.suiteCtx)
	assert.NoError(suite.T(), err)
	var toDeleteID uuid.UUID
	for _, c := range cats {
		if c.Name == reqCreate.Name {
			toDeleteID = c.ID
			break
		}
	}
	err = DeleteCategory(suite.suiteCtx, toDeleteID, coursePhaseID)
	assert.NoError(suite.T(), err, "Deleting category should not produce an error")
	_, err = GetCategory(suite.suiteCtx, toDeleteID)
	assert.Error(suite.T(), err, "Getting deleted category should produce an error")
}

func (suite *CategoryServiceTestSuite) TestDeleteNonExistentCategory() {
	nonExistent := uuid.New()
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	err := DeleteCategory(suite.suiteCtx, nonExistent, coursePhaseID)
	assert.NoError(suite.T(), err, "Deleting non-existent category should not error")
}

func (suite *CategoryServiceTestSuite) TestGetCategoriesWithCompetencies() {
	// Use a known course phase ID from the test data
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02") // Dev Application phase

	catsWithComp, err := GetCategoriesWithCompetencies(suite.suiteCtx, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(catsWithComp), 0, "Should return categories with competencies")

	// Verify the structure of returned data
	for _, cat := range catsWithComp {
		assert.NotEmpty(suite.T(), cat.ID, "Category ID should not be empty")
		assert.NotEmpty(suite.T(), cat.Name, "Category name should not be empty")
		assert.GreaterOrEqual(suite.T(), cat.Weight, int32(1), "Category weight should be at least 1")
		assert.NotNil(suite.T(), cat.Competencies, "Competencies should not be nil")

		// Verify competencies structure if any exist
		for _, comp := range cat.Competencies {
			assert.NotEmpty(suite.T(), comp.ID, "Competency ID should not be empty")
			assert.NotEmpty(suite.T(), comp.Name, "Competency name should not be empty")
			assert.Equal(suite.T(), cat.ID, comp.CategoryID, "Competency should belong to the category")
		}
	}
}

func TestCategoryServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CategoryServiceTestSuite))
}
