package coursePhaseParticipation

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseParticipation/coursePhaseParticipationDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CoursePhaseParticipationTestSuite struct {
	suite.Suite
	ctx                             context.Context
	cleanup                         func()
	coursePhaseParticipationService CoursePhaseParticipationService
}

func (suite *CoursePhaseParticipationTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Set up PostgreSQL container
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../../database_dumps/full_db.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}

	suite.cleanup = cleanup
	suite.coursePhaseParticipationService = CoursePhaseParticipationService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	CoursePhaseParticipationServiceSingleton = &suite.coursePhaseParticipationService

	resolution.InitResolutionModule("localhost:8080")
}

func (suite *CoursePhaseParticipationTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *CoursePhaseParticipationTestSuite) TestGetAllParticipationsForCoursePhase() {
	coursePhaseID := uuid.MustParse("4e736d05-c125-48f0-8fa0-848b03ca6908")

	participationsWithResolution, err := GetAllParticipationsForCoursePhase(suite.ctx, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(participationsWithResolution.Participations), 0, "Expected participations for the course phase")

	for _, participation := range participationsWithResolution.Participations {
		log.Info(participation.PrevData)
	}
}

func (suite *CoursePhaseParticipationTestSuite) TestGetParticipationsWithPrevData() {
	coursePhaseID := uuid.MustParse("2b1a55ad-8b1d-453f-b2b4-2373ecb35bc1")

	participationsWithResolution, err := GetAllParticipationsForCoursePhase(suite.ctx, coursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(participationsWithResolution.Participations), 0, "Expected participations for the course phase")
	for _, participation := range participationsWithResolution.Participations {
		assert.NotNil(suite.T(), participation.PrevData, "Expected prev data to be present")
		assert.NotNil(suite.T(), participation.PrevData["score"], "Expected score to be present")
	}
}

func (suite *CoursePhaseParticipationTestSuite) TestCreateCoursePhaseParticipation() {
	jsonData := `{"a": "b"}`
	var data meta.MetaData
	err := json.Unmarshal([]byte(jsonData), &data)
	assert.NoError(suite.T(), err)
	pass := db.PassStatusPassed

	jsonData = `{"skills2": "more than none", "other-value2": "some skills"}`
	var studentReadableData meta.MetaData
	err = json.Unmarshal([]byte(jsonData), &studentReadableData)
	assert.NoError(suite.T(), err)

	newParticipation := coursePhaseParticipationDTO.CreateCoursePhaseParticipation{
		CoursePhaseID:         uuid.MustParse("4e736d05-c125-48f0-8fa0-848b03ca6908"),
		CourseParticipationID: uuid.MustParse("ca41772a-e06d-40eb-9c4b-ab44e06a890c"),
		PassStatus:            &pass,
		RestrictedData:        data,
		StudentReadableData:   studentReadableData,
	}

	createdParticipation, err := CreateOrUpdateCoursePhaseParticipation(suite.ctx, nil, newParticipation)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newParticipation.CoursePhaseID, createdParticipation.CoursePhaseID, "CoursePhaseID should match")
	assert.Equal(suite.T(), newParticipation.RestrictedData, createdParticipation.RestrictedData, "Meta data should match")
	assert.Equal(suite.T(), newParticipation.StudentReadableData, createdParticipation.StudentReadableData, "Meta data should match")
	assert.Equal(suite.T(), "passed", createdParticipation.PassStatus, "PassStatus should match")
}

func (suite *CoursePhaseParticipationTestSuite) TestUpdateCoursePhaseParticipation() {
	courseParticipationID := uuid.MustParse("6a49b717-a8ca-4d16-bcd0-0bb059525269")
	coursePhaseID := uuid.MustParse("4e736d05-c125-48f0-8fa0-848b03ca6908")
	jsonData := `{"other-value": "some skills"}`
	var data meta.MetaData
	err := json.Unmarshal([]byte(jsonData), &data)
	assert.NoError(suite.T(), err)
	pass := db.PassStatusPassed

	jsonData = `{"skills2": "more than none", "other-value2": "some skills"}`
	var studentReadableData meta.MetaData
	err = json.Unmarshal([]byte(jsonData), &studentReadableData)
	assert.NoError(suite.T(), err)

	updatedParticipation := coursePhaseParticipationDTO.UpdateCoursePhaseParticipation{
		CourseParticipationID: courseParticipationID,
		RestrictedData:        data,
		StudentReadableData:   studentReadableData,
		PassStatus:            &pass,
		CoursePhaseID:         coursePhaseID,
	}

	err = UpdateCoursePhaseParticipation(suite.ctx, nil, updatedParticipation)
	assert.NoError(suite.T(), err)
	result, err := GetCoursePhaseParticipation(suite.ctx, coursePhaseID, courseParticipationID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), updatedParticipation.CourseParticipationID, result.Participation.CourseParticipationID, "Participation ID should match")
	assert.Equal(suite.T(), "passed", result.Participation.PassStatus, "PassStatus should match")
	assert.Equal(suite.T(), updatedParticipation.RestrictedData["other-value"], result.Participation.RestrictedData["other-value"], "New Meta data should match")
	assert.Equal(suite.T(), updatedParticipation.StudentReadableData["other-value"], result.Participation.StudentReadableData["other-value"], "New Meta data should match")

}

