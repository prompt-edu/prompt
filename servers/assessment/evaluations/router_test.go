package evaluations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/evaluations/evaluationDTO"
	"github.com/prompt-edu/prompt/servers/assessment/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type EvaluationRouterTestSuite struct {
	suite.Suite
	router            *gin.Engine
	suiteCtx          context.Context
	cleanup           func()
	mockCoreCleanup   func()
	evaluationService *EvaluationService
	testCoursePhaseID uuid.UUID
	ownTutorID        uuid.UUID
	ownTutorFound     bool
	ownTutorErr       error
}

// resolveOwnTutor stands in for the team phase lookup, so each test decides who the caller is.
func (suite *EvaluationRouterTestSuite) resolveOwnTutor(string, uuid.UUID) (uuid.UUID, bool, error) {
	return suite.ownTutorID, suite.ownTutorFound, suite.ownTutorErr
}

func (suite *EvaluationRouterTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.suiteCtx, "../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../database_dumps/evaluations.sql")
	if err != nil {
		suite.T().Fatalf("Failed to setup test database: %v", err)
	}

	_, mockCleanup := testutils.SetupMockCoreService()
	suite.mockCoreCleanup = mockCleanup

	suite.cleanup = cleanup
	coursePhaseConfigService := coursePhaseConfig.NewCoursePhaseConfigService(*testDB.Queries, testDB.Conn, assessmentSchemas.NewAssessmentSchemaService(*testDB.Queries, testDB.Conn))
	suite.evaluationService = NewEvaluationService(*testDB.Queries, testDB.Conn,
		evaluationCompletion.NewEvaluationCompletionService(*testDB.Queries, testDB.Conn, coursePhaseConfig.GetTeamsForCoursePhase, coursePhaseConfigService),
		coursePhaseConfigService, suite.resolveOwnTutor)

	suite.testCoursePhaseID = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")

	suite.router = gin.Default()

	// Add global middleware to set courseParticipationID for all requests
	suite.router.Use(func(c *gin.Context) {
		// Set the courseParticipationID that the evaluations router expects
		// This matches one of the course participations in our test database
		testCourseParticipationID := uuid.MustParse("01234567-1234-1234-1234-123456789012")
		c.Set("courseParticipationID", testCourseParticipationID)
		c.Next()
	})

	api := suite.router.Group("/assessment/api/course_phase/:coursePhaseID")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "existingstudent@example.com", "03711111", "ab12cde")
	}
	RegisterRoutes(api, suite.evaluationService, testMiddleware)
}

func (suite *EvaluationRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
	if suite.mockCoreCleanup != nil {
		suite.mockCoreCleanup()
	}
}

func TestEvaluationRouterTestSuite(t *testing.T) {
	suite.Run(t, new(EvaluationRouterTestSuite))
}

// Test Admin/Lecturer endpoints
func (suite *EvaluationRouterTestSuite) TestGetAllEvaluationsByPhase() {
	req, _ := http.NewRequest(http.MethodGet, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var evaluations []evaluationDTO.Evaluation
	err := json.Unmarshal(w.Body.Bytes(), &evaluations)
	assert.NoError(suite.T(), err)
}

func (suite *EvaluationRouterTestSuite) TestGetSelfEvaluationsForParticipantInPhase() {
	courseParticipationID := uuid.MustParse("01234567-1234-1234-1234-123456789012")
	req, _ := http.NewRequest(http.MethodGet, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation/self/"+courseParticipationID.String(), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var evaluations []evaluationDTO.Evaluation
	err := json.Unmarshal(w.Body.Bytes(), &evaluations)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), evaluations)
	for _, evaluation := range evaluations {
		assert.Equal(suite.T(), assessmentType.Self, evaluation.Type)
	}
}

