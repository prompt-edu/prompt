package actionItem

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/actionItem/actionItemDTO"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type ActionItemRouterTestSuite struct {
	suite.Suite
	router   *gin.Engine
	suiteCtx context.Context
	cleanup  func()
	service  ActionItemService
}

func (suite *ActionItemRouterTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../../database_dumps/assessments.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup

	suite.service = ActionItemService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	ActionItemServiceSingleton = &suite.service

	// Initialize the assessment completion module with the test database
	// This creates a dummy router group just to initialize the singleton
	dummyRouter := gin.Default()
	dummyGroup := dummyRouter.Group("/dummy")
	assessmentCompletion.InitAssessmentCompletionModule(dummyGroup, *testDB.Queries, testDB.Conn)

	coursePhaseConfig.CoursePhaseConfigSingleton = coursePhaseConfig.NewCoursePhaseConfigService(*testDB.Queries, testDB.Conn)

	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "user@example.com", "1234", "id")
	}
	// attach routes
	setupActionItemRouter(api, testMiddleware)
}

func (suite *ActionItemRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *ActionItemRouterTestSuite) TestCreateActionItemInvalidJSON() {
	phaseID := uuid.New()
	req, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/student-assessment/action-item", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *ActionItemRouterTestSuite) TestCreateActionItemValid() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")

	payload := actionItemDTO.CreateActionItemRequest{
		CoursePhaseID:         phaseID,
		CourseParticipationID: partID,
		Action:                "Test action item",
		Author:                "tester",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/student-assessment/action-item", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
}

func (suite *ActionItemRouterTestSuite) TestGetActionItemsForStudentValid() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")
	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/action-item/course-participation/"+partID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var items []actionItemDTO.ActionItem
	err := json.Unmarshal(resp.Body.Bytes(), &items)
	assert.NoError(suite.T(), err)
}

func (suite *ActionItemRouterTestSuite) TestUpdateActionItemInvalidID() {
	phaseID := uuid.New()
	req, _ := http.NewRequest("PUT", "/api/course_phase/"+phaseID.String()+"/student-assessment/action-item/invalid-uuid", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *ActionItemRouterTestSuite) TestDeleteActionItemInvalidID() {
	phaseID := uuid.New()
	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+phaseID.String()+"/student-assessment/action-item/invalid-uuid", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *ActionItemRouterTestSuite) TestGetActionItemsForStudentInvalidIDs() {
	// invalid phase
	req1, _ := http.NewRequest("GET", "/api/course_phase/invalid-phase/student-assessment/action-item/course-participation/"+uuid.New().String(), nil)
	rep1 := httptest.NewRecorder()
	suite.router.ServeHTTP(rep1, req1)
	assert.Equal(suite.T(), http.StatusBadRequest, rep1.Code)

	// invalid participation
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	req2, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/action-item/course-participation/invalid-uuid", nil)
	rep2 := httptest.NewRecorder()
	suite.router.ServeHTTP(rep2, req2)
	assert.Equal(suite.T(), http.StatusBadRequest, rep2.Code)
}

func (suite *ActionItemRouterTestSuite) TestGetAllActionItemsForCoursePhaseCommunication() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/action-item/action", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var items []actionItemDTO.ActionItemWithParticipation
	err := json.Unmarshal(resp.Body.Bytes(), &items)
	assert.NoError(suite.T(), err, "Should be able to unmarshal action items with participation")

	// Verify structure of response
	for _, item := range items {
		assert.NotEmpty(suite.T(), item.CourseParticipationID, "CourseParticipationID should not be empty")
		assert.NotNil(suite.T(), item.ActionItems, "ActionItems should not be nil")
	}
}

func (suite *ActionItemRouterTestSuite) TestGetAllActionItemsForCoursePhaseCommunicationInvalidID() {
	req, _ := http.NewRequest("GET", "/api/course_phase/invalid-phase-id/student-assessment/action-item/action", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *ActionItemRouterTestSuite) TestGetStudentActionItemsForCoursePhaseCommunication() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")
	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/action-item/action/course-participation/"+partID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var items []string
	err := json.Unmarshal(resp.Body.Bytes(), &items)
	assert.NoError(suite.T(), err, "Should be able to unmarshal action items as string array")
}

func (suite *ActionItemRouterTestSuite) TestGetStudentActionItemsForCoursePhaseCommunicationInvalidPhaseID() {
	partID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")
	req, _ := http.NewRequest("GET", "/api/course_phase/invalid-phase-id/student-assessment/action-item/action/course-participation/"+partID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *ActionItemRouterTestSuite) TestGetStudentActionItemsForCoursePhaseCommunicationInvalidParticipationID() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/action-item/action/course-participation/invalid-participation-id", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *ActionItemRouterTestSuite) TestGetStudentActionItemsForCoursePhaseCommunicationBothInvalidIDs() {
	req, _ := http.NewRequest("GET", "/api/course_phase/invalid-phase-id/student-assessment/action-item/action/course-participation/invalid-participation-id", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *ActionItemRouterTestSuite) TestGetMyActionItemsWhenVisible() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	customRouter := gin.Default()
	api := customRouter.Group("/api/course_phase/:coursePhaseID")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("courseParticipationID", partID)
			sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "user@example.com", "1234", "id")(c)
		}
	}
	setupActionItemRouter(api, testMiddleware)

	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/action-item/my-action-items", nil)
	resp := httptest.NewRecorder()
	customRouter.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	var items []actionItemDTO.ActionItem
	err := json.Unmarshal(resp.Body.Bytes(), &items)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(items), 0, "Should have at least one action item")
}

func (suite *ActionItemRouterTestSuite) TestGetMyActionItemsWhenNotVisible() {
	phaseID := uuid.MustParse("3517a3e3-fe60-40e0-8a5e-8f39049c12c3")
	partID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	customRouter := gin.Default()
	api := customRouter.Group("/api/course_phase/:coursePhaseID")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("courseParticipationID", partID)
			sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "user@example.com", "1234", "id")(c)
		}
	}
	setupActionItemRouter(api, testMiddleware)

	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/action-item/my-action-items", nil)
	resp := httptest.NewRecorder()
	customRouter.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusForbidden, resp.Code)
	var errorResponse map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &errorResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorResponse["error"], "not visible")
}

func (suite *ActionItemRouterTestSuite) TestGetMyActionItemsBeforeDeadline() {
	phaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	partID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	customRouter := gin.Default()
	api := customRouter.Group("/api/course_phase/:coursePhaseID")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("courseParticipationID", partID)
			sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "user@example.com", "1234", "id")(c)
		}
	}
	setupActionItemRouter(api, testMiddleware)

	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/action-item/my-action-items", nil)
	resp := httptest.NewRecorder()
	customRouter.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	var items []actionItemDTO.ActionItem
	err := json.Unmarshal(resp.Body.Bytes(), &items)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, len(items), "Should return empty array before deadline")
}

func TestActionItemRouterTestSuite(t *testing.T) {
	suite.Run(t, new(ActionItemRouterTestSuite))
}