func (suite *CoursePhaseParticipationTestSuite) TestUpdateCoursePhaseParticipationWithMetaDataOverride() {
	courseParticipationID := uuid.MustParse("6a49b717-a8ca-4d16-bcd0-0bb059525269")
	coursePhaseID := uuid.MustParse("4e736d05-c125-48f0-8fa0-848b03ca6908")
	jsonData := `{"skills": "more than none", "other-value": "some skills"}`
	var data meta.MetaData
	err := json.Unmarshal([]byte(jsonData), &data)
	assert.NoError(suite.T(), err)

	jsonData = `{"skills2": "more than none", "other-value2": "some skills"}`
	var studentReadableData meta.MetaData
	err = json.Unmarshal([]byte(jsonData), &studentReadableData)
	assert.NoError(suite.T(), err)

	updatedParticipation := coursePhaseParticipationDTO.UpdateCoursePhaseParticipation{
		CourseParticipationID: courseParticipationID,
		RestrictedData:        data,
		StudentReadableData:   studentReadableData,
		PassStatus:            nil, // Updated to use the ENUM value
		CoursePhaseID:         uuid.MustParse("4e736d05-c125-48f0-8fa0-848b03ca6908"),
	}

	BeforeResult, err := GetCoursePhaseParticipation(suite.ctx, coursePhaseID, courseParticipationID)
	assert.NoError(suite.T(), err)

	err = UpdateCoursePhaseParticipation(suite.ctx, nil, updatedParticipation)
	assert.NoError(suite.T(), err)
	result, err := GetCoursePhaseParticipation(suite.ctx, coursePhaseID, courseParticipationID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), updatedParticipation.CourseParticipationID, result.Participation.CourseParticipationID, "Participation ID should match")
	assert.Equal(suite.T(), BeforeResult.Participation.PassStatus, result.Participation.PassStatus, "PassStatus should match")
	for key, value := range updatedParticipation.RestrictedData {
		assert.Equal(suite.T(), result.Participation.RestrictedData[key], value, "Updated Meta data should be stored")
	}
	for key, value := range updatedParticipation.StudentReadableData {
		assert.Equal(suite.T(), result.Participation.StudentReadableData[key], value, "Updated Meta data should be stored")
	}
}

func (suite *CoursePhaseParticipationTestSuite) TestNewCoursePhaseParticipation() {
	courseParticipation := uuid.MustParse("f6744410-cfe2-456d-96fa-e857cf989569")
	jsonData := `{"skills": "more than none", "other-value": "some skills"}`
	var data meta.MetaData
	err := json.Unmarshal([]byte(jsonData), &data)
	assert.NoError(suite.T(), err)

	jsonData = `{"skills2": "more than none", "other-value2": "some skills"}`
	var studentReadableData meta.MetaData
	err = json.Unmarshal([]byte(jsonData), &studentReadableData)
	assert.NoError(suite.T(), err)

	createParticipation := coursePhaseParticipationDTO.CreateCoursePhaseParticipation{
		CourseParticipationID: courseParticipation,
		RestrictedData:        data,
		StudentReadableData:   studentReadableData,
		PassStatus:            nil, // Updated to use the ENUM value
		CoursePhaseID:         uuid.MustParse("4e736d05-c125-48f0-8fa0-848b03ca6908"),
	}

	participation, err := CreateOrUpdateCoursePhaseParticipation(suite.ctx, nil, createParticipation)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), participation.PassStatus, "not_assessed", "PassStatus should be not assessed")
	for key, value := range createParticipation.RestrictedData {
		assert.Equal(suite.T(), participation.RestrictedData[key], value, "Updated Meta data should be stored")
	}
	for key, value := range createParticipation.StudentReadableData {
		assert.Equal(suite.T(), participation.StudentReadableData[key], value, "Updated Meta data should be stored")
	}
}

func (suite *CoursePhaseParticipationTestSuite) TestNewCoursePhaseParticipationWrongCoursePhase() {
	courseParticipation := uuid.MustParse("f6744410-cfe2-456d-96fa-e857cf989569")
	jsonData := `{"skills": "more than none", "other-value": "some skills"}`
	var restrictedData meta.MetaData
	err := json.Unmarshal([]byte(jsonData), &restrictedData)
	assert.NoError(suite.T(), err)

	jsonData = `{"skills2": "more than none", "other-value2": "some skills"}`
	var studentReadableData meta.MetaData
	err = json.Unmarshal([]byte(jsonData), &studentReadableData)
	assert.NoError(suite.T(), err)

	createParticipation := coursePhaseParticipationDTO.CreateCoursePhaseParticipation{
		CourseParticipationID: courseParticipation,
		RestrictedData:        restrictedData,
		StudentReadableData:   studentReadableData,
		PassStatus:            nil,
		CoursePhaseID:         uuid.MustParse("7062236a-e290-487c-be41-29b24e0afc64"), // belongs to wrong course
	}

	_, err = CreateOrUpdateCoursePhaseParticipation(suite.ctx, nil, createParticipation)
	assert.Error(suite.T(), err)
}

func TestCoursePhaseParticipationTestSuite(t *testing.T) {
	suite.Run(t, new(CoursePhaseParticipationTestSuite))
}
