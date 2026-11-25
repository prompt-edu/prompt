package coursePhase

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Course phases of full_db.sql: the Application phase passes its answers to the Interview
// phase, which passes only its score on to the Matching phase.
var (
	applicationPhaseID = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	interviewPhaseID   = uuid.MustParse("2b1a55ad-8b1d-453f-b2b4-2373ecb35bc1")
	matchingPhaseID    = uuid.MustParse("7ffffd38-2454-4c67-821d-5692d8086e6c")
)

type ExportedApplicationQuestionsTestSuite struct {
	suite.Suite
	ctx                context.Context
	cleanup            func()
	coursePhaseService *CoursePhaseService
	router             *gin.Engine
}

func (suite *ExportedApplicationQuestionsTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	testDB, cleanup, err := sdkTestUtils.SetupTestDBWithMigrations(suite.ctx, "../db/migration", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) }, "../database_dumps/full_db.sql")
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}

	suite.cleanup = cleanup
	suite.coursePhaseService = NewCoursePhaseService(*testDB.Queries, testDB.Conn, resolution.NewResolutionService("localhost:8080"))
	suite.router = setupRouter(suite.coursePhaseService)
}

func (suite *ExportedApplicationQuestionsTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *ExportedApplicationQuestionsTestSuite) TestListsTheQuestionsExportedToThePhase() {
	exported, err := suite.coursePhaseService.GetExportedApplicationQuestions(suite.ctx, interviewPhaseID)
	assert.NoError(suite.T(), err)

	if assert.NotNil(suite.T(), exported.ApplicationCoursePhaseID) {
		assert.Equal(suite.T(), applicationPhaseID, *exported.ApplicationCoursePhaseID)
	}

	// Only questions that are accessible for other phases, in display order. Motivation
	// (order 3) is answered but not exported.
	expected := []coursePhaseDTO.ExportedApplicationQuestion{
		{QuestionID: uuid.MustParse("fc8bda6d-280e-4a5e-9ebd-4bd8b68aab75"), Title: "Expierenced7", Type: "text", OrderNum: 1},
		{QuestionID: uuid.MustParse("65e25b73-ce47-4536-b651-a1632347d733"), Title: "Taken Courses", Type: "multiselect", OrderNum: 4},
		{QuestionID: uuid.MustParse("326d1c5c-992f-427d-8d99-e119149b46c0"), Title: "Data Rights Acceptance", Type: "multiselect", OrderNum: 5},
		{QuestionID: uuid.MustParse("8c5558c7-a1e5-42bc-89c3-d5b8b427396b"), Title: "New MultiSelect Test", Type: "multiselect", OrderNum: 7},
		{QuestionID: uuid.MustParse("35687df9-d180-4b37-b71d-69ca419fb8e9"), Title: "New Test", Type: "text", OrderNum: 10},
	}
	assert.Equal(suite.T(), expected, exported.Questions)
}

func (suite *ExportedApplicationQuestionsTestSuite) TestPhaseWithoutApplicationAnswersGetsNoQuestions() {
	// Matching receives a score, but no application answers.
	for _, coursePhaseID := range []uuid.UUID{matchingPhaseID, applicationPhaseID} {
		exported, err := suite.coursePhaseService.GetExportedApplicationQuestions(suite.ctx, coursePhaseID)
		assert.NoError(suite.T(), err)
		assert.Nil(suite.T(), exported.ApplicationCoursePhaseID)
		assert.NotNil(suite.T(), exported.Questions, "Questions must serialize as an empty list, not null")
		assert.Empty(suite.T(), exported.Questions)
	}
}

func (suite *ExportedApplicationQuestionsTestSuite) TestRoute() {
	req := httptest.NewRequest(http.MethodGet, "/api/course_phases/"+interviewPhaseID.String()+"/exported_application_questions", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var body struct {
		ApplicationCoursePhaseID string `json:"applicationCoursePhaseID"`
		Questions                []struct {
			QuestionID string `json:"questionID"`
			Title      string `json:"title"`
			Type       string `json:"type"`
			OrderNum   int    `json:"orderNum"`
		} `json:"questions"`
	}
	assert.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(suite.T(), applicationPhaseID.String(), body.ApplicationCoursePhaseID)
	if assert.Len(suite.T(), body.Questions, 5) {
		assert.Equal(suite.T(), "fc8bda6d-280e-4a5e-9ebd-4bd8b68aab75", body.Questions[0].QuestionID)
		assert.Equal(suite.T(), "Expierenced7", body.Questions[0].Title)
		assert.Equal(suite.T(), "text", body.Questions[0].Type)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/course_phases/not-a-uuid/exported_application_questions", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func TestExportedApplicationQuestionsTestSuite(t *testing.T) {
	suite.Run(t, new(ExportedApplicationQuestionsTestSuite))
}