func (suite *EvaluationRouterTestSuite) TestGetPeerEvaluationsForParticipantInPhase() {
	courseParticipationID := uuid.MustParse("01234567-1234-1234-1234-123456789012")
	req, _ := http.NewRequest(http.MethodGet, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation/peer/"+courseParticipationID.String(), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var evaluations []evaluationDTO.Evaluation
	err := json.Unmarshal(w.Body.Bytes(), &evaluations)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), evaluations)
	for _, evaluation := range evaluations {
		assert.Equal(suite.T(), assessmentType.Peer, evaluation.Type)
	}
}

func (suite *EvaluationRouterTestSuite) TestInvalidCourseParticipationIDForTypedEvaluationResults() {
	req, _ := http.NewRequest(http.MethodGet, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation/self/invalid-uuid", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// Test Student endpoints
func (suite *EvaluationRouterTestSuite) TestGetMyEvaluations() {
	req, _ := http.NewRequest(http.MethodGet, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation/my-evaluations", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var evaluations []evaluationDTO.Evaluation
	err := json.Unmarshal(w.Body.Bytes(), &evaluations)
	assert.NoError(suite.T(), err)
}

func (suite *EvaluationRouterTestSuite) TestCreateOrUpdateEvaluation() {
	requestBody := evaluationDTO.CreateOrUpdateEvaluationRequest{
		CourseParticipationID:       uuid.MustParse("01234567-1234-1234-1234-123456789012"), // Self evaluation: subject is the author
		CompetencyID:                uuid.MustParse("c4234567-1234-1234-1234-123456789012"), // In the phase's self evaluation schema
		ScoreLevel:                  scoreLevelDTO.ScoreLevelGood,
		AuthorCourseParticipationID: uuid.MustParse("01234567-1234-1234-1234-123456789012"), // Must match test middleware
		Type:                        assessmentType.Self,
	}

	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest(http.MethodPost, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
}

func (suite *EvaluationRouterTestSuite) TestCreateOrUpdateEvaluationRejectsCompetencyOutsideEvaluationSchema() {
	requestBody := evaluationDTO.CreateOrUpdateEvaluationRequest{
		CourseParticipationID:       uuid.MustParse("01234567-1234-1234-1234-123456789012"),
		CompetencyID:                uuid.MustParse("c1234567-1234-1234-1234-123456789012"), // In the assessment schema, not the self evaluation schema
		ScoreLevel:                  scoreLevelDTO.ScoreLevelGood,
		AuthorCourseParticipationID: uuid.MustParse("01234567-1234-1234-1234-123456789012"),
		Type:                        assessmentType.Self,
	}

	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest(http.MethodPost, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *EvaluationRouterTestSuite) TestCreateOrUpdateEvaluationInvalidJSON() {
	req, _ := http.NewRequest(http.MethodPost, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var errResp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errResp, "error")
}

func (suite *EvaluationRouterTestSuite) TestCreateOrUpdateEvaluationForbiddenAuthor() {
	requestBody := evaluationDTO.CreateOrUpdateEvaluationRequest{
		CourseParticipationID:       uuid.MustParse("02234567-1234-1234-1234-123456789012"), // Valid from test data
		CompetencyID:                uuid.MustParse("c4234567-1234-1234-1234-123456789012"), // In the phase's self evaluation schema
		ScoreLevel:                  scoreLevelDTO.ScoreLevelGood,
		AuthorCourseParticipationID: uuid.MustParse("02234567-1234-1234-1234-123456789012"), // Different from test middleware user
		Type:                        assessmentType.Self,
	}

	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest(http.MethodPost, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	var errResp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errResp["error"], "Students can only create evaluations as the author")
}

func (suite *EvaluationRouterTestSuite) TestDeleteEvaluation() {
	// Use an existing evaluation ID from our test data
	existingEvaluationID := "e1234567-1234-1234-1234-123456789012"
	req, _ := http.NewRequest(http.MethodDelete, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation/"+existingEvaluationID, nil)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// Regression test for GHSA-h36r-xgjv-r8cx: the phase in the URL must own the evaluation.
func (suite *EvaluationRouterTestSuite) TestDeleteEvaluationWrongPhase() {
	// e2234567 is the current student's own self evaluation in testCoursePhaseID
	evaluationID := "e2234567-1234-1234-1234-123456789012"
	otherPhaseID := "5179d58a-d00d-4fa7-94a5-397bc69fab03"
	req, _ := http.NewRequest(http.MethodDelete, "/assessment/api/course_phase/"+otherPhaseID+"/evaluation/"+evaluationID, nil)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *EvaluationRouterTestSuite) TestDeleteEvaluationNonAuthor() {
	// e3234567 is authored by another student
	evaluationID := "e3234567-1234-1234-1234-123456789012"
	req, _ := http.NewRequest(http.MethodDelete, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation/"+evaluationID, nil)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

// Test that deleting a peer evaluation whose subject is no longer in the author's team answers
// 403 rather than the generic 500 the delete path used to produce.
func (suite *EvaluationRouterTestSuite) TestDeleteEvaluationRejectsTargetOutsideTeam() {
	evaluationID := "e7234567-1234-1234-1234-123456789012"
	req, _ := http.NewRequest(http.MethodDelete, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation/"+evaluationID, nil)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *EvaluationRouterTestSuite) TestInvalidCoursePhaseID() {
	req, _ := http.NewRequest(http.MethodGet, "/assessment/api/course_phase/invalid-uuid/evaluation", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var errResp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errResp, "error")
}

func (suite *EvaluationRouterTestSuite) TestInvalidEvaluationID() {
	req, _ := http.NewRequest(http.MethodDelete, "/assessment/api/course_phase/"+suite.testCoursePhaseID.String()+"/evaluation/invalid-uuid", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var errResp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errResp, "error")
}

var (
	evaluatedTutorID  = uuid.MustParse("0a234567-1234-1234-1234-123456789012")
	tutorCompetencyID = uuid.MustParse("ca234567-1234-1234-1234-123456789012")
	firstStudentID    = uuid.MustParse("01234567-1234-1234-1234-123456789012")
	secondStudentID   = uuid.MustParse("02234567-1234-1234-1234-123456789012")
)

// withTutorEvaluations stores two students' tutor evaluations and feedback about evaluatedTutorID,
// sets the release state, and undoes both when the test ends.
func (suite *EvaluationRouterTestSuite) withTutorEvaluations(resultsReleased bool) {
	ctx := suite.suiteCtx
	conn := suite.evaluationService.conn
	_, err := conn.Exec(ctx, `
		INSERT INTO evaluation (id, course_participation_id, course_phase_id, competency_id, score_level, author_course_participation_id, type)
		VALUES (gen_random_uuid(), $1, $2, $3, 'very_good', $4, 'tutor'),
		       (gen_random_uuid(), $1, $2, $3, 'good', $5, 'tutor')`,
		evaluatedTutorID, suite.testCoursePhaseID, tutorCompetencyID, firstStudentID, secondStudentID)
	suite.Require().NoError(err)
	_, err = conn.Exec(ctx, `
		INSERT INTO feedback_items (id, feedback_type, feedback_text, course_participation_id, course_phase_id, author_course_participation_id, type)
		VALUES (gen_random_uuid(), 'positive', 'Explains things well', $1, $2, $3, 'tutor'),
		       (gen_random_uuid(), 'negative', 'Answers late', $1, $2, $4, 'tutor')`,
		evaluatedTutorID, suite.testCoursePhaseID, firstStudentID, secondStudentID)
	suite.Require().NoError(err)
	_, err = conn.Exec(ctx, `UPDATE course_phase_config SET results_released = $2 WHERE course_phase_id = $1`, suite.testCoursePhaseID, resultsReleased)
	suite.Require().NoError(err)

	suite.T().Cleanup(func() {
		_, _ = conn.Exec(ctx, `DELETE FROM evaluation WHERE course_participation_id = $1`, evaluatedTutorID)
		_, _ = conn.Exec(ctx, `DELETE FROM feedback_items WHERE course_participation_id = $1`, evaluatedTutorID)
		_, _ = conn.Exec(ctx, `UPDATE course_phase_config SET results_released = false WHERE course_phase_id = $1`, suite.testCoursePhaseID)
		suite.ownTutorID, suite.ownTutorFound, suite.ownTutorErr = uuid.Nil, false, nil
	})
}

func (suite *EvaluationRouterTestSuite) getMyTutorResults(coursePhaseID uuid.UUID) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", "/assessment/api/course_phase/"+coursePhaseID.String()+"/evaluation/my-tutor-results", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

func (suite *EvaluationRouterTestSuite) TestMyTutorResultsAreAnonymousAverages() {
	suite.withTutorEvaluations(true)
	suite.ownTutorID, suite.ownTutorFound = evaluatedTutorID, true

	w := suite.getMyTutorResults(suite.testCoursePhaseID)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var results evaluationDTO.TutorEvaluationResults
	assert.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &results))
	assert.Equal(suite.T(), evaluatedTutorID, results.TutorID)
	assert.Equal(suite.T(), 2, results.EvaluatorCount)
	if assert.Len(suite.T(), results.Results, 1) {
		assert.Equal(suite.T(), tutorCompetencyID, results.Results[0].CompetencyID)
		assert.InDelta(suite.T(), 1.5, results.Results[0].AverageScoreNumeric, 0.0001)
	}
	assert.Equal(suite.T(), []evaluationDTO.AnonymousFeedbackItem{
		{FeedbackType: db.FeedbackTypeNegative, FeedbackText: "Answers late"},
		{FeedbackType: db.FeedbackTypePositive, FeedbackText: "Explains things well"},
	}, results.FeedbackItems)
	for _, author := range []uuid.UUID{firstStudentID, secondStudentID} {
		assert.NotContains(suite.T(), w.Body.String(), author.String(), "the response must not name an evaluator")
	}
}

func (suite *EvaluationRouterTestSuite) TestMyTutorResultsNoContentUntilReleased() {
	suite.withTutorEvaluations(false)
	suite.ownTutorID, suite.ownTutorFound = evaluatedTutorID, true

	w := suite.getMyTutorResults(suite.testCoursePhaseID)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
	assert.Empty(suite.T(), w.Body.String())
}

func (suite *EvaluationRouterTestSuite) TestMyTutorResultsNoContentForNonTutor() {
	suite.withTutorEvaluations(true)
	suite.ownTutorFound = false

	w := suite.getMyTutorResults(suite.testCoursePhaseID)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
}

func (suite *EvaluationRouterTestSuite) TestMyTutorResultsNoContentWithoutTutorEvaluation() {
	// The second phase of the dump has the tutor evaluation disabled.
	otherPhaseID := uuid.MustParse("5179d58a-d00d-4fa7-94a5-397bc69fab03")
	conn := suite.evaluationService.conn
	_, err := conn.Exec(suite.suiteCtx, `UPDATE course_phase_config SET results_released = true WHERE course_phase_id = $1`, otherPhaseID)
	suite.Require().NoError(err)
	suite.T().Cleanup(func() {
		_, _ = conn.Exec(suite.suiteCtx, `UPDATE course_phase_config SET results_released = false WHERE course_phase_id = $1`, otherPhaseID)
		suite.ownTutorID, suite.ownTutorFound = uuid.Nil, false
	})
	suite.ownTutorID, suite.ownTutorFound = evaluatedTutorID, true

	w := suite.getMyTutorResults(otherPhaseID)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
}

func (suite *EvaluationRouterTestSuite) TestMyTutorResultsBadGatewayWhenTeamPhaseFails() {
	suite.withTutorEvaluations(true)
	suite.ownTutorErr = errors.New("team phase unavailable")

	w := suite.getMyTutorResults(suite.testCoursePhaseID)

	assert.Equal(suite.T(), http.StatusBadGateway, w.Code)
}
