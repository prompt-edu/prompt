package applicationAdministration

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/applicationAdministration/applicationDTO"
	"github.com/prompt-edu/prompt/servers/core/course/courseParticipation"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseParticipation"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/mailing"
	"github.com/prompt-edu/prompt/servers/core/student"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ApplicationAdminRouterTestSuite struct {
	suite.Suite
	router                  *gin.Engine
	ctx                     context.Context
	cleanup                 func()
	applicationAdminService ApplicationService
}

var (
	requiredFileUploadQuestionID = uuid.MustParse("b1b04042-95d1-4765-8592-caf9560c8c3d")
	seededUploadFileID           = uuid.MustParse("d3d04042-95d1-4765-8592-caf9560c8c3f")
)

func (suite *ApplicationAdminRouterTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Set up PostgreSQL container
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/application_administration.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}

	suite.cleanup = cleanup
	suite.applicationAdminService = ApplicationService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}

	ApplicationServiceSingleton = &suite.applicationAdminService
	suite.router = gin.Default()
	api := suite.router.Group("/api")
	testMiddleware := func() gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail([]string{"PROMPT_Admin", "ios24245-iPraktikum-Lecturer"}, "existingstudent@example.com", "03711111", "ab12cde")
	}
	setupApplicationRouter(api, testMiddleware, testMiddleware, sdkTestUtils.MockPermissionMiddleware)
	student.InitStudentModule(suite.router.Group("/api"), *testDB.Queries, testDB.Conn)
	courseParticipation.InitCourseParticipationModule(suite.router.Group("/api"), *testDB.Queries, testDB.Conn)
	coursePhaseParticipation.InitCoursePhaseParticipationModule(suite.router.Group("/api"), *testDB.Queries, testDB.Conn)
	mailing.InitMailingModule(api, *testDB.Queries, testDB.Conn, "localhost", "25", "", "", "Test-Email-Sender", "test@test.de", "localhost")
}

func (suite *ApplicationAdminRouterTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *ApplicationAdminRouterTestSuite) TestGetApplicationFormEndpoint_Success() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	req := httptest.NewRequest(http.MethodGet, "/api/applications/"+coursePhaseID+"/form", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	var form applicationDTO.Form
	err := json.Unmarshal(resp.Body.Bytes(), &form)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), form)
	assert.NotEmpty(suite.T(), form.QuestionsText)
	assert.NotEmpty(suite.T(), form.QuestionsMultiSelect)
}

func (suite *ApplicationAdminRouterTestSuite) TestUpdateApplicationFormEndpoint_Success() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	updateForm := applicationDTO.UpdateForm{
		DeleteQuestionsText:        []uuid.UUID{uuid.MustParse("a6a04042-95d1-4765-8592-caf9560c8c3c")},
		DeleteQuestionsMultiSelect: []uuid.UUID{uuid.MustParse("65e25b73-ce47-4536-b651-a1632347d733")},
		CreateQuestionsText: []applicationDTO.CreateQuestionText{
			{
				CoursePhaseID: uuid.MustParse(coursePhaseID),
				Title:         "New Motivation",
				AllowedLength: 300,
			},
		},
		CreateQuestionsMultiSelect: []applicationDTO.CreateQuestionMultiSelect{
			{
				CoursePhaseID: uuid.MustParse(coursePhaseID),
				Title:         "New Devices",
				MinSelect:     1,
				MaxSelect:     5,
				Options:       []string{"Option1", "Option2"},
			},
		},
	}

	jsonBody, err := json.Marshal(updateForm)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPut, "/api/applications/"+coursePhaseID+"/form", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "application form updated", responseBody["message"])

	// Verify updates via GET
	getReq := httptest.NewRequest(http.MethodGet, "/api/applications/"+coursePhaseID+"/form", nil)
	getResp := httptest.NewRecorder()
	suite.router.ServeHTTP(getResp, getReq)

	assert.Equal(suite.T(), http.StatusOK, getResp.Code)
	var updatedForm applicationDTO.Form
	err = json.Unmarshal(getResp.Body.Bytes(), &updatedForm)
	assert.NoError(suite.T(), err)

	// Verify QuestionsText
	assert.Equal(suite.T(), 2, len(updatedForm.QuestionsText))
	for _, question := range updatedForm.QuestionsText {
		switch question.Title {
		case "New Motivation":
			assert.Equal(suite.T(), 300, question.AllowedLength)
		case "Expierence":
			assert.Equal(suite.T(), 500, question.AllowedLength)
		default:
			suite.T().Errorf("Unexpected question title: %s", question.Title)
		}
	}

	// Verify QuestionsMultiSelect
	assert.Equal(suite.T(), 2, len(updatedForm.QuestionsMultiSelect))
	for _, question := range updatedForm.QuestionsMultiSelect {
		switch question.Title {
		case "New Devices":
			assert.ElementsMatch(suite.T(), []string{"Option1", "Option2"}, question.Options)
		case "Available Devices":
			assert.ElementsMatch(suite.T(), []string{"iPhone", "iPad", "MacBook", "Vision"}, question.Options)
		default:
			suite.T().Errorf("Unexpected question title: %s", question.Title)
		}
	}
}

