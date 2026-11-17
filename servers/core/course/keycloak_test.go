package course

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/course/courseDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// courseKeycloakSpy records the course identifiers the Keycloak hooks are called with.
type courseKeycloakSpy struct {
	created    []string
	cleanedUp  []string
	createErr  error
	cleanUpErr error
	onCreate   func(ctx context.Context, identifier string)
	onCleanUp  func(identifier string)
}

func (k *courseKeycloakSpy) create(ctx context.Context, courseName, iterationName, _ string) error {
	identifier := permissionValidation.CourseIdentifier(iterationName, courseName)
	k.created = append(k.created, identifier)
	if k.onCreate != nil {
		k.onCreate(ctx, identifier)
	}
	return k.createErr
}

func (k *courseKeycloakSpy) cleanUp(_ context.Context, courseName, iterationName string) error {
	identifier := permissionValidation.CourseIdentifier(iterationName, courseName)
	k.cleanedUp = append(k.cleanedUp, identifier)
	if k.onCleanUp != nil {
		k.onCleanUp(identifier)
	}
	return k.cleanUpErr
}

type CourseKeycloakTestSuite struct {
	suite.Suite
	ctx      context.Context
	cleanup  func()
	conn     *pgxpool.Pool
	queries  *db.Queries
	keycloak *courseKeycloakSpy
	service  *CourseService
}

func (suite *CourseKeycloakTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.ctx, "../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.conn = testDB.Conn
	suite.queries = testDB.Queries
	suite.keycloak = &courseKeycloakSpy{}
	suite.service = NewCourseService(*testDB.Queries, testDB.Conn, stubPhaseProvider{}, suite.keycloak.create, suite.keycloak.cleanUp,
		func(context.Context, uuid.UUID) error { return nil })
}

func (suite *CourseKeycloakTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *CourseKeycloakTestSuite) SetupTest() {
	*suite.keycloak = courseKeycloakSpy{}
}

func courseToCreate(name, semesterTag string) courseDTO.CreateCourse {
	return courseDTO.CreateCourse{
		Name:        name,
		StartDate:   pgtype.Date{Valid: true, Time: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)},
		EndDate:     pgtype.Date{Valid: true, Time: time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC)},
		SemesterTag: pgtype.Text{String: semesterTag, Valid: true},
		CourseType:  db.CourseType("practical course"),
	}
}

func (suite *CourseKeycloakTestSuite) countCourses(name, semesterTag string) int {
	var count int
	err := suite.conn.QueryRow(suite.ctx, "SELECT COUNT(*) FROM course WHERE name = $1 AND semester_tag = $2", name, semesterTag).Scan(&count)
	require.NoError(suite.T(), err)
	return count
}

// blocksOn reports whether another transaction has to wait for the given statement, which shows that
// the creating transaction still holds what the statement needs.
func (suite *CourseKeycloakTestSuite) blocksOn(statement func(tx *db.Queries) error) bool {
	tx, err := suite.conn.Begin(suite.ctx)
	require.NoError(suite.T(), err)
	defer func() { _ = tx.Rollback(suite.ctx) }()

	_, err = tx.Exec(suite.ctx, "SET LOCAL lock_timeout = '200ms'")
	require.NoError(suite.T(), err)
	err = statement(suite.queries.WithTx(tx))

	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "55P03"
}

func (suite *CourseKeycloakTestSuite) identifierLocked(identifier string) bool {
	return suite.blocksOn(func(tx *db.Queries) error { return tx.LockCourseIdentifier(suite.ctx, identifier) })
}

func (suite *CourseKeycloakTestSuite) TestCreateCourseProvisionsKeycloakUnderTheIdentifierLock() {
	identifierLocked, hadDeadline := false, false
	suite.keycloak.onCreate = func(ctx context.Context, identifier string) {
		identifierLocked = suite.identifierLocked(identifier)
		_, hadDeadline = ctx.Deadline()
	}

	_, err := suite.service.CreateCourse(suite.ctx, courseToCreate("provisioned", "ws24"), "creator")

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), []string{"ws24-provisioned"}, suite.keycloak.created)
	assert.Empty(suite.T(), suite.keycloak.cleanedUp)
	assert.True(suite.T(), identifierLocked, "a course with a colliding identifier must wait until this one is committed")
	assert.True(suite.T(), hadDeadline, "the Keycloak calls must be bounded")
	assert.Equal(suite.T(), 1, suite.countCourses("provisioned", "ws24"))
}

