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
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/feedbackItem/feedbackItemDTO"
)

type FeedbackItemRouterTestSuite struct {
	suite.Suite
	router   *gin.Engine
	suiteCtx context.Context
	cleanup  func()
	service  *FeedbackItemService
}

func (suite *FeedbackItemRouterTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../../database_dumps/feedbackItems.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup

	suite.service = NewFeedbackItemService(*testDB.Queries, testDB.Conn,
		evaluationCompletion.NewEvaluationCompletionService(*testDB.Queries, testDB.Conn, coursePhaseConfig.GetTeamsForCoursePhase))

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
	RegisterRoutes(api, suite.service, testMiddleware)
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
	RegisterRoutes(api, suite.service, lecturerMiddleware)
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
	authorID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7") // current student

	payload := feedbackItemDTO.CreateFeedbackItemRequest{
		FeedbackType:                db.FeedbackTypePositive,
		FeedbackText:                "Test positive feedback",
		CourseParticipationID:       authorID, // self feedback: subject is the author
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

func (suite *FeedbackItemRouterTestSuite) TestCreateFeedbackItemSelfTargetMismatch() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	victimID := uuid.MustParse("da42e447-60f9-4fe0-b297-2dae3f924fd7") // peer, not the author
	authorID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7") // current student

	payload := feedbackItemDTO.CreateFeedbackItemRequest{
		FeedbackType:                db.FeedbackTypeNegative,
		FeedbackText:                "Injected into a peer's self feedback",
		CourseParticipationID:       victimID,
		CoursePhaseID:               phaseID,
		AuthorCourseParticipationID: authorID,
		Type:                        assessmentType.Self,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/evaluation/feedback-items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *FeedbackItemRouterTestSuite) TestCreateFeedbackItemBeforeWindowOpens() {
	notStartedPhaseID := uuid.MustParse("44561b6b-3c3a-4bc6-ba42-69eeb1514da9")
	authorID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")

	payload := feedbackItemDTO.CreateFeedbackItemRequest{
		FeedbackType:                db.FeedbackTypePositive,
		FeedbackText:                "Submitted before the window opens",
		CourseParticipationID:       authorID,
		CoursePhaseID:               notStartedPhaseID,
		AuthorCourseParticipationID: authorID,
		Type:                        assessmentType.Self,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+notStartedPhaseID.String()+"/evaluation/feedback-items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusForbidden, resp.Code)
}

func (suite *FeedbackItemRouterTestSuite) TestCreateFeedbackItemAfterCompletion() {
	completedPhaseID := uuid.MustParse("34561b6b-3c3a-4bc6-ba42-69eeb1514da9")
	authorID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")

	payload := feedbackItemDTO.CreateFeedbackItemRequest{
		FeedbackType:                db.FeedbackTypePositive,
		FeedbackText:                "Added after marking the evaluation complete",
		CourseParticipationID:       authorID,
		CoursePhaseID:               completedPhaseID,
		AuthorCourseParticipationID: authorID,
		Type:                        assessmentType.Self,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+completedPhaseID.String()+"/evaluation/feedback-items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusConflict, resp.Code)
}

func (suite *FeedbackItemRouterTestSuite) TestCreateFeedbackItemRejectsAssessmentType() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	authorID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")

	payload := feedbackItemDTO.CreateFeedbackItemRequest{
		FeedbackType:                db.FeedbackTypePositive,
		FeedbackText:                "Forged tutor-side assessment feedback",
		CourseParticipationID:       authorID,
		CoursePhaseID:               phaseID,
		AuthorCourseParticipationID: authorID,
		Type:                        assessmentType.Assessment,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/evaluation/feedback-items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
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

// seedOwnFeedbackItem inserts a self feedback item authored by the current test student
// so update tests do not depend on rows other tests mutate.
func (suite *FeedbackItemRouterTestSuite) seedOwnFeedbackItem(phaseID uuid.UUID) uuid.UUID {
	currentStudentID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")
	feedbackItemID := uuid.New()

	err := suite.service.queries.CreateFeedbackItem(suite.suiteCtx, db.CreateFeedbackItemParams{
		ID:                          feedbackItemID,
		FeedbackType:                db.FeedbackTypeNegative,
		FeedbackText:                "Original feedback text",
		CourseParticipationID:       currentStudentID,
		CoursePhaseID:               phaseID,
		AuthorCourseParticipationID: currentStudentID,
		Type:                        db.AssessmentTypeSelf,
	})
	assert.NoError(suite.T(), err)
	return feedbackItemID
}

func (suite *FeedbackItemRouterTestSuite) updateRequest(phaseID, feedbackItemID uuid.UUID) *httptest.ResponseRecorder {
	payload := feedbackItemDTO.UpdateFeedbackItemRequest{
		FeedbackType: db.FeedbackTypePositive,
		FeedbackText: "Rewritten feedback text",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PUT", "/api/course_phase/"+phaseID.String()+"/evaluation/feedback-items/"+feedbackItemID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	return resp
}

func (suite *FeedbackItemRouterTestSuite) TestUpdateFeedbackItemValid() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	feedbackItemID := suite.seedOwnFeedbackItem(phaseID)

	resp := suite.updateRequest(phaseID, feedbackItemID)
	assert.Equal(suite.T(), http.StatusCreated, resp.Code)

	updated, err := suite.service.GetFeedbackItem(suite.suiteCtx, feedbackItemID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Rewritten feedback text", updated.FeedbackText)
	assert.Equal(suite.T(), db.FeedbackTypePositive, updated.FeedbackType)
}

func (suite *FeedbackItemRouterTestSuite) TestUpdateFeedbackItemNotAuthor() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	feedbackItemID := uuid.MustParse("22222222-2222-2222-2222-222222222222") // authored by ea42e447 (lecturer)

	resp := suite.updateRequest(phaseID, feedbackItemID)
	assert.Equal(suite.T(), http.StatusForbidden, resp.Code)

	unchanged, err := suite.service.GetFeedbackItem(suite.suiteCtx, feedbackItemID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Need to improve time management", unchanged.FeedbackText)
	assert.Equal(suite.T(), uuid.MustParse("ea42e447-60f9-4fe0-b297-2dae3f924fd7"), unchanged.AuthorCourseParticipationID)
}

func (suite *FeedbackItemRouterTestSuite) TestUpdateFeedbackItemWrongPhase() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	otherPhaseID := uuid.MustParse("34561b6b-3c3a-4bc6-ba42-69eeb1514da9")
	feedbackItemID := suite.seedOwnFeedbackItem(phaseID)

	resp := suite.updateRequest(otherPhaseID, feedbackItemID)
	assert.Equal(suite.T(), http.StatusNotFound, resp.Code)

	unchanged, err := suite.service.GetFeedbackItem(suite.suiteCtx, feedbackItemID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Original feedback text", unchanged.FeedbackText)
}

func (suite *FeedbackItemRouterTestSuite) TestUpdateFeedbackItemUnknownID() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")

	resp := suite.updateRequest(phaseID, uuid.New())
	assert.Equal(suite.T(), http.StatusNotFound, resp.Code)
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
