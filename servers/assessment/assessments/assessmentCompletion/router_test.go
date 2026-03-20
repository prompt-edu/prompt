package assessmentCompletion

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	dto "github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion/assessmentCompletionDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type AssessmentCompletionRouterTestSuite struct {
	suite.Suite
	router   *gin.Engine
	suiteCtx context.Context
	cleanup  func()
	service  AssessmentCompletionService
}

func (suite *AssessmentCompletionRouterTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../../database_dumps/assessmentCompletions.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup

	suite.service = AssessmentCompletionService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	AssessmentCompletionServiceSingleton = &suite.service

	coursePhaseConfig.CoursePhaseConfigSingleton = coursePhaseConfig.NewCoursePhaseConfigService(*testDB.Queries, testDB.Conn)

	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "user@example.com", "1234", "id")
	}
	// attach routes
	setupAssessmentCompletionRouter(api, testMiddleware)
}

func (suite *AssessmentCompletionRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *AssessmentCompletionRouterTestSuite) TestMarkAssessmentInvalidJSON() {
	phaseID := uuid.New()
	req, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *AssessmentCompletionRouterTestSuite) TestMarkAssessmentAsCompletedReturnsError() {
	// Use the course phase ID from test data to ensure there are competencies,
	// but use a random participant ID to ensure there are remaining assessments (no assessments for this participant)
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.New() // Random participant ID - no assessments for this participant
	// minimal JSON to bind
	payload := dto.AssessmentCompletion{
		CoursePhaseID:         phaseID,
		CourseParticipationID: partID,
		Author:                "tester",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/mark-complete", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	// service returns error when assessments remain (or student doesn't exist)
	assert.Equal(suite.T(), http.StatusInternalServerError, resp.Code)
}

func (suite *AssessmentCompletionRouterTestSuite) TestGetAssessmentCompletionNotFound() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.New()
	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/course-participation/"+partID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusInternalServerError, resp.Code)
}

func (suite *AssessmentCompletionRouterTestSuite) TestGetAssessmentCompletionInvalidIDs() {
	// invalid phase
	req1, _ := http.NewRequest("GET", "/api/course_phase/invalid-phase/student-assessment/completed/course-participation/"+uuid.New().String(), nil)
	rep1 := httptest.NewRecorder()
	suite.router.ServeHTTP(rep1, req1)
	assert.Equal(suite.T(), http.StatusBadRequest, rep1.Code)

	// invalid participation
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	req2, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/course-participation/invalid-uuid", nil)
	rep2 := httptest.NewRecorder()
	suite.router.ServeHTTP(rep2, req2)
	assert.Equal(suite.T(), http.StatusBadRequest, rep2.Code)
}

