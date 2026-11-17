package copy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/course"
	"github.com/prompt-edu/prompt/servers/core/course/copy/courseCopyDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// copyKeycloakSpy records the names of the courses the Keycloak hooks are called with.
type copyKeycloakSpy struct {
	created   []string
	cleanedUp []string
	createErr error
}

func (k *copyKeycloakSpy) create(_ context.Context, courseName, _, _ string) error {
	k.created = append(k.created, courseName)
	return k.createErr
}

func (k *copyKeycloakSpy) cleanUp(_ context.Context, courseName, _ string) error {
	k.cleanedUp = append(k.cleanedUp, courseName)
	return nil
}

type CopyCourseTestSuite struct {
	suite.Suite
	ctx               context.Context
	cleanup           func()
	queries           *db.Queries
	keycloak          *copyKeycloakSpy
	courseCopyService *CourseCopyService
	sourceCourse      db.Course
}

func (suite *CopyCourseTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.ctx, "../../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../../database_dumps/copy_course_test.sql")
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}

	suite.cleanup = cleanup
	suite.queries = testDB.Queries
	suite.keycloak = &copyKeycloakSpy{}
	coursePhaseService := coursePhase.NewCoursePhaseService(*testDB.Queries, testDB.Conn, resolution.NewResolutionService("localhost:8080"))
	suite.courseCopyService = NewCourseCopyService(*testDB.Queries, testDB.Conn, coursePhaseService, suite.keycloak.create, suite.keycloak.cleanUp)

	// Use the known UUID from the dump
	sourceID := uuid.MustParse("c1f8060d-7381-4b64-a6ea-5ba8e8ac88dd")
	suite.sourceCourse, err = testDB.Queries.GetCourse(suite.ctx, sourceID)
	assert.NoError(suite.T(), err)
}

func (suite *CopyCourseTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *CopyCourseTestSuite) SetupTest() {
	*suite.keycloak = copyKeycloakSpy{}
}

func (suite *CopyCourseTestSuite) copyRequest(name, semesterTag string) courseCopyDTO.CopyCourseRequest {
	return courseCopyDTO.CopyCourseRequest{
		Name:        name,
		SemesterTag: pgtype.Text{Valid: true, String: semesterTag},
		StartDate:   pgtype.Date{Valid: true, Time: time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)},
		EndDate:     pgtype.Date{Valid: true, Time: time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)},
	}
}

func (suite *CopyCourseTestSuite) courseExists(name, semesterTag string) bool {
	exists, err := suite.queries.CheckCourseNameExists(suite.ctx, db.CheckCourseNameExistsParams{
		Name:        name,
		SemesterTag: pgtype.Text{Valid: true, String: semesterTag},
	})
	require.NoError(suite.T(), err)
	return exists
}

func (suite *CopyCourseTestSuite) TestCopyCourseInternal() {
	newName := "Cloned Course"
	newTag := "SS2025"
	newStartDate := pgtype.Date{Valid: true, Time: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	newEndDate := pgtype.Date{Valid: true, Time: time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)}

	copyReq := courseCopyDTO.CopyCourseRequest{
		Name:        newName,
		SemesterTag: pgtype.Text{Valid: true, String: newTag},
		StartDate:   newStartDate,
		EndDate:     newEndDate,
		Template:    false,
	}

	// Create a dummy *gin.Context for testing
	ginCtx, _ := gin.CreateTestContext(nil)
	result, err := suite.courseCopyService.copyCourseInternal(ginCtx, suite.sourceCourse.ID, copyReq, "test_user")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newName, result.Name)
	assert.Equal(suite.T(), newTag, result.SemesterTag.String)
	assert.Equal(suite.T(), string(suite.sourceCourse.CourseType), result.CourseType)
	assert.Equal(suite.T(), suite.sourceCourse.Ects, result.Ects)

	// Verify phase graph
	newGraph, err := suite.courseCopyService.queries.GetCoursePhaseGraph(suite.ctx, result.ID)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), newGraph, "Expected course phase graph to be copied")

	// Verify initial phase
	sequence, err := suite.courseCopyService.queries.GetCoursePhaseSequence(suite.ctx, result.ID)
	assert.NoError(suite.T(), err)
	foundInitial := false
	for _, p := range sequence {
		if p.IsInitialPhase {
			foundInitial = true
			break
		}
	}
	assert.True(suite.T(), foundInitial, "Expected initial phase to be set")
	assert.NotEmpty(suite.T(), sequence, "Expected course phase sequence to be copied")

	partGraph, err := suite.courseCopyService.queries.GetParticipationDataGraph(suite.ctx, result.ID)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), partGraph, "Expected participation meta data graph to be copied")

	// Verify application form questions
	applicationPhaseID, err := suite.courseCopyService.queries.GetApplicationPhaseIDForCourse(suite.ctx, result.ID)
	assert.NoError(suite.T(), err)

	textQuestions, err := suite.courseCopyService.queries.GetApplicationQuestionsTextForCoursePhase(suite.ctx, applicationPhaseID)
	assert.NoError(suite.T(), err)

	multiSelectQuestions, err := suite.courseCopyService.queries.GetApplicationQuestionsMultiSelectForCoursePhase(suite.ctx, applicationPhaseID)
	assert.NoError(suite.T(), err)

	assert.True(suite.T(), len(textQuestions) > 0 || len(multiSelectQuestions) > 0, "Expected application form to be copied")
}

