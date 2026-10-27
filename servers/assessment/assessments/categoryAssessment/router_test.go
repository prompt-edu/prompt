package categoryAssessment

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentSchemas"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/assessmentCompletion"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CategoryAssessmentRouterTestSuite struct {
	suite.Suite
	suiteCtx context.Context
	cleanup  func()
	router   *gin.Engine
	service  *CategoryAssessmentService
}

func (suite *CategoryAssessmentRouterTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.suiteCtx, "../../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../../database_dumps/assessments.sql")
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	coursePhaseConfigService := coursePhaseConfig.NewCoursePhaseConfigService(*testDB.Queries, testDB.Conn, assessmentSchemas.NewAssessmentSchemaService(*testDB.Queries, testDB.Conn))
	suite.service = NewCategoryAssessmentService(*testDB.Queries, testDB.Conn, assessmentCompletion.NewAssessmentCompletionService(*testDB.Queries, testDB.Conn, coursePhaseConfigService))

	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "user@example.com", "1234", "id")
	}
	RegisterRoutes(api, suite.service, coursePhaseConfigService, testMiddleware)
}

func (suite *CategoryAssessmentRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *CategoryAssessmentRouterTestSuite) TestCreateOrUpdateIgnoresBodyCoursePhase() {
	phaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	otherPhaseID := uuid.MustParse("24461b6b-3c3a-4bc6-ba42-69eeb1514da9")
	partID := uuid.New()

	body, _ := json.Marshal(map[string]string{
		"categoryID":            "25f1c984-ba31-4cf2-aa8e-5662721bf44e",
		"coursePhaseID":         otherPhaseID.String(),
		"courseParticipationID": partID.String(),
		"comment":               "Cross-phase comment",
	})
	req, _ := http.NewRequest("POST", "/api/course_phase/"+phaseID.String()+"/category-assessment", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	inPathPhase, err := suite.service.ListCategoryAssessmentsByStudentInPhase(suite.suiteCtx, partID, phaseID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), inPathPhase, 1)

	inBodyPhase, err := suite.service.ListCategoryAssessmentsByStudentInPhase(suite.suiteCtx, partID, otherPhaseID)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), inBodyPhase)
}

func TestCategoryAssessmentRouterTestSuite(t *testing.T) {
	suite.Run(t, new(CategoryAssessmentRouterTestSuite))
}