func (suite *AssessmentCompletionRouterTestSuite) TestUnmarkAssessmentAsCompletedNonExisting() {
	phaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	partID := uuid.New()
	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/course-participation/"+partID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *AssessmentCompletionRouterTestSuite) TestUnmarkAssessmentAsCompletedInvalidIDs() {
	// invalid phase
	req1, _ := http.NewRequest("DELETE", "/api/course_phase/invalid-phase/student-assessment/completed/course-participation/"+uuid.New().String(), nil)
	rep1 := httptest.NewRecorder()
	suite.router.ServeHTTP(rep1, req1)
	assert.Equal(suite.T(), http.StatusBadRequest, rep1.Code)

	// invalid participation
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	req2, _ := http.NewRequest("DELETE", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/course-participation/invalid-uuid", nil)
	rep2 := httptest.NewRecorder()
	suite.router.ServeHTTP(rep2, req2)
	assert.Equal(suite.T(), http.StatusBadRequest, rep2.Code)
}

func (suite *AssessmentCompletionRouterTestSuite) TestCreateOrUpdateAssessmentCompletion() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.New()

	// Test successful creation with Completed: false so we can update it later
	payload := dto.AssessmentCompletion{
		CoursePhaseID:         phaseID,
		CourseParticipationID: partID,
		Author:                "Test Author",
		Comment:               "Test comment",
		GradeSuggestion:       3.5,
		Completed:             false, // Start with incomplete so we can update it
		CompletedAt:           pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	body, _ := json.Marshal(payload)

	// Test POST endpoint
	req, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	// Test PUT endpoint (update) - keep it incomplete so update is allowed
	payload.Comment = "Updated comment"
	body, _ = json.Marshal(payload)
	req, _ = http.NewRequest("PUT", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *AssessmentCompletionRouterTestSuite) TestCreateOrUpdateAssessmentCompletionInvalidJSON() {
	phaseID := uuid.New()

	// Test with invalid JSON
	req, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)

	// Test PUT with invalid JSON
	req, _ = http.NewRequest("PUT", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *AssessmentCompletionRouterTestSuite) TestDeleteAssessmentCompletion() {
	phaseID := uuid.New()
	partID := uuid.New()

	// Test successful deletion
	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/course-participation/"+partID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *AssessmentCompletionRouterTestSuite) TestDeleteAssessmentCompletionInvalidIDs() {
	// Test with invalid phase ID
	req1, _ := http.NewRequest("DELETE", "/api/course_phase/invalid-phase/student-assessment/completed/course-participation/"+uuid.New().String(), nil)
	resp1 := httptest.NewRecorder()
	suite.router.ServeHTTP(resp1, req1)
	assert.Equal(suite.T(), http.StatusBadRequest, resp1.Code)

	// Test with invalid participation ID
	phaseID := uuid.New()
	req2, _ := http.NewRequest("DELETE", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/course-participation/invalid-uuid", nil)
	resp2 := httptest.NewRecorder()
	suite.router.ServeHTTP(resp2, req2)
	assert.Equal(suite.T(), http.StatusBadRequest, resp2.Code)
}

func (suite *AssessmentCompletionRouterTestSuite) TestMarkAssessmentAsCompleteEndpoint() {
	// Use the course phase ID from test data to ensure there are competencies,
	// but use a random participant ID to ensure there are remaining assessments
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.New() // Random participant ID - no assessments for this participant

	// Test with valid payload but expect error due to remaining assessments
	payload := dto.AssessmentCompletion{
		CoursePhaseID:         phaseID,
		CourseParticipationID: partID,
		Author:                "Test Author",
		CompletedAt:           pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/mark-complete", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	// Should return error because there are remaining assessments for the student
	assert.Equal(suite.T(), http.StatusInternalServerError, resp.Code)
}

func (suite *AssessmentCompletionRouterTestSuite) TestMarkAssessmentAsCompleteInvalidJSON() {
	phaseID := uuid.New()

	req, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/mark-complete", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *AssessmentCompletionRouterTestSuite) TestGetAssessmentCompletionSuccess() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")

	// First create an assessment completion using HTTP POST request with Completed: false
	payload := dto.AssessmentCompletion{
		CoursePhaseID:         phaseID,
		CourseParticipationID: partID,
		Author:                "Test Author",
		Comment:               "Test comment",
		GradeSuggestion:       4.0,
		Completed:             false, // Use false to avoid validation issues
		CompletedAt:           pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	body, _ := json.Marshal(payload)

	// Send HTTP POST request to create the assessment completion
	createReq, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed", bytes.NewBuffer(body))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()

	suite.router.ServeHTTP(createResp, createReq)
	assert.Equal(suite.T(), http.StatusOK, createResp.Code)

	// Verify the creation response
	var createResponse map[string]interface{}
	err := json.Unmarshal(createResp.Body.Bytes(), &createResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), createResponse, "message")

	// Now test GET endpoint
	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/course-participation/"+partID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	// Verify response contains the assessment completion
	var response dto.AssessmentCompletion
	err = json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Test Author", response.Author)
	assert.Equal(suite.T(), "Test comment", response.Comment)
}

func (suite *AssessmentCompletionRouterTestSuite) TestListAssessmentCompletionsByCoursePhase() {
	phaseID := uuid.New()

	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	// Verify response is a list
	var response []dto.AssessmentCompletion
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
}

func (suite *AssessmentCompletionRouterTestSuite) TestListAssessmentCompletionsInvalidPhaseID() {
	req, _ := http.NewRequest("GET", "/api/course_phase/invalid-phase/student-assessment/completed", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *AssessmentCompletionRouterTestSuite) TestUnmarkAssessmentAsCompletedEndpoint() {
	phaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	partID := uuid.New()

	// First create an assessment completion to unmark
	payload := dto.AssessmentCompletion{
		CoursePhaseID:         phaseID,
		CourseParticipationID: partID,
		Author:                "Test Author",
		Comment:               "Test comment",
		GradeSuggestion:       4.0,
		Completed:             true,
		CompletedAt:           pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	err := CreateOrUpdateAssessmentCompletion(suite.suiteCtx, payload)
	assert.NoError(suite.T(), err)

	// Test PUT /unmark endpoint
	req, _ := http.NewRequest("PUT", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/course-participation/"+partID.String()+"/unmark", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	// Verify the assessment completion was unmarked (not deleted, just unmarked)
	completion, err := GetAssessmentCompletion(suite.suiteCtx, partID, phaseID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), completion.Completed, "Expected completion to be unmarked")
	assert.Equal(suite.T(), "Test Author", completion.Author) // Other fields should remain
}

func (suite *AssessmentCompletionRouterTestSuite) TestUnmarkAssessmentAsCompletedEndpointNonExisting() {
	phaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	partID := uuid.New()

	// Test PUT /unmark endpoint on non-existing completion
	req, _ := http.NewRequest("PUT", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/course-participation/"+partID.String()+"/unmark", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code) // Should not error even if doesn't exist
}

func (suite *AssessmentCompletionRouterTestSuite) TestUnmarkAssessmentAsCompletedEndpointInvalidIDs() {
	// Test with invalid phase ID
	req1, _ := http.NewRequest("PUT", "/api/course_phase/invalid-phase/student-assessment/completed/course-participation/"+uuid.New().String()+"/unmark", nil)
	resp1 := httptest.NewRecorder()
	suite.router.ServeHTTP(resp1, req1)
	assert.Equal(suite.T(), http.StatusBadRequest, resp1.Code)

	// Test with invalid participation ID
	phaseID := uuid.New()
	req2, _ := http.NewRequest("PUT", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/course-participation/invalid-uuid/unmark", nil)
	resp2 := httptest.NewRecorder()
	suite.router.ServeHTTP(resp2, req2)
	assert.Equal(suite.T(), http.StatusBadRequest, resp2.Code)
}

func (suite *AssessmentCompletionRouterTestSuite) TestCreateOrUpdateAssessmentCompletionResponseFormat() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.New()

	payload := dto.AssessmentCompletion{
		CoursePhaseID:         phaseID,
		CourseParticipationID: partID,
		Author:                "Test Author",
		Comment:               "Test comment",
		GradeSuggestion:       4.0,
		Completed:             true,
		CompletedAt:           pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	body, _ := json.Marshal(payload)

	// Test POST response format
	req, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	// Verify response contains success message
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "message")
	assert.Equal(suite.T(), "Assessment completion created/updated successfully", response["message"])
}

func (suite *AssessmentCompletionRouterTestSuite) TestMarkAssessmentAsCompleteResponseFormat() {
	// Use the course phase ID from test data to ensure there are competencies,
	// but use a random participant ID to ensure there are remaining assessments
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.New() // Random participant ID - no assessments for this participant

	payload := dto.AssessmentCompletion{
		CoursePhaseID:         phaseID,
		CourseParticipationID: partID,
		Author:                "Test Author",
		CompletedAt:           pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/mark-complete", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	// Should return error because there are remaining assessments but verify format
	assert.Equal(suite.T(), http.StatusInternalServerError, resp.Code)

	// Verify error response format
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "error")
}

// Grade endpoint tests

func (suite *AssessmentCompletionRouterTestSuite) TestGetAllGradesEndpoint() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/grade", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var grades []dto.GradeWithParticipation
	err := json.Unmarshal(resp.Body.Bytes(), &grades)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), grades)

	// Verify structure of returned grades
	for _, grade := range grades {
		assert.NotEqual(suite.T(), uuid.Nil, grade.CourseParticipationID)
		assert.GreaterOrEqual(suite.T(), grade.Grade, 0.0)
	}
}

func (suite *AssessmentCompletionRouterTestSuite) TestGetAllGradesEndpointInvalidPhaseID() {
	req, _ := http.NewRequest("GET", "/api/course_phase/invalid-phase-id/student-assessment/completed/grade", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *AssessmentCompletionRouterTestSuite) TestGetAllGradesEndpointNonExistentPhase() {
	nonExistentPhaseID := uuid.New()
	req, _ := http.NewRequest("GET", "/api/course_phase/"+nonExistentPhaseID.String()+"/student-assessment/completed/grade", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var grades []dto.GradeWithParticipation
	err := json.Unmarshal(resp.Body.Bytes(), &grades)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), grades) // Should return empty array for non-existent phase
}

func (suite *AssessmentCompletionRouterTestSuite) TestGetStudentGradeEndpoint() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	participationID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")

	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/grade/course-participation/"+participationID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var grade float64
	err := json.Unmarshal(resp.Body.Bytes(), &grade)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), grade, 0.0)
}