func (suite *ApplicationAdminRouterTestSuite) TestGetAllOpenApplicationsEndpoint_Success() {
	req := httptest.NewRequest(http.MethodGet, "/api/apply", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var openApplications []applicationDTO.OpenApplication
	err := json.Unmarshal(resp.Body.Bytes(), &openApplications)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), openApplications)

	for _, app := range openApplications {
		assert.NotEmpty(suite.T(), app.CourseName)
		assert.NotEmpty(suite.T(), app.CoursePhaseID)
	}
}

func (suite *ApplicationAdminRouterTestSuite) TestGetApplicationFormWithCourseDetailsEndpoint_Success() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	req := httptest.NewRequest(http.MethodGet, "/api/apply/"+coursePhaseID, nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var formWithDetails applicationDTO.FormWithDetails
	err := json.Unmarshal(resp.Body.Bytes(), &formWithDetails)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), formWithDetails)
	assert.NotEmpty(suite.T(), formWithDetails.QuestionsText)
	assert.NotEmpty(suite.T(), formWithDetails.QuestionsMultiSelect)
}

func (suite *ApplicationAdminRouterTestSuite) TestPostApplicationExternEndpoint_Success() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	application := applicationDTO.PostApplication{
		Student: studentDTO.CreateStudent{
			FirstName:       "John",
			LastName:        "Doe",
			Email:           "johndoe@example.com",
			Gender:          db.GenderDiverse,
			Nationality:     "DE",
			CurrentSemester: pgtype.Int4{Valid: true, Int32: 1},
			StudyProgram:    "Computer Science",
			StudyDegree:     "bachelor",
		},
		AnswersText: []applicationDTO.CreateAnswerText{
			{
				ApplicationQuestionID: uuid.MustParse("a6a04042-95d1-4765-8592-caf9560c8c3c"),
				Answer:                "This is a valid answer.",
			},
		},
		AnswersMultiSelect: []applicationDTO.CreateAnswerMultiSelect{
			{
				ApplicationQuestionID: uuid.MustParse("383a9590-fba2-4e6b-a32b-88895d55fb9b"),
				Answer:                []string{"MacBook"},
			},
		},
		AnswersFileUpload: []applicationDTO.CreateAnswerFileUpload{
			{
				ApplicationQuestionID: requiredFileUploadQuestionID,
				FileID:                seededUploadFileID,
			},
		},
	}

	jsonBody, err := json.Marshal(application)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/api/apply/"+coursePhaseID, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "application posted", responseBody["message"])
}

func (suite *ApplicationAdminRouterTestSuite) TestGetApplicationAuthenticatedEndpoint_Success() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	req := httptest.NewRequest(http.MethodGet, "/api/apply/authenticated/"+coursePhaseID, nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var application applicationDTO.Application
	err := json.Unmarshal(resp.Body.Bytes(), &application)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), application)
	assert.Equal(suite.T(), applicationDTO.StatusNotApplied, application.Status)
}

func (suite *ApplicationAdminRouterTestSuite) TestPostApplicationAuthenticatedEndpoint_Success() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	application := applicationDTO.PostApplication{
		Student: studentDTO.CreateStudent{
			FirstName:            "John",
			LastName:             "Doe",
			Email:                "existingstudent@example.com",
			Gender:               db.GenderDiverse,
			HasUniversityAccount: true,
			MatriculationNumber:  "03711111",
			UniversityLogin:      "ab12cde",
			Nationality:          "DE",
			CurrentSemester:      pgtype.Int4{Valid: true, Int32: 1},
			StudyProgram:         "Computer Science",
			StudyDegree:          "bachelor",
		},
		AnswersText: []applicationDTO.CreateAnswerText{
			{
				ApplicationQuestionID: uuid.MustParse("a6a04042-95d1-4765-8592-caf9560c8c3c"),
				Answer:                "Valid answer.",
			},
		},
		AnswersMultiSelect: []applicationDTO.CreateAnswerMultiSelect{
			{
				ApplicationQuestionID: uuid.MustParse("383a9590-fba2-4e6b-a32b-88895d55fb9b"),
				Answer:                []string{"MacBook"},
			},
		},
		AnswersFileUpload: []applicationDTO.CreateAnswerFileUpload{
			{
				ApplicationQuestionID: requiredFileUploadQuestionID,
				FileID:                seededUploadFileID,
			},
		},
	}

	jsonBody, err := json.Marshal(application)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/api/apply/authenticated/"+coursePhaseID, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "application posted", responseBody["message"])
}

