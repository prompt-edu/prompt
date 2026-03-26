package courseParticipation

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/course/courseParticipation/courseParticipationDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CourseParticipationTestSuite struct {
	suite.Suite
	ctx                        context.Context
	cleanup                    func()
	courseParticipationService CourseParticipationService
}

func (suite *CourseParticipationTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Set up PostgreSQL container
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../../database_dumps/course_participation_test.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}

	suite.cleanup = cleanup
	suite.courseParticipationService = CourseParticipationService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	CourseParticipationServiceSingleton = &suite.courseParticipationService
}

func (suite *CourseParticipationTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *CourseParticipationTestSuite) TestGetAllCourseParticipationsForCourse() {
	courseID := uuid.MustParse("3f42d322-e5bf-4faa-b576-51f2cab14c2e")
	participations, err := GetAllCourseParticipationsForCourse(suite.ctx, courseID)

	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(participations), 0, "Expected participations for the course")

	expectedStudentIDs := []uuid.UUID{
		uuid.MustParse("3d1f3b00-87f3-433b-a713-178c4050411a"),
		uuid.MustParse("7dc1c4e8-4255-4874-80a0-0c12b958744b"),
		uuid.MustParse("500db7ed-2eb2-42d0-82b3-8750e12afa8b"),
	}

	for i, participation := range participations {
		assert.Equal(suite.T(), courseID, participation.CourseID, "Expected CourseID to match")
		assert.Equal(suite.T(), expectedStudentIDs[i], participation.StudentID, "Expected StudentID to match")
	}
}

func (suite *CourseParticipationTestSuite) TestGetAllCourseParticipationsForStudent() {
	studentID := uuid.MustParse("7dc1c4e8-4255-4874-80a0-0c12b958744b")
	participations, err := GetAllCourseParticipationsForStudent(suite.ctx, studentID)

	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(participations), 0, "Expected participations for the student")

	expectedCourseIDs := []uuid.UUID{
		uuid.MustParse("3f42d322-e5bf-4faa-b576-51f2cab14c2e"),
		uuid.MustParse("918977e1-2d27-4b55-9064-8504ff027a1a"),
	}

	for i, participation := range participations {
		assert.Equal(suite.T(), studentID, participation.StudentID, "Expected StudentID to match")
		assert.Equal(suite.T(), expectedCourseIDs[i], participation.CourseID, "Expected CourseID to match")
	}
}

func (suite *CourseParticipationTestSuite) TestCreateCourseParticipation() {
	newParticipation := courseParticipationDTO.CreateCourseParticipation{
		CourseID:  uuid.MustParse("918977e1-2d27-4b55-9064-8504ff027a1a"),
		StudentID: uuid.MustParse("3d1f3b00-87f3-433b-a713-178c4050411a"),
	}

	createdParticipation, err := CreateCourseParticipation(suite.ctx, nil, newParticipation)
	assert.NoError(suite.T(), err)

	// Verify the created participation
	assert.Equal(suite.T(), newParticipation.CourseID, createdParticipation.CourseID, "Expected CourseID to match")
	assert.Equal(suite.T(), newParticipation.StudentID, createdParticipation.StudentID, "Expected StudentID to match")
	assert.NotEqual(suite.T(), uuid.Nil, createdParticipation.ID, "Expected a valid UUID for the new participation")
}

func TestCourseParticipationTestSuite(t *testing.T) {
	suite.Run(t, new(CourseParticipationTestSuite))
}