func (suite *AssessmentCompletionRouterTestSuite) TestGetStudentGradeEndpointInvalidPhaseID() {
	participationID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")

	req, _ := http.NewRequest("GET", "/api/course_phase/invalid-phase/student-assessment/completed/grade/course-participation/"+participationID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *AssessmentCompletionRouterTestSuite) TestGetStudentGradeEndpointInvalidParticipationID() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")

	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/grade/course-participation/invalid-participation", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *AssessmentCompletionRouterTestSuite) TestGetStudentGradeEndpointNonExistentParticipation() {
	phaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	nonExistentParticipationID := uuid.New()

	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/grade/course-participation/"+nonExistentParticipationID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var grade float64
	err := json.Unmarshal(resp.Body.Bytes(), &grade)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0.0, grade) // Should return 0 for non-existent participation
}

func (suite *AssessmentCompletionRouterTestSuite) TestGetStudentGradeEndpointNonExistentPhase() {
	nonExistentPhaseID := uuid.New()
	participationID := uuid.MustParse("ca42e447-60f9-4fe0-b297-2dae3f924fd7")

	req, _ := http.NewRequest("GET", "/api/course_phase/"+nonExistentPhaseID.String()+"/student-assessment/completed/grade/course-participation/"+participationID.String(), nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var grade float64
	err := json.Unmarshal(resp.Body.Bytes(), &grade)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0.0, grade) // Should return 0 for non-existent phase
}

