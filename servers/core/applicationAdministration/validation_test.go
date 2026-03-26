package applicationAdministration

import (
	"context"
	"encoding/json"
	"log"
	"math/big"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/applicationAdministration/applicationDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ApplicationAdminValidationTestSuite struct {
	suite.Suite
	router                  *gin.Engine
	ctx                     context.Context
	cleanup                 func()
	applicationAdminService ApplicationService
}

func (suite *ApplicationAdminValidationTestSuite) SetupSuite() {
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
}

func (suite *ApplicationAdminValidationTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateUpdateForm_Success() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	updateForm := applicationDTO.UpdateForm{
		DeleteQuestionsText:        []uuid.UUID{uuid.MustParse("a6a04042-95d1-4765-8592-caf9560c8c3c")},
		DeleteQuestionsMultiSelect: []uuid.UUID{uuid.MustParse("65e25b73-ce47-4536-b651-a1632347d733")},
		CreateQuestionsText: []applicationDTO.CreateQuestionText{
			{
				CoursePhaseID: coursePhaseID,
				Title:         "Valid Title",
				AllowedLength: 100,
			},
		},
		CreateQuestionsMultiSelect: []applicationDTO.CreateQuestionMultiSelect{
			{
				CoursePhaseID: coursePhaseID,
				Title:         "Valid MultiSelect",
				MinSelect:     1,
				MaxSelect:     3,
				Options:       []string{"Option1", "Option2"},
			},
		},
	}
	err := validateUpdateForm(suite.ctx, coursePhaseID, updateForm)
	assert.NoError(suite.T(), err)
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateUpdateForm_InvalidDeleteQuestion() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	updateForm := applicationDTO.UpdateForm{
		DeleteQuestionsText: []uuid.UUID{uuid.New()}, // Non-existent question
	}
	err := validateUpdateForm(suite.ctx, coursePhaseID, updateForm)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "question does not belong to this course", err.Error())
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateUpdateForm_InvalidCreateTextQuestion() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	updateForm := applicationDTO.UpdateForm{
		CreateQuestionsText: []applicationDTO.CreateQuestionText{
			{
				CoursePhaseID: uuid.New(),
				Title:         "",
				AllowedLength: 0,
			},
		},
	}
	err := validateUpdateForm(suite.ctx, coursePhaseID, updateForm)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "course phase id is not correct", err.Error())
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateUpdateForm_InvalidCreateMultiSelect() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	updateForm := applicationDTO.UpdateForm{
		CreateQuestionsMultiSelect: []applicationDTO.CreateQuestionMultiSelect{
			{
				CoursePhaseID: coursePhaseID,
				Title:         "Invalid MultiSelect",
				MinSelect:     -1,
				MaxSelect:     0,
				Options:       []string{},
			},
		},
	}
	err := validateUpdateForm(suite.ctx, coursePhaseID, updateForm)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "minimum selection must be at least 0", err.Error())
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateQuestionText_EmptyTitle() {
	err := validateQuestionText("", "", 100, pgtype.Bool{Bool: true, Valid: true}, pgtype.Text{String: "test", Valid: true})
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "title is required", err.Error())
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateQuestionText_InvalidRegex() {
	err := validateQuestionText("Valid Title", "[", 100, pgtype.Bool{Bool: true, Valid: true}, pgtype.Text{String: "test", Valid: true})
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid regex pattern")
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateQuestionMultiSelect_EmptyTitle() {
	err := validateQuestionMultiSelect("", 1, 3, []string{"Option1", "Option2"}, pgtype.Bool{Bool: true, Valid: true}, pgtype.Text{String: "test", Valid: true})
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "title is required", err.Error())
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateQuestionMultiSelect_EmptyOptions() {
	err := validateQuestionMultiSelect("Valid Title", 1, 3, []string{}, pgtype.Bool{Bool: true, Valid: true}, pgtype.Text{String: "test", Valid: true})
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "options cannot be empty", err.Error())
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateQuestionMultiSelect_MinSelectNegative() {
	err := validateQuestionMultiSelect("Valid Title", -1, 3, []string{"Option1", "Option2"}, pgtype.Bool{Bool: true, Valid: true}, pgtype.Text{String: "test", Valid: true})
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "minimum selection must be at least 0", err.Error())
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateQuestionMultiSelect_MaxSelectLessThanOne() {
	err := validateQuestionMultiSelect("Valid Title", 0, 0, []string{"Option1", "Option2"}, pgtype.Bool{Bool: true, Valid: true}, pgtype.Text{String: "test", Valid: true})
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "maximum selection must be at least 1", err.Error())
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateApplication_InvalidStudent() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	application := applicationDTO.PostApplication{
		Student: studentDTO.CreateStudent{
			ID:                   uuid.New(),
			HasUniversityAccount: false,
		},
		AnswersText:        []applicationDTO.CreateAnswerText{},
		AnswersMultiSelect: []applicationDTO.CreateAnswerMultiSelect{},
	}

	err := validateApplication(suite.ctx, coursePhaseID, application, false)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "invalid student", err.Error())
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateApplication_InvalidTextAnswers() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	application := applicationDTO.PostApplication{
		Student: studentDTO.CreateStudent{
			ID: uuid.New(),
			// Valid student details
			FirstName:            "John",
			LastName:             "Doe",
			Email:                "test@test.de",
			HasUniversityAccount: false,
			Nationality:          "DE",
			CurrentSemester:      pgtype.Int4{Valid: true, Int32: 1},
			StudyProgram:         "Computer Science",
			StudyDegree:          "bachelor",
		},
		AnswersText: []applicationDTO.CreateAnswerText{
			{
				ApplicationQuestionID: uuid.New(), // Non-existent question ID
				Answer:                "Invalid Answer",
			},
		},
	}

	err := validateApplication(suite.ctx, coursePhaseID, application, false)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "required question")
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateTextAnswers_ExceedsAllowedLength() {
	textQuestions := []db.ApplicationQuestionText{
		{
			ID:            uuid.New(),
			AllowedLength: pgtype.Int4{Int32: 10, Valid: true},
			IsRequired:    pgtype.Bool{Bool: true, Valid: true},
		},
	}
	textAnswers := []applicationDTO.CreateAnswerText{
		{
			ApplicationQuestionID: textQuestions[0].ID,
			Answer:                "This answer is way too long",
		},
	}

	err := validateTextAnswers(textQuestions, textAnswers)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "exceeds allowed length")
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateMultiSelectAnswers_InvalidSelection() {
	multiSelectQuestions := []db.ApplicationQuestionMultiSelect{
		{
			ID:         uuid.New(),
			MinSelect:  pgtype.Int4{Int32: 1, Valid: true},
			MaxSelect:  pgtype.Int4{Int32: 3, Valid: true},
			IsRequired: pgtype.Bool{Bool: true, Valid: true},
			Options:    []string{"Option1", "Option2", "Option3"},
		},
	}
	multiSelectAnswers := []applicationDTO.CreateAnswerMultiSelect{
		{
			ApplicationQuestionID: multiSelectQuestions[0].ID,
			Answer:                []string{"InvalidOption"},
		},
	}

	err := validateMultiSelectAnswers(multiSelectQuestions, multiSelectAnswers)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid selection")
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateMultiSelectAnswers_MissingRequiredAnswer() {
	multiSelectQuestions := []db.ApplicationQuestionMultiSelect{
		{
			ID:         uuid.New(),
			MinSelect:  pgtype.Int4{Int32: 1, Valid: true},
			IsRequired: pgtype.Bool{Bool: true, Valid: true},
			Options:    []string{"Option1", "Option2", "Option3"},
		},
	}
	multiSelectAnswers := []applicationDTO.CreateAnswerMultiSelect{}

	err := validateMultiSelectAnswers(multiSelectQuestions, multiSelectAnswers)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "required question")
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateMultiSelectAnswers_SelectionOutOfRange() {
	multiSelectQuestions := []db.ApplicationQuestionMultiSelect{
		{
			ID:        uuid.New(),
			MinSelect: pgtype.Int4{Int32: 1, Valid: true},
			MaxSelect: pgtype.Int4{Int32: 3, Valid: true},
			Options:   []string{"Option1", "Option2", "Option3"},
		},
	}
	multiSelectAnswers := []applicationDTO.CreateAnswerMultiSelect{
		{
			ApplicationQuestionID: multiSelectQuestions[0].ID,
			Answer:                []string{"Option1", "Option2", "Option3", "Option4"}, // Exceeds MaxSelect
		},
	}

	err := validateMultiSelectAnswers(multiSelectQuestions, multiSelectAnswers)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "does not meet selection requirements")
}

