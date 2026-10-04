package coursePhaseDeletion

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt/servers/presentation/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var (
	deletedCoursePhaseID  = uuid.MustParse("11000000-0000-0000-0000-000000000001")
	retainedCoursePhaseID = uuid.MustParse("11000000-0000-0000-0000-000000000002")

	deletedMaterialKeys = []string{
		"presentations/11000000-0000-0000-0000-000000000001/41000000-0000-0000-0000-000000000001/61000000-0000-0000-0000-000000000001/slides.pdf",
		"presentations/11000000-0000-0000-0000-000000000001/41000000-0000-0000-0000-000000000001/61000000-0000-0000-0000-000000000002/draft.pdf",
		// An object without a row, as an upload leaves behind when its row was reclaimed first.
		"presentations/11000000-0000-0000-0000-000000000001/41000000-0000-0000-0000-000000000001/61000000-0000-0000-0000-000000000009/orphan.pdf",
	}
	retainedMaterialKey = "presentations/11000000-0000-0000-0000-000000000002/41000000-0000-0000-0000-000000000002/61000000-0000-0000-0000-000000000003/slides.pdf"
)

// Every table the service stores data in. The fixture seeds exactly one row per table for the
// retained course phase.
var presentationTables = []string{
	"course_phase_config",
	"feedback_category",
	"presentation_slot",
	"presentation",
	"presentation_material",
	"feedback_form",
	"feedback_answer",
	"feedback_contributor",
}

type CoursePhaseDeletionServiceTestSuite struct {
	suite.Suite
	suiteCtx context.Context
	cleanup  func()
	conn     *pgxpool.Pool
	storage  *testutils.FakeStorage
	service  *CoursePhaseDeletionService
}

func (suite *CoursePhaseDeletionServiceTestSuite) SetupTest() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(suite.suiteCtx, "../database_dumps/coursePhaseDeletion.sql")
	suite.Require().NoError(err, "Failed to set up test database")
	suite.cleanup = cleanup
	suite.conn = testDB.Conn

	suite.storage = testutils.NewFakeStorage()
	for _, key := range append([]string{retainedMaterialKey}, deletedMaterialKeys...) {
		suite.storage.Put(key, "application/pdf", 1024)
	}
	suite.service = NewCoursePhaseDeletionService(testDB.Queries, testDB.Conn, suite.storage)
}

func (suite *CoursePhaseDeletionServiceTestSuite) TearDownTest() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

// countRows counts the whole table rather than the rows of one phase: feedback_answer,
// feedback_contributor, feedback_form and presentation_material carry no course phase of their
// own, so a raw count is what catches a migration that weakens one of the ON DELETE CASCADE
// constraints the handler relies on.
func (suite *CoursePhaseDeletionServiceTestSuite) countRows(table string) int64 {
	var count int64
	err := suite.conn.QueryRow(suite.suiteCtx, "SELECT count(*) FROM "+table).Scan(&count)
	suite.Require().NoError(err, "Failed to count rows of %s", table)
	return count
}

func (suite *CoursePhaseDeletionServiceTestSuite) countPhaseRows(table string, coursePhaseID uuid.UUID) int64 {
	var count int64
	err := suite.conn.QueryRow(suite.suiteCtx,
		"SELECT count(*) FROM "+table+" WHERE course_phase_id = $1", coursePhaseID).Scan(&count)
	suite.Require().NoError(err, "Failed to count rows of %s", table)
	return count
}

// testContext builds the gin context the SDK handler signature expects, backed by the suite context.
func (suite *CoursePhaseDeletionServiceTestSuite) testContext() *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodDelete, "/", nil).WithContext(suite.suiteCtx)
	return c
}

func (suite *CoursePhaseDeletionServiceTestSuite) TestHandleCoursePhaseDeletion() {
	t := suite.T()

	require.NoError(t, suite.service.HandleCoursePhaseDeletion(suite.testContext(), deletedCoursePhaseID))

	for _, table := range presentationTables {
		assert.EqualValues(t, 1, suite.countRows(table),
			"Expected only the retained course phase to keep its row in %s", table)
	}
	for _, key := range deletedMaterialKeys {
		assert.False(t, suite.storage.Has(key), "Expected the stored material %s to be deleted", key)
	}
}

func (suite *CoursePhaseDeletionServiceTestSuite) TestHandleCoursePhaseDeletionKeepsOtherCoursePhases() {
	t := suite.T()

	require.NoError(t, suite.service.HandleCoursePhaseDeletion(suite.testContext(), deletedCoursePhaseID))

	for _, table := range []string{"course_phase_config", "feedback_category", "presentation_slot", "presentation"} {
		assert.EqualValues(t, 1, suite.countPhaseRows(table, retainedCoursePhaseID),
			"Expected the other course phase to keep its row in %s", table)
	}
	assert.True(t, suite.storage.Has(retainedMaterialKey), "Expected the other course phase to keep its stored material")
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

// A storage failure after the commit has to fail the request, so core keeps its rows and retries,
// and the retry has to reach the objects although their rows are already gone.
func (suite *CoursePhaseDeletionServiceTestSuite) TestHandleCoursePhaseDeletionRetriesFailedStorageCleanup() {
	t := suite.T()

	suite.storage.DeletePrefixErr = errors.New("storage unavailable")
	assert.Error(t, suite.service.HandleCoursePhaseDeletion(suite.testContext(), deletedCoursePhaseID),
		"Expected a failed storage cleanup to fail the deletion")
	assert.Zero(t, suite.countPhaseRows("presentation", deletedCoursePhaseID),
		"Expected the database deletion to be committed before the storage cleanup")
	for _, key := range deletedMaterialKeys {
		assert.True(t, suite.storage.Has(key), "Expected %s to survive the failed cleanup", key)
	}

	suite.storage.DeletePrefixErr = nil
	require.NoError(t, suite.service.HandleCoursePhaseDeletion(suite.testContext(), deletedCoursePhaseID))
	for _, key := range deletedMaterialKeys {
		assert.False(t, suite.storage.Has(key), "Expected the retry to delete %s", key)
	}
	assert.True(t, suite.storage.Has(retainedMaterialKey), "Expected the other course phase to keep its stored material")
}

func TestCoursePhaseDeletionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CoursePhaseDeletionServiceTestSuite))
}
