package coursePhaseDeletion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	deletedCoursePhaseID   = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	retainedCoursePhaseID  = uuid.MustParse("5179d58a-d00d-4fa7-94a5-397bc69fab03")
	studentParticipationID = uuid.MustParse("99999999-9999-9999-9999-999999999991")
	tutorParticipationID   = uuid.MustParse("99999999-9999-9999-9999-999999999993")
	retainedTutorID        = uuid.MustParse("99999999-9999-9999-9999-999999999994")
)

type CoursePhaseDeletionServiceTestSuite struct {
	suite.Suite
	suiteCtx context.Context
	cleanup  func()
	conn     *pgxpool.Pool
	queries  *db.Queries
	service  *CoursePhaseDeletionService
}

func (suite *CoursePhaseDeletionServiceTestSuite) SetupTest() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/coursePhaseDeletion.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.conn = testDB.Conn
	suite.queries = testDB.Queries
	suite.service = NewCoursePhaseDeletionService(*testDB.Queries, testDB.Conn)
}

// countRows counts every row of a table. The generated queries for the child tables join their
// parent, so they cannot tell a cascaded delete apart from an orphaned row. Counting the raw table
// is what catches a migration that weakens one of the ON DELETE CASCADE constraints the handler
// relies on.
func (suite *CoursePhaseDeletionServiceTestSuite) countRows(table string) int64 {
	var count int64
	err := suite.conn.QueryRow(suite.suiteCtx, "SELECT count(*) FROM "+table).Scan(&count)
	suite.Require().NoError(err, "Failed to count rows of %s", table)
	return count
}

func (suite *CoursePhaseDeletionServiceTestSuite) TearDownTest() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

// testContext builds the gin context the SDK handler signature expects, backed by the suite context.
func (suite *CoursePhaseDeletionServiceTestSuite) testContext() *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/delete", nil).WithContext(suite.suiteCtx)
	return c
}

func (suite *CoursePhaseDeletionServiceTestSuite) TestHandleCoursePhaseDeletion() {
	t := suite.T()

	// The seeded course phase holds data before the deletion.
	teams, err := suite.queries.GetTeamsByCoursePhase(suite.suiteCtx, deletedCoursePhaseID)
	assert.NoError(t, err)
	assert.NotEmpty(t, teams, "Expected seeded teams for the course phase under deletion")

	err = suite.service.HandleCoursePhaseDeletion(suite.testContext(), deletedCoursePhaseID)
	assert.NoError(t, err)

	teams, err = suite.queries.GetTeamsByCoursePhase(suite.suiteCtx, deletedCoursePhaseID)
	assert.NoError(t, err)
	assert.Empty(t, teams, "Expected all teams of the course phase to be deleted")

	skills, err := suite.queries.GetSkillsByCoursePhase(suite.suiteCtx, deletedCoursePhaseID)
	assert.NoError(t, err)
	assert.Empty(t, skills, "Expected all skills of the course phase to be deleted")

	_, err = suite.queries.GetSurveyTimeframe(suite.suiteCtx, deletedCoursePhaseID)
	assert.ErrorIs(t, err, pgx.ErrNoRows, "Expected the survey timeframe to be deleted")

	_, err = suite.queries.GetTeaseWorkspace(suite.suiteCtx, deletedCoursePhaseID)
	assert.ErrorIs(t, err, pgx.ErrNoRows, "Expected the tease workspace to be deleted")

	// The remaining tables are emptied by the ON DELETE CASCADE constraints.
	allocations, err := suite.queries.GetAllocationsByCoursePhase(suite.suiteCtx, deletedCoursePhaseID)
	assert.NoError(t, err)
	assert.Empty(t, allocations, "Expected allocations to be cascaded away with the teams")

	// student_team_preference_response and student_skill_response carry no course phase of their
	// own, so the rows left over are exactly the ones of the retained course phase. Counting the
	// raw tables also fails if a cascade stops firing and leaves the rows orphaned.
	assert.EqualValues(t, 1, suite.countRows("student_team_preference_response"),
		"Expected only the retained course phase to keep a team preference response")
	assert.EqualValues(t, 1, suite.countRows("student_skill_response"),
		"Expected only the retained course phase to keep a skill response")

	_, err = suite.queries.GetTutorByCourseParticipationID(suite.suiteCtx, db.GetTutorByCourseParticipationIDParams{
		CourseParticipationID: tutorParticipationID,
		CoursePhaseID:         deletedCoursePhaseID,
	})
	assert.ErrorIs(t, err, pgx.ErrNoRows, "Expected tutors to be cascaded away with the teams")
}

