package coursePhaseAuth

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type CoursePhaseAuthTestSuite struct {
	suite.Suite
	ctx                    context.Context
	cleanup                func()
	coursePhaseAuthService CoursePhaseAuthService
}

func (suite *CoursePhaseAuthTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Set up PostgreSQL container with test dump
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/full_db.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}

	suite.cleanup = cleanup
	suite.coursePhaseAuthService = CoursePhaseAuthService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	CoursePhaseAuthServiceSingleton = &suite.coursePhaseAuthService
}

func (suite *CoursePhaseAuthTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *CoursePhaseAuthTestSuite) TestCoursePhaseParticipation() {
	// Use known values from the test dump.
	// Adjust these UUIDs and credentials to match your test data.
	courseParticipationID := uuid.MustParse("1378db54-4ab4-4225-ac67-68e4345f21e2")
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	matriculationNumber := "09999999"
	universityLogin := "as45fgh"

	result, err := GetCoursePhaseParticipation(suite.ctx, coursePhaseID, matriculationNumber, universityLogin)
	suite.Require().NoError(err)
	suite.True(result.IsStudentOfCoursePhase, "Expected student to be part of the course phase")
	suite.Equal(courseParticipationID, result.CourseParticipationID, "Unexpected course participation id")
}

func (suite *CoursePhaseAuthTestSuite) TestGetCourseRoles() {
	// Use a coursePhaseID from your test data.
	coursePhaseID := uuid.MustParse("4e736d05-c125-48f0-8fa0-848b03ca6908")

	result, err := GetCourseRoles(suite.ctx, coursePhaseID)
	suite.Require().NoError(err)

	// Adjust expected values according to your test database contents.
	expectedLecturerRole := "ios2425-TestCourse-Lecturer"
	expectedEditorRole := "ios2425-TestCourse-Editor"
	expectedCustomRolePrefix := "ios2425-TestCourse-cg-"

	suite.Equal(expectedLecturerRole, result.CourseLecturerRole, "Unexpected lecturer role")
	suite.Equal(expectedEditorRole, result.CourseEditorRole, "Unexpected editor role")
	suite.Equal(expectedCustomRolePrefix, result.CustomRolePrefix, "Unexpected custom role prefix")
}

func TestCoursePhaseAuthTestSuite(t *testing.T) {
	suite.Run(t, new(CoursePhaseAuthTestSuite))
}