func (suite *CourseKeycloakTestSuite) TestCreateCourseCleansUpKeycloakBeforeTheRollback() {
	suite.keycloak.createErr = errors.New("keycloak unavailable")
	identifierLocked := false
	suite.keycloak.onCleanUp = func(identifier string) {
		identifierLocked = suite.identifierLocked(identifier)
	}

	_, err := suite.service.CreateCourse(suite.ctx, courseToCreate("held", "ws24"), "creator")

	assert.ErrorIs(suite.T(), err, suite.keycloak.createErr)
	assert.Equal(suite.T(), []string{"ws24-held"}, suite.keycloak.cleanedUp)
	assert.True(suite.T(), identifierLocked, "the cleanup must run before the rollback releases the identifier")
	assert.Zero(suite.T(), suite.countCourses("held", "ws24"))
}

func (suite *CourseKeycloakTestSuite) TestCreateCourseReportsTheProvisioningErrorWhenTheCleanupFails() {
	suite.keycloak.createErr = errors.New("keycloak unavailable")
	suite.keycloak.cleanUpErr = errors.New("cleanup failed")

	_, err := suite.service.CreateCourse(suite.ctx, courseToCreate("broken", "ws24"), "creator")

	assert.ErrorIs(suite.T(), err, suite.keycloak.createErr)
	assert.Zero(suite.T(), suite.countCourses("broken", "ws24"))
}

func (suite *CourseKeycloakTestSuite) TestCreateCourseSucceedsOnRetryAfterAKeycloakFailure() {
	suite.keycloak.createErr = errors.New("keycloak unavailable")
	_, err := suite.service.CreateCourse(suite.ctx, courseToCreate("retried", "ws24"), "creator")
	require.Error(suite.T(), err)

	suite.keycloak.createErr = nil
	created, err := suite.service.CreateCourse(suite.ctx, courseToCreate("retried", "ws24"), "creator")

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "retried", created.Name)
	assert.Equal(suite.T(), []string{"ws24-retried", "ws24-retried"}, suite.keycloak.created)
	assert.Equal(suite.T(), 1, suite.countCourses("retried", "ws24"))
}

func (suite *CourseKeycloakTestSuite) TestCreateCourseRejectsADuplicateBeforeKeycloak() {
	_, err := suite.service.CreateCourse(suite.ctx, courseToCreate("duplicate", "ws24"), "creator")
	require.NoError(suite.T(), err)

	_, err = suite.service.CreateCourse(suite.ctx, courseToCreate("duplicate", "ws24"), "creator")

	assert.ErrorIs(suite.T(), err, ErrDuplicateCourseIdentifier)
	assert.Equal(suite.T(), []string{"ws24-duplicate"}, suite.keycloak.created)
	assert.Empty(suite.T(), suite.keycloak.cleanedUp)
}

// "ws24" with "ios-lab" and "ws24-ios" with "lab" are different courses in the database, but would
// share the Keycloak groups and roles named "ws24-ios-lab".
func (suite *CourseKeycloakTestSuite) TestCreateCourseRejectsAnIdentifierAnotherCourseHas() {
	_, err := suite.service.CreateCourse(suite.ctx, courseToCreate("ios-lab", "ws24"), "first-creator")
	require.NoError(suite.T(), err)

	_, err = suite.service.CreateCourse(suite.ctx, courseToCreate("lab", "ws24-ios"), "second-creator")

	assert.ErrorIs(suite.T(), err, ErrCourseIdentifierTaken)
	assert.Equal(suite.T(), []string{"ws24-ios-lab"}, suite.keycloak.created, "Keycloak must not be touched for the second course")
	assert.Empty(suite.T(), suite.keycloak.cleanedUp, "the first course's Keycloak groups must survive")
	assert.Zero(suite.T(), suite.countCourses("lab", "ws24-ios"))
}

func (suite *CourseKeycloakTestSuite) TestCreateCourseAnswersAnIdentifierAnotherCourseHasWithConflict() {
	_, err := suite.service.CreateCourse(suite.ctx, courseToCreate("router-lab", "ss25"), "first-creator")
	require.NoError(suite.T(), err)
	body, err := json.Marshal(courseToCreate("lab", "ss25-router"))
	require.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/api/courses/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	newAdminCourseRouter(suite.service).ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusConflict, resp.Code)
	assert.Contains(suite.T(), resp.Body.String(), "ss25-router-lab")
}

func TestCourseKeycloakTestSuite(t *testing.T) {
	suite.Run(t, new(CourseKeycloakTestSuite))
}
