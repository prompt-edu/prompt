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
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	// Seeded in database_dumps/certificate.sql with a template and two downloads.
	deletedCoursePhaseID = uuid.MustParse("10000000-0000-0000-0000-000000000001")
	// Seeded with a config row; the suite records a download for it before each deletion.
	retainedCoursePhaseID = uuid.MustParse("10000000-0000-0000-0000-000000000002")
)

type CoursePhaseDeletionServiceTestSuite struct {
	suite.Suite
	suiteCtx context.Context
	cleanup  func()
	queries  *db.Queries
	service  *CoursePhaseDeletionService
}

func (suite *CoursePhaseDeletionServiceTestSuite) SetupTest() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/certificate.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.queries = testDB.Queries
	suite.service = NewCoursePhaseDeletionService(*testDB.Queries, testDB.Conn)

	_, err = suite.queries.RecordCertificateDownload(suite.suiteCtx, db.RecordCertificateDownloadParams{
		StudentID:     uuid.New(),
		CoursePhaseID: retainedCoursePhaseID,
	})
	suite.Require().NoError(err)
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

func (suite *CoursePhaseDeletionServiceTestSuite) TestHandleCoursePhaseDeletion() {
	t := suite.T()

	// The seeded course phase holds data before the deletion.
	downloads, err := suite.queries.ListCertificateDownloadsByCoursePhase(suite.suiteCtx, deletedCoursePhaseID)
	assert.NoError(t, err)
	assert.NotEmpty(t, downloads, "Expected seeded downloads for the course phase under deletion")

	err = suite.service.HandleCoursePhaseDeletion(suite.testContext(), deletedCoursePhaseID)
	assert.NoError(t, err)

	downloads, err = suite.queries.ListCertificateDownloadsByCoursePhase(suite.suiteCtx, deletedCoursePhaseID)
	assert.NoError(t, err)
	assert.Empty(t, downloads, "Expected all downloads of the course phase to be deleted")

	_, err = suite.queries.GetCoursePhaseConfig(suite.suiteCtx, deletedCoursePhaseID)
	assert.ErrorIs(t, err, pgx.ErrNoRows, "Expected the config with its template to be deleted")
}

func (suite *CoursePhaseDeletionServiceTestSuite) TestHandleCoursePhaseDeletionKeepsOtherCoursePhases() {
	t := suite.T()

	err := suite.service.HandleCoursePhaseDeletion(suite.testContext(), deletedCoursePhaseID)
	assert.NoError(t, err)

	_, err = suite.queries.GetCoursePhaseConfig(suite.suiteCtx, retainedCoursePhaseID)
	assert.NoError(t, err, "Expected the other course phase to keep its config")

	downloads, err := suite.queries.ListCertificateDownloadsByCoursePhase(suite.suiteCtx, retainedCoursePhaseID)
	assert.NoError(t, err)
	assert.Len(t, downloads, 1, "Expected the other course phase to keep its download")
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
