package coursePhaseDeletion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/stretchr/testify/suite"
)

// The ids below match database_dumps/coursePhaseDeletion.sql, which describes the scenario.
var (
	deletedCoursePhaseID  = uuid.MustParse("d1000000-0000-0000-0000-000000000000")
	retainedCoursePhaseID = uuid.MustParse("d2000000-0000-0000-0000-000000000000")
	earlierCoursePhaseID  = uuid.MustParse("d3000000-0000-0000-0000-000000000000")
	sharedSchemaID        = uuid.MustParse("d5000000-0000-0000-0000-000000000001")
	dataOnlySchemaID      = uuid.MustParse("d5000000-0000-0000-0000-000000000002")
	privateSchemaID       = uuid.MustParse("d5000000-0000-0000-0000-000000000003")
	unusedSchemaID        = uuid.MustParse("d5000000-0000-0000-0000-000000000004")
	retainedSchemaID      = uuid.MustParse("d5000000-0000-0000-0000-000000000005")
)

// coursePhaseTables lists every table that stores course phase scoped data. A new table keyed by
// course_phase_id belongs here and in HandleCoursePhaseDeletion.
var coursePhaseTables = []string{
	"assessment",
	"independent_assessment",
	"category_assessment",
	"assessment_completion",
	"evaluation",
	"evaluation_completion",
	"feedback_items",
	"action_item",
	"evaluation_reminder",
	"results_released_mail",
	"course_phase_config",
}

type CoursePhaseDeletionServiceTestSuite struct {
	suite.Suite
	suiteCtx context.Context
	cleanup  func()
	conn     *pgxpool.Pool
	service  *CoursePhaseDeletionService
}

func (suite *CoursePhaseDeletionServiceTestSuite) SetupTest() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.suiteCtx, "../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../database_dumps/coursePhaseDeletion.sql")
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.conn = testDB.Conn
	suite.service = NewCoursePhaseDeletionService(*testDB.Queries, testDB.Conn)
}

func (suite *CoursePhaseDeletionServiceTestSuite) TearDownTest() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

// testContext builds the gin context the SDK handler signature expects, backed by the suite context.
func (suite *CoursePhaseDeletionServiceTestSuite) testContext() *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodDelete, "/", nil).WithContext(suite.suiteCtx)
	return c
}

func (suite *CoursePhaseDeletionServiceTestSuite) deletePhase(coursePhaseID uuid.UUID) {
	suite.Require().NoError(suite.service.HandleCoursePhaseDeletion(suite.testContext(), coursePhaseID))
}

func (suite *CoursePhaseDeletionServiceTestSuite) countPhaseRows(table string, coursePhaseID uuid.UUID) int64 {
	var count int64
	err := suite.conn.QueryRow(suite.suiteCtx, "SELECT count(*) FROM "+table+" WHERE course_phase_id = $1", coursePhaseID).Scan(&count)
	suite.Require().NoError(err, "Failed to count rows of %s", table)
	return count
}

// schemaOwner returns the owning phase of a schema and whether the schema still exists.
func (suite *CoursePhaseDeletionServiceTestSuite) schemaOwner(schemaID uuid.UUID) (pgtype.UUID, bool) {
	var owner pgtype.UUID
	err := suite.conn.QueryRow(suite.suiteCtx, "SELECT source_phase_id FROM assessment_schema WHERE id = $1", schemaID).Scan(&owner)
	if err != nil {
		suite.Require().ErrorContains(err, "no rows", "Failed to read the owner of schema %s", schemaID)
		return pgtype.UUID{}, false
	}
	return owner, true
}

// countSchemaContent counts the categories and competencies of a schema.
func (suite *CoursePhaseDeletionServiceTestSuite) countSchemaContent(schemaID uuid.UUID) (categories, competencies int64) {
	err := suite.conn.QueryRow(suite.suiteCtx, `
		SELECT (SELECT count(*) FROM category WHERE assessment_schema_id = $1),
		       (SELECT count(*)
		        FROM competency co
		                 JOIN category cat ON cat.id = co.category_id
		        WHERE cat.assessment_schema_id = $1)`,
		schemaID).Scan(&categories, &competencies)
	suite.Require().NoError(err, "Failed to count the content of schema %s", schemaID)
	return categories, competencies
}

func (suite *CoursePhaseDeletionServiceTestSuite) requireOwner(schemaID, expectedOwner uuid.UUID) {
	owner, exists := suite.schemaOwner(schemaID)
	suite.Require().True(exists, "Expected schema %s to be kept", schemaID)
	suite.Require().Equal(pgtype.UUID{Bytes: expectedOwner, Valid: true}, owner, "Unexpected owner of schema %s", schemaID)
}