func TestValidateTextAnswers_InvalidQuestionID(t *testing.T) {
	textQuestions := []db.ApplicationQuestionText{
		{
			ID: uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02"),
			ValidationRegex: pgtype.Text{
				String: `^[a-zA-Z\s]+$`,
				Valid:  true,
			},
			AllowedLength: pgtype.Int4{
				Int32: 100,
				Valid: true,
			},
			IsRequired: pgtype.Bool{
				Bool:  true,
				Valid: true,
			},
		},
	}

	textAnswers := []applicationDTO.CreateAnswerText{
		{
			ApplicationQuestionID: uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02"),
			Answer:                "This is a valid answer",
		},
		{
			ApplicationQuestionID: uuid.MustParse("d1e74f8b-9f7f-4b87-94a5-1234567890ab"), // Invalid ID
			Answer:                "This is a invalid answer",
		},
	}

	err := validateTextAnswers(textQuestions, textAnswers)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "answer to question d1e74f8b-9f7f-4b87-94a5-1234567890ab does not belong to this course")
}

func TestValidateTextAnswers_RegexMismatch(t *testing.T) {
	textQuestions := []db.ApplicationQuestionText{
		{
			ID: uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02"),
			ValidationRegex: pgtype.Text{
				String: `^[a-zA-Z\s]+$`,
				Valid:  true,
			},
			AllowedLength: pgtype.Int4{
				Int32: 100,
				Valid: true,
			},
			IsRequired: pgtype.Bool{
				Bool:  true,
				Valid: true,
			},
		},
	}

	textAnswers := []applicationDTO.CreateAnswerText{
		{
			ApplicationQuestionID: uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02"), // Valid ID
			Answer:                "123InvalidRegexAnswer!",                               // Regex mismatch
		},
	}

	err := validateTextAnswers(textQuestions, textAnswers)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "answer to question 4179d58a-d00d-4fa7-94a5-397bc69fab02 does not match validation regex")
}

