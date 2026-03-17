package feedbackItem

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
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/feedbackItem/feedbackItemDTO"
)

type FeedbackItemRouterTestSuite struct {
	suite.Suite
	router   *gin.Engine
	suiteCtx context.Context
	cleanup  func()
	service  FeedbackItemService
}

func (suite *FeedbackItemRouterTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../../database_dumps/feedbackItems.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup

	suite.service = FeedbackItemService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	FeedbackItemServiceSingleton = &suite.service

	suite.router = gin.Default()

	// Add global middleware to set courseParticipationID for all requests
	suite.router.Use(func(c *gin.Context) {
		// Set the courseParticipationID that the feedback item router expects
		// This matches one of the course participations in our test database
		testCourseParticipationID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")
		c.Set("courseParticipationID", testCourseParticipationID)
		c.Next()
	})

	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "student1@example.com", "1234", "id")
	}

	// Setup router with middleware
	setupFeedbackItemRouter(api, testMiddleware)
}

func (suite *FeedbackItemRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

// Helper method to create a router with lecturer permissions
func (suite *FeedbackItemRouterTestSuite) createLecturerRouter() *gin.Engine {
	router := gin.Default()

	// Set lecturer course participation ID
	router.Use(func(c *gin.Context) {
		lecturerCourseParticipationID := uuid.MustParse("ea42e447-60f9-4fe0-b297-2dae3f924fd7")
		c.Set("courseParticipationID", lecturerCourseParticipationID)
		c.Next()
	})

	api := router.Group("/api/course_phase/:coursePhaseID")
	lecturerMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "lecturer@example.com", "1234", "lecturer_id")
	}
	setupFeedbackItemRouter(api, lecturerMiddleware)
	return router
}

func (suite *FeedbackItemRouterTestSuite) TestCreateFeedbackItemInvalidJSON() {
	phaseID := uuid.New()
	req, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/evaluation/feedback-items", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *FeedbackItemRouterTestSuite) TestCreateFeedbackItemValid() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	studentID := uuid.MustParse("da42e447-60f9-4fe0-b297-2dae3f924fd7") // target student
	authorID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")  // current student

	payload := feedbackItemDTO.CreateFeedbackItemRequest{
		FeedbackType:                db.FeedbackTypePositive,
		FeedbackText:                "Test positive feedback",
		CourseParticipationID:       studentID,
		CoursePhaseID:               phaseID,
		AuthorCourseParticipationID: authorID,
		Type:                        assessmentType.Self,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/evaluation/feedback-items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
}

func (suite *FeedbackItemRouterTestSuite) TestCreateFeedbackItemUnauthorizedAuthor() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	studentID := uuid.MustParse("da42e447-60f9-4fe0-b297-2dae3f924fd7")
	wrongAuthorID := uuid.MustParse("ea42e447-60f9-4fe0-b297-2dae3f924fd7") // different author

	payload := feedbackItemDTO.CreateFeedbackItemRequest{
		FeedbackType:                db.FeedbackTypePositive,
		FeedbackText:                "Test feedback",
		CourseParticipationID:       studentID,
		CoursePhaseID:               phaseID,
		AuthorCourseParticipationID: wrongAuthorID,
		Type:                        assessmentType.Self,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/evaluation/feedback-items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusForbidden, resp.Code)
}

func (suite *FeedbackItemRouterTestSuite) TestGetMyFeedbackItems() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/evaluation/feedback-items/my-feedback", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var feedbackItems []feedbackItemDTO.FeedbackItem
	err := json.Unmarshal(resp.Body.Bytes(), &feedbackItems)
	assert.NoError(suite.T(), err)
}

func (suite *FeedbackItemRouterTestSuite) TestDeleteFeedbackItemValid() {
	feedbackItemID := uuid.MustParse("33333333-3333-3333-3333-333333333333") // authored by ca42e447 (current user)
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")

	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+phaseID.String()+"/evaluation/feedback-items/"+feedbackItemID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *FeedbackItemRouterTestSuite) TestDeleteFeedbackItemUnauthorized() {
	feedbackItemID := uuid.MustParse("11111111-1111-1111-1111-111111111111") // authored by da42e447 (different user)
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")

	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+phaseID.String()+"/evaluation/feedback-items/"+feedbackItemID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusForbidden, resp.Code)
}

func (suite *FeedbackItemRouterTestSuite) TestDeleteFeedbackItemInvalidID() {
	phaseID := uuid.New()
	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+phaseID.String()+"/evaluation/feedback-items/invalid-uuid", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *FeedbackItemRouterTestSuite) TestLecturerGetFeedbackItemsForStudent() {
	lecturerRouter := suite.createLecturerRouter()
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	studentID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")

	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/evaluation/feedback-items/course-participation/"+studentID.String(), nil)
	resp := httptest.NewRecorder()

	lecturerRouter.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var feedbackItems []feedbackItemDTO.FeedbackItem
	err := json.Unmarshal(resp.Body.Bytes(), &feedbackItems)
	assert.NoError(suite.T(), err)
}

func (suite *FeedbackItemRouterTestSuite) TestInvalidCoursePhaseID() {
	req, _ := http.NewRequest("GET", "/api/course_phase/invalid-uuid/evaluation/feedback-items", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *FeedbackItemRouterTestSuite) TestInvalidCourseParticipationID() {
	lecturerRouter := suite.createLecturerRouter()
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/evaluation/feedback-items/course-participation/invalid-uuid", nil)
	resp := httptest.NewRecorder()

	lecturerRouter.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func TestFeedbackItemRouterTestSuite(t *testing.T) {
	suite.Run(t, new(FeedbackItemRouterTestSuite))
}