func (suite *CoursePhaseDeletionServiceTestSuite) TestDeletesEveryCoursePhaseScopedRow() {
	for _, table := range coursePhaseTables {
		suite.Require().EqualValues(1, suite.countPhaseRows(table, deletedCoursePhaseID), "Expected seeded rows in %s", table)
	}

	suite.deletePhase(deletedCoursePhaseID)

	for _, table := range coursePhaseTables {
		suite.Equal(int64(0), suite.countPhaseRows(table, deletedCoursePhaseID), "Expected %s to hold no rows of the deleted phase", table)
	}
}

func (suite *CoursePhaseDeletionServiceTestSuite) TestKeepsOtherCoursePhases() {
	suite.deletePhase(deletedCoursePhaseID)

	for _, table := range coursePhaseTables {
		suite.Equal(int64(1), suite.countPhaseRows(table, retainedCoursePhaseID), "Expected %s to keep the row of the retained phase", table)
	}
	suite.Equal(int64(1), suite.countPhaseRows("course_phase_config", earlierCoursePhaseID))
	suite.requireOwner(retainedSchemaID, retainedCoursePhaseID)
}

func (suite *CoursePhaseDeletionServiceTestSuite) TestDeletesOwnedSchemasNoOtherPhaseUses() {
	suite.deletePhase(deletedCoursePhaseID)

	for _, schemaID := range []uuid.UUID{privateSchemaID, unusedSchemaID} {
		_, exists := suite.schemaOwner(schemaID)
		suite.False(exists, "Expected schema %s to be deleted", schemaID)
		categories, competencies := suite.countSchemaContent(schemaID)
		suite.Zero(categories, "Expected the categories of schema %s to be deleted", schemaID)
		suite.Zero(competencies, "Expected the competencies of schema %s to be deleted", schemaID)
	}
}

func (suite *CoursePhaseDeletionServiceTestSuite) TestHandsSharedSchemaToTheEarliestPhaseUsingIt() {
	suite.deletePhase(deletedCoursePhaseID)

	suite.requireOwner(sharedSchemaID, earlierCoursePhaseID)
	categories, competencies := suite.countSchemaContent(sharedSchemaID)
	suite.EqualValues(1, categories, "Expected the shared schema to keep its category")
	suite.EqualValues(1, competencies, "Expected the shared schema to keep its competency")
}

func (suite *CoursePhaseDeletionServiceTestSuite) TestHandsSchemaToThePhaseHoldingDataInIt() {
	suite.deletePhase(deletedCoursePhaseID)

	suite.requireOwner(dataOnlySchemaID, retainedCoursePhaseID)
	suite.Equal(int64(1), suite.countPhaseRows("category_assessment", retainedCoursePhaseID),
		"Expected the retained phase to keep its category assessment in the handed over schema")
}

// The deleted and the retained phases select the global templates the migrations create.
func (suite *CoursePhaseDeletionServiceTestSuite) TestKeepsGlobalTemplates() {
	countTemplates := func() int64 {
		var count int64
		err := suite.conn.QueryRow(suite.suiteCtx, "SELECT count(*) FROM assessment_schema WHERE source_phase_id IS NULL").Scan(&count)
		suite.Require().NoError(err)
		return count
	}
	before := countTemplates()
	suite.Require().NotZero(before, "Expected the migrations to create global templates")

	suite.deletePhase(deletedCoursePhaseID)

	suite.Equal(before, countTemplates(), "Expected every global template to be kept and stay global")
}

// Deleting the new owner afterwards must clean up the schema it was handed, so no schema outlives
// every phase that used it.
func (suite *CoursePhaseDeletionServiceTestSuite) TestHandedOverSchemaFollowsItsNewOwner() {
	suite.deletePhase(deletedCoursePhaseID)
	suite.deletePhase(earlierCoursePhaseID)
	suite.requireOwner(sharedSchemaID, retainedCoursePhaseID)

	suite.deletePhase(retainedCoursePhaseID)
	for _, schemaID := range []uuid.UUID{sharedSchemaID, dataOnlySchemaID, retainedSchemaID} {
		_, exists := suite.schemaOwner(schemaID)
		suite.False(exists, "Expected schema %s to be deleted with its last owner", schemaID)
	}
}

func (suite *CoursePhaseDeletionServiceTestSuite) TestIsIdempotent() {
	suite.deletePhase(deletedCoursePhaseID)
	suite.NoError(suite.service.HandleCoursePhaseDeletion(suite.testContext(), deletedCoursePhaseID),
		"Expected repeating the deletion to succeed")
}

func (suite *CoursePhaseDeletionServiceTestSuite) TestSucceedsWithoutStoredData() {
	suite.NoError(suite.service.HandleCoursePhaseDeletion(suite.testContext(), uuid.New()),
		"Expected deleting a course phase without stored data to succeed")
}

func TestCoursePhaseDeletionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CoursePhaseDeletionServiceTestSuite))
}
