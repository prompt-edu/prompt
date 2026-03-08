package developerProfile

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
	"github.com/prompt-edu/prompt/servers/intro_course/developerProfile/developerProfileDTO"
	"github.com/prompt-edu/prompt/servers/intro_course/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DeveloperProfileServiceTestSuite struct {
	suite.Suite
	ctx           context.Context
	cleanup       func()
	coursePhaseID uuid.UUID
	studentID     uuid.UUID
	service       DeveloperProfileService
}

func (suite *DeveloperProfileServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(suite.ctx, "../database_dumps/intro_course.sql")
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.coursePhaseID = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	suite.studentID = uuid.MustParse("33333333-3333-3333-3333-333333333333")

	suite.service = DeveloperProfileService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	DeveloperProfileServiceSingleton = &suite.service
}

func (suite *DeveloperProfileServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func TestDeveloperProfileServiceTestSuite(t *testing.T) {
	suite.Run(t, new(DeveloperProfileServiceTestSuite))
}

func (suite *DeveloperProfileServiceTestSuite) TestCreateDeveloperProfile() {
	newCoursePhase := uuid.New()
	newStudent := uuid.New()
	request := developerProfileDTO.PostDeveloperProfile{
		AppleID:        "newstudent@apple.com",
		GitLabUsername: "newstudent",
		HasMacBook:     true,
		IPhoneUDID:     pgtype.Text{String: "AAAABBBB-CCCCDDDD-EEEEFFFF", Valid: true},
	}

	err := CreateDeveloperProfile(suite.ctx, newCoursePhase, newStudent, request)
	assert.NoError(suite.T(), err)

	profile, err := DeveloperProfileServiceSingleton.queries.GetDeveloperProfileByCourseParticipationID(suite.ctx, db.GetDeveloperProfileByCourseParticipationIDParams{
		CoursePhaseID:         newCoursePhase,
		CourseParticipationID: newStudent,
	})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), request.GitLabUsername, profile.GitlabUsername)
}

func (suite *DeveloperProfileServiceTestSuite) TestGetOwnDeveloperProfile() {
	profile, err := GetOwnDeveloperProfile(suite.ctx, suite.coursePhaseID, suite.studentID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "student1git", profile.GitLabUsername)
}

func (suite *DeveloperProfileServiceTestSuite) TestGetOwnDeveloperProfileNotFound() {
	profile, err := GetOwnDeveloperProfile(suite.ctx, suite.coursePhaseID, uuid.New())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), developerProfileDTO.DeveloperProfile{}, profile)
}

func (suite *DeveloperProfileServiceTestSuite) TestGetAllDeveloperProfiles() {
	profiles, err := GetAllDeveloperProfiles(suite.ctx, suite.coursePhaseID)

	assert.NoError(suite.T(), err)
	// At least 2 profiles from the database dump
	assert.GreaterOrEqual(suite.T(), len(profiles), 2)
}

func (suite *DeveloperProfileServiceTestSuite) TestCreateOrUpdateDeveloperProfile() {
	// Use a new student that doesn't exist yet to avoid interfering with other tests
	newStudentID := uuid.New()
	updateProfile := developerProfileDTO.DeveloperProfile{
		AppleID:        "updated@apple.com",
		GitLabUsername: "updateduser",
		HasMacBook:     false,
		IPhoneUDID:     pgtype.Text{String: "FFFFBBBB-CCCCDDDDEEEEAAAA", Valid: true},
	}

	err := CreateOrUpdateDeveloperProfile(suite.ctx, suite.coursePhaseID, newStudentID, updateProfile)
	assert.NoError(suite.T(), err)

	profile, err := GetOwnDeveloperProfile(suite.ctx, suite.coursePhaseID, newStudentID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), updateProfile.AppleID, profile.AppleID)
	assert.False(suite.T(), profile.HasMacBook)
}

func (suite *DeveloperProfileServiceTestSuite) TestGetDevicesForCoursePhase() {
	devices, err := GetDevicesForCoursePhase(suite.ctx, suite.coursePhaseID)

	assert.NoError(suite.T(), err)
	deviceMap := make(map[uuid.UUID][]string)
	for _, d := range devices {
		deviceMap[d.CourseParticipationID] = d.Devices
	}

	assert.ElementsMatch(suite.T(), []string{"Mac", "IPhone", "IPad", "Watch"}, deviceMap[suite.studentID])
	assert.ElementsMatch(suite.T(), []string{"Mac"}, deviceMap[uuid.MustParse("44444444-4444-4444-4444-444444444444")])
}

func (suite *DeveloperProfileServiceTestSuite) TestGetDevicesForCourseParticipation() {
	devices, err := GetDevicesForCourseParticipation(suite.ctx, suite.coursePhaseID, suite.studentID)

	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), devices, "Mac")
}

func (suite *DeveloperProfileServiceTestSuite) TestGetDevicesForCourseParticipationNotFound() {
	_, err := GetDevicesForCourseParticipation(suite.ctx, suite.coursePhaseID, uuid.New())

	assert.Error(suite.T(), err)
	assert.ErrorIs(suite.T(), err, sql.ErrNoRows)
}