func (suite *ApplicationAdminRouterTestSuite) TestPostApplicationExternEndpoint_AlreadyApplied() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	application := applicationDTO.PostApplication{
		Student: studentDTO.CreateStudent{
			FirstName:       "Jane",
			LastName:        "Doe",
			Email:           "janedoe@example.com",
			Gender:          db.GenderFemale,
			Nationality:     "DE",
			CurrentSemester: pgtype.Int4{Valid: true, Int32: 1},
			StudyProgram:    "Computer Science",
			StudyDegree:     "bachelor",
		},
		AnswersText: []applicationDTO.CreateAnswerText{
			{
				ApplicationQuestionID: uuid.MustParse("a6a04042-95d1-4765-8592-caf9560c8c3c"),
				Answer:                "Valid answer.",
			},
		},
		AnswersMultiSelect: []applicationDTO.CreateAnswerMultiSelect{
			{
				ApplicationQuestionID: uuid.MustParse("383a9590-fba2-4e6b-a32b-88895d55fb9b"),
				Answer:                []string{"MacBook"},
			},
		},
		AnswersFileUpload: []applicationDTO.CreateAnswerFileUpload{
			{
				ApplicationQuestionID: requiredFileUploadQuestionID,
				FileID:                seededUploadFileID,
			},
		},
	}

	// Apply once
	jsonBody, err := json.Marshal(application)
	assert.NoError(suite.T(), err)
	req := httptest.NewRequest(http.MethodPost, "/api/apply/"+coursePhaseID, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusCreated, resp.Code)

	// Apply again
	req = httptest.NewRequest(http.MethodPost, "/api/apply/"+coursePhaseID, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()
	suite.router.ServeHTTP(resp, req)
	assert.Equal(suite.T(), http.StatusMethodNotAllowed, resp.Code)

	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "already applied", responseBody["error"])
}

func (suite *ApplicationAdminRouterTestSuite) TestPostApplicationManualEndpoint_UniversityStudent_Success() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	application := applicationDTO.PostApplication{
		Student: studentDTO.CreateStudent{
			FirstName:            "New",
			LastName:             "UniversityStudent",
			Email:                "newunistudent@tum.de",
			Gender:               db.GenderMale,
			HasUniversityAccount: true,
			MatriculationNumber:  "03799999",
			UniversityLogin:      "zz99zzz",
			Nationality:          "DE",
			CurrentSemester:      pgtype.Int4{Valid: true, Int32: 3},
			StudyProgram:         "Computer Science",
			StudyDegree:          "bachelor",
		},
		AnswersText: []applicationDTO.CreateAnswerText{
			{
				ApplicationQuestionID: uuid.MustParse("a6a04042-95d1-4765-8592-caf9560c8c3c"),
				Answer:                "Valid motivation answer.",
			},
		},
		AnswersMultiSelect: []applicationDTO.CreateAnswerMultiSelect{
			{
				ApplicationQuestionID: uuid.MustParse("383a9590-fba2-4e6b-a32b-88895d55fb9b"),
				Answer:                []string{"MacBook"},
			},
		},
		AnswersFileUpload: []applicationDTO.CreateAnswerFileUpload{
			{
				ApplicationQuestionID: requiredFileUploadQuestionID,
				FileID:                seededUploadFileID,
			},
		},
	}

	jsonBody, err := json.Marshal(application)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/api/applications/"+coursePhaseID, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	suite.T().Logf("Response code: %d", resp.Code)
	suite.T().Logf("Response body: %s", resp.Body.String())

	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
	var responseBody map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &responseBody)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "application posted", responseBody["message"])
}

func TestApplicationAdminRouterTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationAdminRouterTestSuite))
}