func (suite *CoursePhaseDeletionServiceTestSuite) TestHandleCoursePhaseDeletionKeepsOtherCoursePhases() {
	t := suite.T()

	err := suite.service.HandleCoursePhaseDeletion(suite.testContext(), deletedCoursePhaseID)
	assert.NoError(t, err)

	teams, err := suite.queries.GetTeamsByCoursePhase(suite.suiteCtx, retainedCoursePhaseID)
	assert.NoError(t, err)
	assert.Len(t, teams, 1, "Expected the other course phase to keep its team")

	skills, err := suite.queries.GetSkillsByCoursePhase(suite.suiteCtx, retainedCoursePhaseID)
	assert.NoError(t, err)
	assert.Len(t, skills, 1, "Expected the other course phase to keep its skill")

	_, err = suite.queries.GetSurveyTimeframe(suite.suiteCtx, retainedCoursePhaseID)
	assert.NoError(t, err, "Expected the other course phase to keep its survey timeframe")

	_, err = suite.queries.GetTeaseWorkspace(suite.suiteCtx, retainedCoursePhaseID)
	assert.NoError(t, err, "Expected the other course phase to keep its tease workspace")

	allocations, err := suite.queries.GetAllocationsByCoursePhase(suite.suiteCtx, retainedCoursePhaseID)
	assert.NoError(t, err)
	assert.Len(t, allocations, 1, "Expected the other course phase to keep its allocation")

	preferences, err := suite.queries.GetStudentTeamPreferences(suite.suiteCtx, db.GetStudentTeamPreferencesParams{
		CourseParticipationID: studentParticipationID,
		CoursePhaseID:         retainedCoursePhaseID,
	})
	assert.NoError(t, err)
	assert.Len(t, preferences, 1, "Expected the other course phase to keep its team preference response")

	skillResponses, err := suite.queries.GetStudentSkillResponses(suite.suiteCtx, db.GetStudentSkillResponsesParams{
		CourseParticipationID: studentParticipationID,
		CoursePhaseID:         retainedCoursePhaseID,
	})
	assert.NoError(t, err)
	assert.Len(t, skillResponses, 1, "Expected the other course phase to keep its skill response")

	_, err = suite.queries.GetTutorByCourseParticipationID(suite.suiteCtx, db.GetTutorByCourseParticipationIDParams{
		CourseParticipationID: retainedTutorID,
		CoursePhaseID:         retainedCoursePhaseID,
	})
	assert.NoError(t, err, "Expected the other course phase to keep its tutor")
}

func (suite *CoursePhaseDeletionServiceTestSuite) TestHandleCoursePhaseDeletionIsIdempotent() {
	t := suite.T()

	assert.NoError(t, suite.service.HandleCoursePhaseDeletion(suite.testContext(), deletedCoursePhaseID))
	assert.NoError(t, suite.service.HandleCoursePhaseDeletion(suite.testContext(), deletedCoursePhaseID),
		"Expected repeating the deletion to succeed")
}

func (suite *CoursePhaseDeletionServiceTestSuite) TestHandleCoursePhaseDeletionWithoutStoredData() {
	assert.NoError(suite.T(), suite.service.HandleCoursePhaseDeletion(suite.testContext(), uuid.New()),
		"Expected deleting a course phase without stored data to succeed")
}

func TestCoursePhaseDeletionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CoursePhaseDeletionServiceTestSuite))
}