func TestValidateMultiSelectAnswers_InvalidQuestionID(t *testing.T) {
	multiSelectQuestions := []db.ApplicationQuestionMultiSelect{
		{
			ID: uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02"),
			MinSelect: pgtype.Int4{
				Int32: 1,
				Valid: true,
			},
			MaxSelect: pgtype.Int4{
				Int32: 3,
				Valid: true,
			},
			Options: []string{"Option1", "Option2", "Option3"},
			IsRequired: pgtype.Bool{
				Bool:  true,
				Valid: true,
			},
		},
	}

	multiSelectAnswers := []applicationDTO.CreateAnswerMultiSelect{
		{
			ApplicationQuestionID: uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02"),
			Answer:                []string{"Option1", "Option2"},
		},
		{
			ApplicationQuestionID: uuid.MustParse("d1e74f8b-9f7f-4b87-94a5-1234567890ab"), // Invalid ID
			Answer:                []string{"Option1", "Option2"},
		},
	}

	err := validateMultiSelectAnswers(multiSelectQuestions, multiSelectAnswers)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "answer to question d1e74f8b-9f7f-4b87-94a5-1234567890ab does not belong to this course")
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateUpdateAssessment_Success() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	courseParticipationID := uuid.MustParse("82d7efae-d545-4cc5-9b94-5d0ee1e50d25")
	passStatus := db.PassStatusFailed
	assessment := applicationDTO.PutAssessment{
		PassStatus: &passStatus,
	}

	err := validateUpdateAssessment(suite.ctx, coursePhaseID, courseParticipationID, assessment)
	assert.NoError(suite.T(), err)
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateUpdateAssessment_NotAssessmentPhase() {
	coursePhaseID := uuid.MustParse("7062236a-e290-487c-be41-29b24e0afc64")
	courseParticipationID := uuid.MustParse("82d7efae-d545-4cc5-9b94-5d0ee1e50d25")
	assessment := applicationDTO.PutAssessment{}

	err := validateUpdateAssessment(suite.ctx, coursePhaseID, courseParticipationID, assessment)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "course phase is not an assessment phase", err.Error())
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateUpdateAssessment_InvalidMetaDataKey() {
	coursePhaseID := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	courseParticipationID := uuid.MustParse("82d7efae-d545-4cc5-9b94-5d0ee1e50d25")

	jsonData := `{"invalid_key": "value"}`
	var restrictedData meta.MetaData
	err := json.Unmarshal([]byte(jsonData), &restrictedData)
	assert.NoError(suite.T(), err)

	assessment := applicationDTO.PutAssessment{
		RestrictedData: restrictedData,
	}

	err = validateUpdateAssessment(suite.ctx, coursePhaseID, courseParticipationID, assessment)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "invalid meta data key - not allowed to update other meta data", err.Error())
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateAdditionalScore_Success() {
	validScore := applicationDTO.AdditionalScoreUpload{
		Name: "ValidScore",
		Key:  "valid_key",
		Scores: []applicationDTO.IndividualScore{
			{
				CourseParticipationID: uuid.New(),
				Score: pgtype.Numeric{
					Int:   big.NewInt(10),
					Valid: true,
				},
			},
			{
				CourseParticipationID: uuid.New(),
				Score: pgtype.Numeric{
					Int:   big.NewInt(10),
					Valid: true,
				},
			},
		},
	}

	err := validateAdditionalScore(validScore)
	assert.NoError(suite.T(), err)
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateAdditionalScore_EmptyName() {
	invalidScore := applicationDTO.AdditionalScoreUpload{
		Name: "",
		Key:  "invalid_key",
		Scores: []applicationDTO.IndividualScore{
			{
				CourseParticipationID: uuid.New(),
				Score: pgtype.Numeric{
					Int:   big.NewInt(10),
					Valid: true,
				},
			},
		},
	}

	err := validateAdditionalScore(invalidScore)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "name cannot be empty", err.Error())
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateAdditionalScore_NegativeScore() {
	invalidScore := applicationDTO.AdditionalScoreUpload{
		Name: "NegativeScore",
		Key:  "negative_score",
		Scores: []applicationDTO.IndividualScore{
			{
				CourseParticipationID: uuid.New(),
				Score: pgtype.Numeric{
					Int:   big.NewInt(-10),
					Valid: true,
				},
			},
		},
	}

	err := validateAdditionalScore(invalidScore)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "scores must be positive", err.Error())
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateAdditionalScore_InvalidScoreValue() {
	invalidScore := applicationDTO.AdditionalScoreUpload{
		Name: "InvalidScoreValue",
		Key:  "invalid_key",
		Scores: []applicationDTO.IndividualScore{
			{
				CourseParticipationID: uuid.New(),
				Score: pgtype.Numeric{
					Valid: false, // Invalid value
				},
			},
		},
	}

	err := validateAdditionalScore(invalidScore)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "failed to parse score for entry", err.Error())
}

func (suite *ApplicationAdminValidationTestSuite) TestValidateAdditionalScore_MixedValidAndInvalidScores() {
	invalidScore := applicationDTO.AdditionalScoreUpload{
		Name: "MixedScores",
		Key:  "mixed_scores",
		Scores: []applicationDTO.IndividualScore{
			{
				CourseParticipationID: uuid.New(),
				Score: pgtype.Numeric{
					Int:   big.NewInt(10),
					Valid: true,
				},
			},
			{
				CourseParticipationID: uuid.New(),
				Score: pgtype.Numeric{
					Int:   big.NewInt(-10),
					Valid: true,
				},
			},
		},
	}

	err := validateAdditionalScore(invalidScore)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "scores must be positive", err.Error())
}

func TestValidateUpdateFormSuite(t *testing.T) {
	suite.Run(t, new(ApplicationAdminValidationTestSuite))
}
