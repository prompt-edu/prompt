package course

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/course/courseDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CourseRouterTestSuite struct {
	suite.Suite
	router        *gin.Engine
	ctx           context.Context
	cleanup       func()
	courseService CourseService
}

func (suite *CourseRouterTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Set up PostgreSQL container
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/copy_course_test.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}

	mockCreateGroupsAndRoles := func(ctx context.Context, courseName, iterationName, userID string) error {
		// No-op or add assertions for test
		return nil
	}

	suite.cleanup = cleanup
	suite.courseService = CourseService{
		queries:                    *testDB.Queries,
		conn:                       testDB.Conn,
		createCourseGroupsAndRoles: mockCreateGroupsAndRoles,
	}

	CourseServiceSingleton = &suite.courseService

	// Init the permissionValidation service
	permissionValidation.InitValidationService(*testDB.Queries, testDB.Conn)

	// Initialize router
	suite.router = gin.Default()
	api := suite.router.Group("/api")
	setupCourseRouter(api, func() gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddleware([]string{"PROMPT_Admin", "iPraktikum-ios24245-Lecturer"})
	}, sdkTestUtils.MockPermissionMiddleware, sdkTestUtils.MockPermissionMiddleware)
	coursePhase.InitCoursePhaseModule(api, *testDB.Queries, testDB.Conn)
}

func (suite *CourseRouterTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *CourseRouterTestSuite) TestGetAllCourses() {
	req, _ := http.NewRequest("GET", "/api/courses/", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var courses []courseDTO.CourseWithPhases
	err := json.Unmarshal(resp.Body.Bytes(), &courses)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(courses), 0, "Expected at least one course")
}

func (suite *CourseRouterTestSuite) TestGetCourseByID() {
	courseID := "c1f8060d-7381-4b64-a6ea-5ba8e8ac88dd"
	req, _ := http.NewRequest("GET", "/api/courses/"+courseID, nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var course courseDTO.CourseWithPhases
	err := json.Unmarshal(resp.Body.Bytes(), &course)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uuid.MustParse(courseID), course.ID, "Course ID should match")
	assert.NotEmpty(suite.T(), course.CoursePhases, "Course should have phases")
}

func (suite *CourseRouterTestSuite) TestCreateCourse() {
	newCourse := courseDTO.CreateCourse{
		Name:                "Router Test Course",
		StartDate:           pgtype.Date{Valid: true, Time: time.Now()},
		EndDate:             pgtype.Date{Valid: true, Time: time.Now().Add(24 * time.Hour)},
		SemesterTag:         pgtype.Text{String: "WS2024", Valid: true},
		RestrictedData:      meta.MetaData{"key": "value"},
		StudentReadableData: meta.MetaData{"icon": "test-icon"},
		CourseType:          "practical course",
		Ects:                pgtype.Int4{Int32: 10, Valid: true},
	}
	body, _ := json.Marshal(newCourse)
	req, _ := http.NewRequest("POST", "/api/courses/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusCreated, resp.Code)

	var createdCourse courseDTO.Course
	err := json.Unmarshal(resp.Body.Bytes(), &createdCourse)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newCourse.Name, createdCourse.Name, "Course name should match")
	assert.Equal(suite.T(), "practical course", createdCourse.CourseType, "Course type should match")
}

func (suite *CourseRouterTestSuite) TestArchiveCourse() {
	courseID := "c1f8060d-7381-4b64-a6ea-5ba8e8ac88dd"

	// Send archive request
	updateRequest := courseDTO.CourseArchiveStatus{
		Archived: true,
	}
	body, _ := json.Marshal(updateRequest)
	req, _ := http.NewRequest("PUT", "/api/courses/"+courseID+"/archive", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	// Verify DB state
	courseUUID := uuid.MustParse(courseID)
	updatedCourse, err := CourseServiceSingleton.queries.GetCourse(suite.ctx, courseUUID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), updatedCourse.Archived, "Course should be archived")
	assert.True(suite.T(), updatedCourse.ArchivedOn.Valid, "ArchivedOn should be set")
}

func (suite *CourseRouterTestSuite) TestUnarchiveCourse() {
	courseID := "c1f8060d-7381-4b64-a6ea-5ba8e8ac88dd"

	// FIRST archive the course
	firstUpdate := courseDTO.CourseArchiveStatus{
		Archived: true,
	}
	body1, _ := json.Marshal(firstUpdate)
	req1, _ := http.NewRequest("PUT", "/api/courses/"+courseID+"/archive", bytes.NewBuffer(body1))
	req1.Header.Set("Content-Type", "application/json")
	resp1 := httptest.NewRecorder()
	suite.router.ServeHTTP(resp1, req1)
	assert.Equal(suite.T(), http.StatusOK, resp1.Code)

	// THEN unarchive it
	secondUpdate := courseDTO.CourseArchiveStatus{
		Archived: false,
	}
	body2, _ := json.Marshal(secondUpdate)
	req2, _ := http.NewRequest("PUT", "/api/courses/"+courseID+"/archive", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	resp2 := httptest.NewRecorder()

	suite.router.ServeHTTP(resp2, req2)
	assert.Equal(suite.T(), http.StatusOK, resp2.Code)

	// Verify DB state
	courseUUID := uuid.MustParse(courseID)
	updatedCourse, err := CourseServiceSingleton.queries.GetCourse(suite.ctx, courseUUID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), updatedCourse.Archived, "Course should be unarchived")
	assert.False(suite.T(), updatedCourse.ArchivedOn.Valid, "ArchivedOn should be NULL")
}