func (suite *CopyCourseTestSuite) TestCopyCourseInternalWithDifferentCourseType() {
	assert.NotEqual(suite.T(), db.CourseTypeSeminar, suite.sourceCourse.CourseType)

	copyReq := courseCopyDTO.CopyCourseRequest{
		Name:        "Seminar Clone",
		SemesterTag: pgtype.Text{Valid: true, String: "ws2526"},
		StartDate:   pgtype.Date{Valid: true, Time: time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)},
		EndDate:     pgtype.Date{Valid: true, Time: time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)},
		CourseType:  db.CourseTypeSeminar,
		Ects:        pgtype.Int4{Valid: true, Int32: 5},
	}

	ginCtx, _ := gin.CreateTestContext(nil)
	result, err := suite.courseCopyService.copyCourseInternal(ginCtx, suite.sourceCourse.ID, copyReq, "test_user")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), string(db.CourseTypeSeminar), result.CourseType)
	assert.Equal(suite.T(), pgtype.Int4{Valid: true, Int32: 5}, result.Ects)

	sequence, err := suite.courseCopyService.queries.GetCoursePhaseSequence(suite.ctx, result.ID)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), sequence, "Expected course phases to be copied")
}

func (suite *CopyCourseTestSuite) TestCopyCourseCleansUpKeycloakWhenProvisioningFails() {
	suite.keycloak.createErr = errors.New("keycloak unavailable")

	ginCtx, _ := gin.CreateTestContext(nil)
	_, err := suite.courseCopyService.copyCourseInternal(ginCtx, suite.sourceCourse.ID, suite.copyRequest("Failed Clone", "ws2526"), "test_user")

	assert.ErrorIs(suite.T(), err, suite.keycloak.createErr)
	assert.Equal(suite.T(), []string{"Failed Clone"}, suite.keycloak.cleanedUp)
	assert.False(suite.T(), suite.courseExists("Failed Clone", "ws2526"), "the copy must be rolled back")
}

func (suite *CopyCourseTestSuite) TestCopyCourseRejectsAnIdentifierAnotherCourseHas() {
	ginCtx, _ := gin.CreateTestContext(nil)
	_, err := suite.courseCopyService.copyCourseInternal(ginCtx, suite.sourceCourse.ID, suite.copyRequest("ios-lab", "ws25"), "test_user")
	require.NoError(suite.T(), err)

	_, err = suite.courseCopyService.copyCourseInternal(ginCtx, suite.sourceCourse.ID, suite.copyRequest("lab", "ws25-ios"), "test_user")

	assert.ErrorIs(suite.T(), err, course.ErrCourseIdentifierTaken)
	assert.Equal(suite.T(), []string{"ios-lab"}, suite.keycloak.created, "Keycloak must not be touched for the second copy")
	assert.Empty(suite.T(), suite.keycloak.cleanedUp, "the first copy's Keycloak groups must survive")
	assert.False(suite.T(), suite.courseExists("lab", "ws25-ios"))
}

func TestCopyCourseTestSuite(t *testing.T) {
	suite.Run(t, new(CopyCourseTestSuite))
}
