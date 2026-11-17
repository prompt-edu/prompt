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

// slotHeld reports whether inserting another course with the same name and semester tag has to wait,
// which shows that the creating transaction still holds the course's unique slot.
func (suite *CourseKeycloakTestSuite) slotHeld(name, semesterTag string) bool {
	tx, err := suite.conn.Begin(suite.ctx)
	require.NoError(suite.T(), err)
	defer func() { _ = tx.Rollback(suite.ctx) }()

	_, err = tx.Exec(suite.ctx, "SET LOCAL lock_timeout = '200ms'")
	require.NoError(suite.T(), err)
	_, err = tx.Exec(suite.ctx, "INSERT INTO course (id, name, semester_tag, course_type) VALUES ($1, $2, $3, 'practical course')",
		uuid.New(), name, semesterTag)

	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "55P03"
}

func (suite *CourseKeycloakTestSuite) TestCreateCourseProvisionsKeycloakWhileTheCourseSlotIsHeld() {
	slotHeld, hadDeadline := false, false
	suite.keycloak.onCreate = func(ctx context.Context, _ string) {
		slotHeld = suite.slotHeld("provisioned", "ws24")
		_, hadDeadline = ctx.Deadline()
	}

	_, err := suite.service.CreateCourse(suite.ctx, courseToCreate("provisioned", "ws24"), "creator")

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), []string{"ws24-provisioned"}, suite.keycloak.created)
	assert.Empty(suite.T(), suite.keycloak.cleanedUp)
	assert.True(suite.T(), slotHeld, "a second creation of the same course must wait until this one is committed")
	assert.True(suite.T(), hadDeadline, "the Keycloak calls must be bounded")
	assert.Equal(suite.T(), 1, suite.countCourses("provisioned", "ws24"))
}

func (suite *CourseKeycloakTestSuite) TestCreateCourseCleansUpKeycloakBeforeTheRollback() {
	suite.keycloak.createErr = errors.New("keycloak unavailable")
	slotHeld := false
	suite.keycloak.onCleanUp = func(string) {
		slotHeld = suite.slotHeld("held", "ws24")
	}

	_, err := suite.service.CreateCourse(suite.ctx, courseToCreate("held", "ws24"), "creator")

	assert.ErrorIs(suite.T(), err, suite.keycloak.createErr)
	assert.Equal(suite.T(), []string{"ws24-held"}, suite.keycloak.cleanedUp)
	assert.True(suite.T(), slotHeld, "the cleanup must run before the rollback releases the course's slot")
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

// The router validates the name and semester tag first. The service refuses them again before
// Keycloak, because the provisioning relies on them:
//   - "ws24" with "ios-lab" and "ws24-ios" with "lab" would share the groups and roles "ws24-ios-lab".
//   - Keycloak resolves "Prompt/ws24-ios/Lecturer" to the Lecturer group of the course "ws24-ios", so
//     the course "ios/Lecturer" would be provisioned inside it and its creator inherit its role.
func (suite *CourseKeycloakTestSuite) TestCreateCourseRejectsAHyphenOrASlashBeforeKeycloak() {
	for _, invalid := range []struct{ name, semesterTag string }{
		{"lab", "ws24-ios"},
		{"ios-lab", "ws24"},
		{"ios/Lecturer", "ws24"},
		{"lab", "ws24/ios"},
	} {
		suite.Run(invalid.semesterTag+" "+invalid.name, func() {
			*suite.keycloak = courseKeycloakSpy{}

			_, err := suite.service.CreateCourse(suite.ctx, courseToCreate(invalid.name, invalid.semesterTag), "creator")

			assert.ErrorIs(suite.T(), err, ErrInvalidCourseIdentifier)
			assert.Empty(suite.T(), suite.keycloak.created)
			assert.Empty(suite.T(), suite.keycloak.cleanedUp)
			assert.Zero(suite.T(), suite.countCourses(invalid.name, invalid.semesterTag))
		})
	}
}

func (suite *CourseKeycloakTestSuite) TestCreateCourseAnswersAHyphenWithBadRequest() {
	body, err := json.Marshal(courseToCreate("lab", "ws24-ios"))
	require.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/api/courses/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	newAdminCourseRouter(suite.service).ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
	assert.Contains(suite.T(), resp.Body.String(), "the semester tag must not contain")
	assert.Empty(suite.T(), suite.keycloak.created)
}

func TestCourseKeycloakTestSuite(t *testing.T) {
	suite.Run(t, new(CourseKeycloakTestSuite))
}
