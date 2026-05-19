package coursePhaseConfig

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig/coursePhaseConfigDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

// Helper function to create a test course phase config request
func createTestCoursePhaseConfigRequest(schemaID, coursePhaseID uuid.UUID) coursePhaseConfigDTO.CreateOrUpdateCoursePhaseConfigRequest {
	now := time.Now()
	selfSchemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")  // From test data
	peerSchemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")  // From test data
	tutorSchemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440003") // From test data

	return coursePhaseConfigDTO.CreateOrUpdateCoursePhaseConfigRequest{
		AssessmentSchemaID:       schemaID,
		CoursePhaseID:            coursePhaseID,
		Start:                    now,
		Deadline:                 now.Add(7 * 24 * time.Hour),
		SelfEvaluationEnabled:    false,
		SelfEvaluationSchema:     selfSchemaID,
		SelfEvaluationStart:      now,
		SelfEvaluationDeadline:   now.Add(14 * 24 * time.Hour),
		PeerEvaluationEnabled:    false,
		PeerEvaluationSchema:     peerSchemaID,
		PeerEvaluationStart:      now,
		PeerEvaluationDeadline:   now.Add(21 * 24 * time.Hour),
		TutorEvaluationEnabled:   false,
		TutorEvaluationSchema:    tutorSchemaID,
		TutorEvaluationStart:     now,
		TutorEvaluationDeadline:  now.Add(28 * 24 * time.Hour),
		EvaluationResultsVisible: false,
		// GradeSuggestionVisible and ActionItemsVisible are nil by default (pointer fields)
	}
}

type CoursePhaseConfigServiceTestSuite struct {
	suite.Suite
	suiteCtx                 context.Context
	cleanup                  func()
	coursePhaseConfigService CoursePhaseConfigService
	testCoursePhaseID        uuid.UUID
}

func (suite *CoursePhaseConfigServiceTestSuite) SetupSuite() {
	if testing.Short() {
		suite.T().Skip("skipping db-backed course phase config service tests in short mode")
	}
	defer func() {
		if r := recover(); r != nil {
			suite.T().Skipf("skipping db-backed course phase config service tests: %v", r)
		}
	}()

	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/coursePhaseConfig.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Skipf("skipping db-backed course phase config service tests: %v", err)
	}
	suite.cleanup = cleanup
	suite.coursePhaseConfigService = CoursePhaseConfigService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	CoursePhaseConfigSingleton = &suite.coursePhaseConfigService

	// Initialize assessmentSchemas singleton for validation (needed for CheckPhaseHasAssessmentData)
	assessmentSchemas.AssessmentSchemaServiceSingleton = assessmentSchemas.NewAssessmentSchemaService(*testDB.Queries, testDB.Conn)

	// Generate a test course phase ID and insert it with a schema
	suite.testCoursePhaseID = uuid.New()
	schemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")     // From our test data
	selfSchemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001") // Self assessment schema
	peerSchemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002") // Peer assessment schema

	// Insert a course phase config entry to enable updates
	_, err = testDB.Conn.Exec(suite.suiteCtx,
		`INSERT INTO course_phase_config (assessment_schema_id, course_phase_id, self_evaluation_schema, peer_evaluation_schema) 
		 VALUES ($1, $2, $3, $4)`,
		schemaID, suite.testCoursePhaseID, selfSchemaID, peerSchemaID)
	if err != nil {
		suite.T().Fatalf("Failed to insert test course phase config: %v", err)
	}
}

func (suite *CoursePhaseConfigServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *CoursePhaseConfigServiceTestSuite) TestGetCoursePhaseConfig() {
	// Test getting course phase config
	testID := uuid.New()
	schemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	// Insert a course phase config entry first
	_, err := suite.coursePhaseConfigService.conn.Exec(suite.suiteCtx,
		"INSERT INTO course_phase_config (assessment_schema_id, course_phase_id) VALUES ($1, $2)",
		schemaID, testID)
	assert.NoError(suite.T(), err)

	config, err := GetCoursePhaseConfig(suite.suiteCtx, testID)
	assert.NoError(suite.T(), err, "Should be able to get course phase config")
	assert.NotNil(suite.T(), config, "Config should not be nil")
}

func (suite *CoursePhaseConfigServiceTestSuite) TestCreateOrUpdateCoursePhaseConfig_DefaultVisibilitySettings() {
	// Test that when GradeSuggestionVisible and ActionItemsVisible are nil (not provided),
	// they default to TRUE as per database defaults
	testID := uuid.New()
	schemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	// Create request with nil visibility settings (using pointer fields)
	req := createTestCoursePhaseConfigRequest(schemaID, testID)
	// Both fields are nil by default in the test helper

	err := CreateOrUpdateCoursePhaseConfig(suite.suiteCtx, testID, req)
	assert.NoError(suite.T(), err, "Should successfully create config with nil visibility settings")

	// Verify the config was created with TRUE defaults
	config, err := GetCoursePhaseConfig(suite.suiteCtx, testID)
	assert.NoError(suite.T(), err, "Should be able to get the created config")
	assert.True(suite.T(), config.GradeSuggestionVisible, "GradeSuggestionVisible should default to TRUE")
	assert.True(suite.T(), config.ActionItemsVisible, "ActionItemsVisible should default to TRUE")
}