func (suite *AssessmentCompletionRouterTestSuite) TestGetMyGradeSuggestionWhenVisible() {
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
	setupAssessmentCompletionRouter(api, testMiddleware)

	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/my-grade-suggestion", nil)
	resp := httptest.NewRecorder()
	customRouter.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	var grade float64
	err := json.Unmarshal(resp.Body.Bytes(), &grade)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 4.5, grade)
}

func (suite *AssessmentCompletionRouterTestSuite) TestGetMyGradeSuggestionWhenNotVisible() {
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
	setupAssessmentCompletionRouter(api, testMiddleware)

	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/my-grade-suggestion", nil)
	resp := httptest.NewRecorder()
	customRouter.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusForbidden, resp.Code)
	var errorResponse map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &errorResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorResponse["error"], "not visible")
}

func (suite *AssessmentCompletionRouterTestSuite) TestGetMyGradeSuggestionBeforeDeadline() {
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
	setupAssessmentCompletionRouter(api, testMiddleware)

	req, _ := http.NewRequest("GET", "/api/course_phase/"+phaseID.String()+"/student-assessment/completed/my-grade-suggestion", nil)
	resp := httptest.NewRecorder()
	customRouter.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusNoContent, resp.Code)
}

func TestAssessmentCompletionRouterTestSuite(t *testing.T) {
	suite.Run(t, new(AssessmentCompletionRouterTestSuite))
}
