package coursePhase

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CoursePhaseTestSuite struct {
	suite.Suite
	ctx                context.Context
	cleanup            func()
	coursePhaseService CoursePhaseService
}

func (suite *CoursePhaseTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Set up PostgreSQL container
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/course_phase_test.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}

	suite.cleanup = cleanup
	suite.coursePhaseService = CoursePhaseService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	resolution.InitResolutionModule("localhost:8080")

	CoursePhaseServiceSingleton = &suite.coursePhaseService
}

func (suite *CoursePhaseTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *CoursePhaseTestSuite) TestGetCoursePhaseByID() {
	id := uuid.MustParse("3d1f3b00-87f3-433b-a713-178c4050411b")
	coursePhase, err := GetCoursePhaseByID(suite.ctx, id)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Test", coursePhase.Name, "Expected course phase name to match")
	assert.False(suite.T(), coursePhase.IsInitialPhase, "Expected course phase to not be an initial phase")
	assert.Equal(suite.T(), id, coursePhase.ID, "Expected course phase ID to match")
}

func (suite *CoursePhaseTestSuite) TestUpdateCoursePhase() {
	id := uuid.MustParse("3d1f3b00-87f3-433b-a713-178c4050411b")
	jsonData := `{"updated_key": "updated_value"}`
	var restrictedData meta.MetaData
	err := json.Unmarshal([]byte(jsonData), &restrictedData)
	assert.NoError(suite.T(), err)

	jsonData = `{"updated_key2": "updated_value"}`
	var studentData meta.MetaData
	err = json.Unmarshal([]byte(jsonData), &studentData)
	assert.NoError(suite.T(), err)

	update := coursePhaseDTO.UpdateCoursePhase{
		ID:                  id,
		Name:                pgtype.Text{Valid: true, String: "Updated Phase"},
		RestrictedData:      restrictedData,
		StudentReadableData: studentData,
	}

	err = UpdateCoursePhase(suite.ctx, update)
	assert.NoError(suite.T(), err)

	// Verify update
	updatedCoursePhase, err := GetCoursePhaseByID(suite.ctx, id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Phase", updatedCoursePhase.Name, "Expected updated course phase name to match")
	assert.False(suite.T(), updatedCoursePhase.IsInitialPhase, "Expected updated course phase to be an initial phase")
	assert.Equal(suite.T(), meta.MetaData{"test-key": "test-value", "updated_key": "updated_value"}, updatedCoursePhase.RestrictedData, "Expected metadata to match updated data including the old data")
	assert.Equal(suite.T(), meta.MetaData{"updated_key2": "updated_value"}, updatedCoursePhase.StudentReadableData, "Expected student readable data to match updated data")
}

func (suite *CoursePhaseTestSuite) TestUpdateCoursePhaseWithMetaDataOverride() {
	id := uuid.MustParse("3d1f3b00-87f3-433b-a713-178c4050411b")
	jsonData := `{"test-key": "test-value-new", "updated_key": "updated_value"}`
	var data meta.MetaData
	err := json.Unmarshal([]byte(jsonData), &data)
	assert.NoError(suite.T(), err)

	jsonData = `{"updated_key2": "updated_value"}`
	var studentData meta.MetaData
	err = json.Unmarshal([]byte(jsonData), &studentData)
	assert.NoError(suite.T(), err)

	update := coursePhaseDTO.UpdateCoursePhase{
		ID:                  id,
		Name:                pgtype.Text{Valid: true, String: "Updated Phase"},
		RestrictedData:      data,
		StudentReadableData: studentData,
	}

	err = UpdateCoursePhase(suite.ctx, update)
	assert.NoError(suite.T(), err)

	// Verify update
	updatedCoursePhase, err := GetCoursePhaseByID(suite.ctx, id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Phase", updatedCoursePhase.Name, "Expected updated course phase name to match")
	assert.Equal(suite.T(), meta.MetaData{"test-key": "test-value-new", "updated_key": "updated_value"}, updatedCoursePhase.RestrictedData, "Expected metadata to match updated data including the old data")
	assert.Equal(suite.T(), meta.MetaData{"updated_key2": "updated_value"}, updatedCoursePhase.StudentReadableData, "Expected student readable data to match updated data")
}

func (suite *CoursePhaseTestSuite) TestCreateCoursePhase() {
	jsonData := `{"new_key": "new_value"}`
	var data meta.MetaData
	err := json.Unmarshal([]byte(jsonData), &data)
	assert.NoError(suite.T(), err)

	jsonData = `{"updated_key2": "updated_value"}`
	var studentData meta.MetaData
	err = json.Unmarshal([]byte(jsonData), &studentData)
	assert.NoError(suite.T(), err)

	newCoursePhase := coursePhaseDTO.CreateCoursePhase{
		CourseID:            uuid.MustParse("3f42d322-e5bf-4faa-b576-51f2cab14c2e"),
		Name:                "New Phase",
		IsInitialPhase:      false,
		RestrictedData:      data,
		StudentReadableData: studentData,
		CoursePhaseTypeID:   uuid.MustParse("7dc1c4e8-4255-4874-80a0-0c12b958744c"),
	}

	createdCoursePhase, err := CreateCoursePhase(suite.ctx, newCoursePhase)
	assert.NoError(suite.T(), err)

	// Verify creation
	fetchedCoursePhase, err := GetCoursePhaseByID(suite.ctx, createdCoursePhase.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "New Phase", fetchedCoursePhase.Name, "Expected course phase name to match")
	assert.False(suite.T(), fetchedCoursePhase.IsInitialPhase, "Expected course phase to not be an initial phase")
	assert.Equal(suite.T(), data, fetchedCoursePhase.RestrictedData, "Expected metadata to match")
	assert.Equal(suite.T(), studentData, fetchedCoursePhase.StudentReadableData, "Expected student readable data to match")
}

func (suite *CoursePhaseTestSuite) TestGetPrevPhaseDataByCoursePhaseID() {
	// Define UUIDs matching the ones in the dump.
	targetPhaseID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	predecessorPhaseID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	// Call the function under test.
	prevData, err := GetPrevPhaseDataByCoursePhaseID(suite.ctx, targetPhaseID)
	suite.NoError(err)

	// Verify that the core data (from the provided output DTO with endpoint 'core') is merged correctly.
	// Expecting: {"TestDTO": "restricted-test-value"} (since the predecessor phase's restricted_data contains that key)
	jsonData := `{"TestDTO": "restricted-test-value"}`
	var expectedCoreJSON meta.MetaData
	err = json.Unmarshal([]byte(jsonData), &expectedCoreJSON)
	assert.NoError(suite.T(), err)
	suite.Equal(expectedCoreJSON, prevData.PrevData)

	// Verify that resolution data (for non-core endpoint) is returned.
	// The resolution query should return one row with the non-core DTO.
	suite.Len(prevData.Resolutions, 1)
	resolution := prevData.Resolutions[0]
	suite.Equal("ResolutionDTO", resolution.DtoName)
	suite.Equal("non-core", resolution.EndpointPath)
	suite.Equal("http://example.com", resolution.BaseURL)
	suite.Equal(predecessorPhaseID.String(), resolution.CoursePhaseID.String())
}

func TestCoursePhaseTestSuite(t *testing.T) {
	suite.Run(t, new(CoursePhaseTestSuite))
}