func (suite *CoursePhaseConfigServiceTestSuite) TestCreateOrUpdateCoursePhaseConfig_ExplicitFalseVisibilitySettings() {
	// Test that when GradeSuggestionVisible and ActionItemsVisible are explicitly set to false,
	// they are stored as FALSE (not overridden by defaults)
	testID := uuid.New()
	schemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	// Create request with explicit false values
	req := createTestCoursePhaseConfigRequest(schemaID, testID)
	falseValue := false
	req.GradeSuggestionVisible = &falseValue
	req.ActionItemsVisible = &falseValue

	err := CreateOrUpdateCoursePhaseConfig(suite.suiteCtx, testID, req)
	assert.NoError(suite.T(), err, "Should successfully create config with explicit false values")

	// Verify the config was created with FALSE values
	config, err := GetCoursePhaseConfig(suite.suiteCtx, testID)
	assert.NoError(suite.T(), err, "Should be able to get the created config")
	assert.False(suite.T(), config.GradeSuggestionVisible, "GradeSuggestionVisible should be FALSE")
	assert.False(suite.T(), config.ActionItemsVisible, "ActionItemsVisible should be FALSE")
}

func (suite *CoursePhaseConfigServiceTestSuite) TestCreateOrUpdateCoursePhaseConfig_ExplicitTrueVisibilitySettings() {
	// Test that when GradeSuggestionVisible and ActionItemsVisible are explicitly set to true,
	// they are stored as TRUE
	testID := uuid.New()
	schemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	// Create request with explicit true values
	req := createTestCoursePhaseConfigRequest(schemaID, testID)
	trueValue := true
	req.GradeSuggestionVisible = &trueValue
	req.ActionItemsVisible = &trueValue

	err := CreateOrUpdateCoursePhaseConfig(suite.suiteCtx, testID, req)
	assert.NoError(suite.T(), err, "Should successfully create config with explicit true values")

	// Verify the config was created with TRUE values
	config, err := GetCoursePhaseConfig(suite.suiteCtx, testID)
	assert.NoError(suite.T(), err, "Should be able to get the created config")
	assert.True(suite.T(), config.GradeSuggestionVisible, "GradeSuggestionVisible should be TRUE")
	assert.True(suite.T(), config.ActionItemsVisible, "ActionItemsVisible should be TRUE")
}

func (suite *CoursePhaseConfigServiceTestSuite) TestUpdateCoursePhaseConfig_PreservesDefaults() {
	// Test that updating a config with nil visibility settings preserves the defaults (TRUE)
	testID := uuid.New()
	schemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	// First, create with defaults
	req := createTestCoursePhaseConfigRequest(schemaID, testID)
	err := CreateOrUpdateCoursePhaseConfig(suite.suiteCtx, testID, req)
	assert.NoError(suite.T(), err)

	// Verify initial defaults
	config, err := GetCoursePhaseConfig(suite.suiteCtx, testID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), config.GradeSuggestionVisible)
	assert.True(suite.T(), config.ActionItemsVisible)

	// Update with nil values (should preserve TRUE)
	updateReq := createTestCoursePhaseConfigRequest(schemaID, testID)
	err = CreateOrUpdateCoursePhaseConfig(suite.suiteCtx, testID, updateReq)
	assert.NoError(suite.T(), err, "Should successfully update config")

	// Verify values are still TRUE (preserved by COALESCE)
	updatedConfig, err := GetCoursePhaseConfig(suite.suiteCtx, testID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), updatedConfig.GradeSuggestionVisible, "GradeSuggestionVisible should remain TRUE")
	assert.True(suite.T(), updatedConfig.ActionItemsVisible, "ActionItemsVisible should remain TRUE")
}

func (suite *CoursePhaseConfigServiceTestSuite) TestUpdateCoursePhaseConfig_CanToggleFalseToTrue() {
	// Test that we can update from FALSE to TRUE
	testID := uuid.New()
	schemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	// Create with explicit false values
	req := createTestCoursePhaseConfigRequest(schemaID, testID)
	falseValue := false
	req.GradeSuggestionVisible = &falseValue
	req.ActionItemsVisible = &falseValue
	err := CreateOrUpdateCoursePhaseConfig(suite.suiteCtx, testID, req)
	assert.NoError(suite.T(), err)

	// Verify initial values are FALSE
	config, err := GetCoursePhaseConfig(suite.suiteCtx, testID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), config.GradeSuggestionVisible)
	assert.False(suite.T(), config.ActionItemsVisible)

	// Update to TRUE
	updateReq := createTestCoursePhaseConfigRequest(schemaID, testID)
	trueValue := true
	updateReq.GradeSuggestionVisible = &trueValue
	updateReq.ActionItemsVisible = &trueValue
	err = CreateOrUpdateCoursePhaseConfig(suite.suiteCtx, testID, updateReq)
	assert.NoError(suite.T(), err)

	// Verify values are now TRUE
	updatedConfig, err := GetCoursePhaseConfig(suite.suiteCtx, testID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), updatedConfig.GradeSuggestionVisible, "GradeSuggestionVisible should be TRUE")
	assert.True(suite.T(), updatedConfig.ActionItemsVisible, "ActionItemsVisible should be TRUE")
}