func (suite *CourseRouterTestSuite) TestUpdateCoursePhaseOrder() {
	courseID := "c1f8060d-7381-4b64-a6ea-5ba8e8ac88dd"
	firstUUID := uuid.MustParse("bd727106-2dc0-4c44-a804-2efde26101ae")
	secondUUID := uuid.MustParse("93693f81-9c49-4183-ae70-c0ee3742560d")
	thirdUUID := uuid.MustParse("0bf6eb6c-ff6f-40f4-af63-a005e2c8d123")

	// Construct the updated phase graph: first -> second -> third
	updateGraphRequest := courseDTO.UpdateCoursePhaseGraph{
		InitialPhase: firstUUID,
		PhaseGraph: []courseDTO.CoursePhaseGraph{
			{
				FromCoursePhaseID: firstUUID,
				ToCoursePhaseID:   secondUUID,
			},
			{
				FromCoursePhaseID: secondUUID,
				ToCoursePhaseID:   thirdUUID,
			},
		},
	}

	body, _ := json.Marshal(updateGraphRequest)
	req, _ := http.NewRequest("PUT", "/api/courses/"+courseID+"/phase_graph", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code, "Expected 200 OK when updating course phase graph")

	// Verify the updated order by fetching the course again
	req, _ = http.NewRequest("GET", "/api/courses/"+courseID, nil)
	resp = httptest.NewRecorder()
	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code, "Expected 200 OK when fetching the updated course")

	var updatedCourse courseDTO.CourseWithPhases
	err := json.Unmarshal(resp.Body.Bytes(), &updatedCourse)
	assert.NoError(suite.T(), err, "Unmarshalling the response should not produce an error")

	// Validate the phases in the updated course
	var firstCoursePhase *coursePhaseDTO.CoursePhaseSequence
	var secondCoursePhase *coursePhaseDTO.CoursePhaseSequence
	var thirdCoursePhase *coursePhaseDTO.CoursePhaseSequence

	for _, phase := range updatedCourse.CoursePhases {
		switch phase.SequenceOrder {
		case 1:
			firstCoursePhase = &phase
		case 2:
			secondCoursePhase = &phase
		case 3:
			thirdCoursePhase = &phase
		}
	}

	assert.NotNil(suite.T(), firstCoursePhase, "First phase should be present")
	assert.NotNil(suite.T(), secondCoursePhase, "Second phase should be present")
	assert.NotNil(suite.T(), thirdCoursePhase, "Third phase should be present")

	assert.Equal(suite.T(), firstUUID, firstCoursePhase.ID, "First phase ID should match initial phase")
	assert.Equal(suite.T(), secondUUID, secondCoursePhase.ID, "Second phase ID should match")
	assert.Equal(suite.T(), thirdUUID, thirdCoursePhase.ID, "Third phase ID should match")

	assert.True(suite.T(), firstCoursePhase.IsInitialPhase, "First phase should be the initial phase")
	assert.False(suite.T(), secondCoursePhase.IsInitialPhase, "Second phase should not be the initial phase")
	assert.False(suite.T(), thirdCoursePhase.IsInitialPhase, "Third phase should not be the initial phase")
}

func (suite *CourseRouterTestSuite) TestUpdateCourseTemplateStatus() {
	courseID := "c1f8060d-7381-4b64-a6ea-5ba8e8ac88dd"
	updateRequest := courseDTO.CourseTemplateStatus{
		IsTemplate: true,
	}
	body, _ := json.Marshal(updateRequest)
	req, _ := http.NewRequest("PUT", "/api/courses/"+courseID+"/template", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	courseUUID := uuid.MustParse(courseID)
	updatedCourse, err := CourseServiceSingleton.queries.GetTemplateCourseByID(suite.ctx, courseUUID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), updatedCourse.Template, "Course should be marked as a template")
}

func (suite *CourseRouterTestSuite) TestGetTemplateCourses() {
	req, _ := http.NewRequest("GET", "/api/courses/template", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var courses []courseDTO.Course
	err := json.Unmarshal(resp.Body.Bytes(), &courses)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(courses), 0, "Expected at least one template course")
	for _, course := range courses {
		assert.True(suite.T(), course.Template, "All returned courses should be templates")
	}
}

func (suite *CourseRouterTestSuite) TestCheckCourseTemplateStatusTrue() {
	courseID := "c1f8060d-7381-4b64-a6ea-5ba8e8ac88ee"
	req, _ := http.NewRequest("GET", "/api/courses/"+courseID+"/template", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var status courseDTO.CourseTemplateStatus
	err := json.Unmarshal(resp.Body.Bytes(), &status)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), status.IsTemplate, "Course should be a template")
}

func (suite *CourseRouterTestSuite) TestCheckCourseTemplateStatusFalse() {
	courseID := "c1f8060d-7381-4b64-a6ea-5ba8e8ac88dd"
	req, _ := http.NewRequest("GET", "/api/courses/"+courseID+"/template", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var status courseDTO.CourseTemplateStatus
	err := json.Unmarshal(resp.Body.Bytes(), &status)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), status.IsTemplate, "Course should not be a template")
}

func TestCourseRouterTestSuite(t *testing.T) {
	suite.Run(t, new(CourseRouterTestSuite))
}
