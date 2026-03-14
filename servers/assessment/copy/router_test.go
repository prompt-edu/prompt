package copy

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
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CopyRouterTestSuite struct {
	suite.Suite
	router      *gin.Engine
	suiteCtx    context.Context
	cleanup     func()
	copyService CopyService
}

func (suite *CopyRouterTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/coursePhaseConfig.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.copyService = CopyService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	CopyServiceSingleton = &suite.copyService
	suite.router = gin.Default()
	api := suite.router.Group("/api")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "lecturer@example.com", "03711111", "ab12cde")
	}
	setupCopyRouter(api, testMiddleware)
}

func (suite *CopyRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *CopyRouterTestSuite) TestCopyEndpoint_Success() {
	// Create source course phase config
	sourceCoursePhaseID := uuid.New()
	targetCoursePhaseID := uuid.New()
	assessmentSchemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	selfEvalSchemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	peerEvalSchemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
	tutorEvalSchemaID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440003")

	// Create valid timestamps for NOT NULL fields
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}

	// Create source config
	err := suite.copyService.queries.CreateOrUpdateCoursePhaseConfig(suite.suiteCtx, db.CreateOrUpdateCoursePhaseConfigParams{
		AssessmentSchemaID:       assessmentSchemaID,
		CoursePhaseID:            sourceCoursePhaseID,
		Start:                    now,
		Deadline:                 pgtype.Timestamptz{Valid: false},
		SelfEvaluationEnabled:    true,
		SelfEvaluationSchema:     selfEvalSchemaID,
		SelfEvaluationStart:      now,
		SelfEvaluationDeadline:   now,
		PeerEvaluationEnabled:    false,
		PeerEvaluationSchema:     peerEvalSchemaID,
		PeerEvaluationStart:      now,
		PeerEvaluationDeadline:   now,
		TutorEvaluationEnabled:   false,
		TutorEvaluationSchema:    tutorEvalSchemaID,
		TutorEvaluationStart:     pgtype.Timestamptz{Valid: false},
		TutorEvaluationDeadline:  pgtype.Timestamptz{Valid: false},
		EvaluationResultsVisible: true,
		GradeSuggestionVisible:   pgtype.Bool{Bool: true, Valid: true},
		ActionItemsVisible:       pgtype.Bool{Bool: true, Valid: true},
		GradingSheetVisible:      pgtype.Bool{Bool: false, Valid: true},
	})
	assert.NoError(suite.T(), err)

	// Create request
	copyReq := promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: sourceCoursePhaseID,
		TargetCoursePhaseID: targetCoursePhaseID,
	}
	body, _ := json.Marshal(copyReq)
	req, _ := http.NewRequest("POST", "/api/copy", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	// Verify the target config was created
	targetConfig, err := suite.copyService.queries.GetCoursePhaseConfig(suite.suiteCtx, targetCoursePhaseID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), assessmentSchemaID, targetConfig.AssessmentSchemaID)
	assert.Equal(suite.T(), targetCoursePhaseID, targetConfig.CoursePhaseID)
	assert.Equal(suite.T(), true, targetConfig.SelfEvaluationEnabled)
	assert.Equal(suite.T(), selfEvalSchemaID, targetConfig.SelfEvaluationSchema)
}

func (suite *CopyRouterTestSuite) TestCopyEndpoint_InvalidJSON() {
	req, _ := http.NewRequest("POST", "/api/copy", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *CopyRouterTestSuite) TestCopyEndpoint_SameSourceAndTarget() {
	coursePhaseID := uuid.New()

	copyReq := promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: coursePhaseID,
		TargetCoursePhaseID: coursePhaseID,
	}
	body, _ := json.Marshal(copyReq)
	req, _ := http.NewRequest("POST", "/api/copy", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code, "Should succeed when source and target are the same")
}

func (suite *CopyRouterTestSuite) TestCopyEndpoint_NonExistentSource() {
	nonExistentSourceID := uuid.New()
	targetCoursePhaseID := uuid.New()

	copyReq := promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: nonExistentSourceID,
		TargetCoursePhaseID: targetCoursePhaseID,
	}
	body, _ := json.Marshal(copyReq)
	req, _ := http.NewRequest("POST", "/api/copy", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, resp.Code, "Should fail when source doesn't exist")
}

func TestCopyRouterTestSuite(t *testing.T) {
	suite.Run(t, new(CopyRouterTestSuite))
}