func (suite *CoursePhaseConfigServiceTestSuite) TestReleaseAndUnreleaseResults() {
	testID := uuid.New()
	schemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	_, err := suite.coursePhaseConfigService.conn.Exec(suite.suiteCtx,
		"INSERT INTO course_phase_config (assessment_schema_id, course_phase_id, results_released) VALUES ($1, $2, $3)",
		schemaID, testID, false)
	assert.NoError(suite.T(), err)

	err = ReleaseResults(suite.suiteCtx, testID)
	assert.NoError(suite.T(), err)

	config, err := GetCoursePhaseConfig(suite.suiteCtx, testID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), config.ResultsReleased)

	err = UnreleaseResults(suite.suiteCtx, testID)
	assert.NoError(suite.T(), err)

	config, err = GetCoursePhaseConfig(suite.suiteCtx, testID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), config.ResultsReleased)
}

func (suite *CoursePhaseConfigServiceTestSuite) TestCreateOrUpdateCoursePhaseConfig_CannotChangeSchemaWithData() {
	// Test that we cannot change schemas when assessment or evaluation data exists
	testID := uuid.New()
	oldSchemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	newSchemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

	// Create initial config
	req := createTestCoursePhaseConfigRequest(oldSchemaID, testID)
	err := CreateOrUpdateCoursePhaseConfig(suite.suiteCtx, testID, req)
	assert.NoError(suite.T(), err, "Should successfully create initial config")

	// Create test data structure: category -> competency -> assessment
	categoryID := uuid.New()
	competencyID := uuid.New()
	participationID := uuid.New()

	// Insert category for old schema
	_, err = suite.coursePhaseConfigService.conn.Exec(suite.suiteCtx,
		`INSERT INTO category (id, name, assessment_schema_id) VALUES ($1, $2, $3)`,
		categoryID, "Test Category", oldSchemaID)
	assert.NoError(suite.T(), err)

	// Insert competency
	_, err = suite.coursePhaseConfigService.conn.Exec(suite.suiteCtx,
		`INSERT INTO competency (id, category_id, name) VALUES ($1, $2, $3)`,
		competencyID, categoryID, "Test Competency")
	assert.NoError(suite.T(), err)

	// Insert assessment data to block schema change
	assessmentID := uuid.New()
	_, err = suite.coursePhaseConfigService.conn.Exec(suite.suiteCtx,
		`INSERT INTO assessment (id, course_participation_id, course_phase_id, competency_id, score_level) 
		 VALUES ($1, $2, $3, $4, $5)`,
		assessmentID, participationID, testID, competencyID, "good")
	assert.NoError(suite.T(), err)

	// Try to update config with new schema - should fail
	updateReq := createTestCoursePhaseConfigRequest(newSchemaID, testID)
	err = CreateOrUpdateCoursePhaseConfig(suite.suiteCtx, testID, updateReq)
	assert.Error(suite.T(), err, "Should fail to change schema when assessment data exists")
	assert.Equal(suite.T(), ErrCannotChangeSchemaWithData, err, "Should return correct error type")

	// Verify schema was not changed
	config, err := GetCoursePhaseConfig(suite.suiteCtx, testID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), oldSchemaID, config.AssessmentSchemaID, "Schema should not have changed")
}

func (suite *CoursePhaseConfigServiceTestSuite) TestCreateOrUpdateCoursePhaseConfig_CanChangeSchemaWithoutData() {
	// Test that we CAN change schemas when no assessment or evaluation data exists
	testID := uuid.New()
	oldSchemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	newSchemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

	// Create initial config
	req := createTestCoursePhaseConfigRequest(oldSchemaID, testID)
	err := CreateOrUpdateCoursePhaseConfig(suite.suiteCtx, testID, req)
	assert.NoError(suite.T(), err, "Should successfully create initial config")

	// Update config with new schema - should succeed since no data exists
	updateReq := createTestCoursePhaseConfigRequest(newSchemaID, testID)
	err = CreateOrUpdateCoursePhaseConfig(suite.suiteCtx, testID, updateReq)
	assert.NoError(suite.T(), err, "Should successfully change schema when no assessment data exists")

	// Verify schema was changed
	config, err := GetCoursePhaseConfig(suite.suiteCtx, testID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newSchemaID, config.AssessmentSchemaID, "Schema should have changed")
}

// Note: GetTeamsForCoursePhase testing is limited because it requires external HTTP calls
// to the core service. The router tests above verify the endpoint behavior and error handling.
// The service function includes safe type assertions that prevent runtime panics when
// the external API returns unexpected data structures.
func TestCoursePhaseConfigServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CoursePhaseConfigServiceTestSuite))
}
