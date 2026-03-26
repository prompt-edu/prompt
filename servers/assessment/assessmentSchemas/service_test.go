package assessmentSchemas

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas/assessmentSchemaDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type AssessmentSchemaServiceTestSuite struct {
	suite.Suite
	suiteCtx                context.Context
	cleanup                 func()
	assessmentSchemaService AssessmentSchemaService
}

func (suite *AssessmentSchemaServiceTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/categories.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.assessmentSchemaService = AssessmentSchemaService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	AssessmentSchemaServiceSingleton = &suite.assessmentSchemaService
}

func (suite *AssessmentSchemaServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *AssessmentSchemaServiceTestSuite) TestListAssessmentSchemas() {
	schemas, err := ListAssessmentSchemas(suite.suiteCtx)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(schemas), 0, "Expected at least one assessment schema")

	// Check that the default schema from the dump exists
	found := false
	for _, schema := range schemas {
		if schema.Name == "Intro Course Assessment Schema" {
			found = true
			assert.Equal(suite.T(), "This is the default assessment schema.", schema.Description)
			break
		}
	}
	assert.True(suite.T(), found, "Default assessment schema should be found")
}

func (suite *AssessmentSchemaServiceTestSuite) TestGetAssessmentSchema() {
	// Use the default assessment schema ID from the database dump
	defaultID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	schema, err := GetAssessmentSchema(suite.suiteCtx, defaultID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), defaultID, schema.ID)
	assert.Equal(suite.T(), "Intro Course Assessment Schema", schema.Name)
	assert.Equal(suite.T(), "This is the default assessment schema.", schema.Description)
}

func (suite *AssessmentSchemaServiceTestSuite) TestCreateAssessmentSchema() {
	req := assessmentSchemaDTO.CreateAssessmentSchemaRequest{
		Name:        "Test Schema",
		Description: "Test Description",
	}

	schema, err := CreateAssessmentSchema(suite.suiteCtx, req)
	assert.NoError(suite.T(), err, "Should be able to create assessment schema")
	assert.NotEqual(suite.T(), uuid.Nil, schema.ID, "Should have a valid ID")
	assert.Equal(suite.T(), req.Name, schema.Name, "Name should match")
	assert.Equal(suite.T(), req.Description, schema.Description, "Description should match")
}

func (suite *AssessmentSchemaServiceTestSuite) TestGetAssessmentSchemaNotFound() {
	nonExistentID := uuid.New()

	_, err := GetAssessmentSchema(suite.suiteCtx, nonExistentID)
	assert.Error(suite.T(), err, "Should return error for non-existent schema")
}

func (suite *AssessmentSchemaServiceTestSuite) TestUpdateAssessmentSchema() {
	// First create a schema
	createReq := assessmentSchemaDTO.CreateAssessmentSchemaRequest{
		Name:        "Original Schema",
		Description: "Original Description",
	}

	schema, err := CreateAssessmentSchema(suite.suiteCtx, createReq)
	assert.NoError(suite.T(), err, "Should be able to create assessment schema")

	// Now update it
	updateReq := assessmentSchemaDTO.UpdateAssessmentSchemaRequest{
		Name:        "Updated Schema",
		Description: "Updated Description",
	}

	err = UpdateAssessmentSchema(suite.suiteCtx, schema.ID, updateReq)
	assert.NoError(suite.T(), err, "Should be able to update assessment schema")

	// Verify the update
	updatedSchema, err := GetAssessmentSchema(suite.suiteCtx, schema.ID)
	assert.NoError(suite.T(), err, "Should be able to retrieve updated schema")
	assert.Equal(suite.T(), updateReq.Name, updatedSchema.Name, "Name should be updated")
	assert.Equal(suite.T(), updateReq.Description, updatedSchema.Description, "Description should be updated")
}

func (suite *AssessmentSchemaServiceTestSuite) TestUpdateAssessmentSchemaNotFound() {
	nonExistentID := uuid.New()
	updateReq := assessmentSchemaDTO.UpdateAssessmentSchemaRequest{
		Name:        "Updated Schema",
		Description: "Updated Description",
	}

	err := UpdateAssessmentSchema(suite.suiteCtx, nonExistentID, updateReq)
	assert.Error(suite.T(), err, "Should return error for non-existent schema")
}

func (suite *AssessmentSchemaServiceTestSuite) TestDeleteAssessmentSchema() {
	// First create a schema
	createReq := assessmentSchemaDTO.CreateAssessmentSchemaRequest{
		Name:        "Schema to Delete",
		Description: "Will be deleted",
	}

	schema, err := CreateAssessmentSchema(suite.suiteCtx, createReq)
	assert.NoError(suite.T(), err, "Should be able to create assessment schema")

	// Now delete it
	err = DeleteAssessmentSchema(suite.suiteCtx, schema.ID)
	assert.NoError(suite.T(), err, "Should be able to delete assessment schema")

	// Verify it's gone
	_, err = GetAssessmentSchema(suite.suiteCtx, schema.ID)
	assert.Error(suite.T(), err, "Should return error for deleted schema")
}

func (suite *AssessmentSchemaServiceTestSuite) TestDeleteAssessmentSchemaNotFound() {
	nonExistentID := uuid.New()

	err := DeleteAssessmentSchema(suite.suiteCtx, nonExistentID)
	// This may or may not error depending on implementation - test that it doesn't panic
	_ = err // Ignore the error for this test
	assert.NotPanics(suite.T(), func() {
		_ = DeleteAssessmentSchema(suite.suiteCtx, nonExistentID)
	})
}

func (suite *AssessmentSchemaServiceTestSuite) TestGetCoursePhasesByAssessmentSchema() {
	defaultID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	coursePhases, err := GetCoursePhasesByAssessmentSchema(suite.suiteCtx, defaultID)
	assert.NoError(suite.T(), err, "Should be able to get course phases")
	assert.NotNil(suite.T(), coursePhases, "Should return non-nil slice")
	assert.IsType(suite.T(), []uuid.UUID{}, coursePhases, "Should return correct type")
}

func TestAssessmentSchemaServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AssessmentSchemaServiceTestSuite))
}
