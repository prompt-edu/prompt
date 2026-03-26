package student

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
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RouterTestSuite struct {
	suite.Suite
	router         *gin.Engine
	ctx            context.Context
	cleanup        func()
	studentService StudentService
}

func (suite *RouterTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Set up the test database
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/student_test.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}

	suite.cleanup = cleanup
	suite.studentService = StudentService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}

	StudentServiceSingleton = &suite.studentService

	suite.router = setupRouter()
}

func (suite *RouterTestSuite) TearDownSuite() {
	suite.cleanup()
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	authMiddleware := func() gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddleware([]string{"PROMPT_Admin"})
	}
	permissionMiddleware := sdkTestUtils.MockPermissionMiddleware
	setupStudentRouter(api, authMiddleware, permissionMiddleware)
	return router
}

func (suite *RouterTestSuite) TestRouterGetAllStudents() {
	req, _ := http.NewRequest("GET", "/api/students/", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var students []studentDTO.Student
	err := json.Unmarshal(w.Body.Bytes(), &students)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(students), 0, "Expected at least one student in the initial data")
}

func (suite *RouterTestSuite) TestRouterGetStudentByID() {
	var expectedID uuid.UUID

	// Create a student first
	newStudent := studentDTO.CreateStudent{
		FirstName:            "Alice",
		LastName:             "Smith",
		Email:                "alice.smith@example.com",
		HasUniversityAccount: true,
		MatriculationNumber:  "01234567",
		UniversityLogin:      "as12xyz",
		Gender:               "female",
		Nationality:          "DE",
		CurrentSemester:      pgtype.Int4{Valid: true, Int32: 1},
		StudyProgram:         "Computer Science",
		StudyDegree:          "bachelor",
	}
	createdStudent, err := CreateStudent(suite.ctx, nil, newStudent)
	assert.NoError(suite.T(), err)
	expectedID = createdStudent.ID

	req, _ := http.NewRequest("GET", "/api/students/"+expectedID.String(), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var student studentDTO.Student
	err = json.Unmarshal(w.Body.Bytes(), &student)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedID, student.ID, "Student ID should match the expected ID")
}

func (suite *RouterTestSuite) TestRouterCreateStudent() {
	newStudent := studentDTO.CreateStudent{
		FirstName:            "Bob",
		LastName:             "Smith",
		Email:                "bob.smith@example.com",
		HasUniversityAccount: true,
		MatriculationNumber:  "01234568",
		UniversityLogin:      "bb12xyz",
		Gender:               "male",
		Nationality:          "DE",
		CurrentSemester:      pgtype.Int4{Valid: true, Int32: 1},
		StudyProgram:         "Computer Science",
		StudyDegree:          "bachelor",
	}
	jsonValue, err := json.Marshal(newStudent)
	if err != nil {
		suite.T().Fatalf("Failed to marshal new student: %v", err)
	}

	req, _ := http.NewRequest("POST", "/api/students/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	var createdStudent studentDTO.Student

	err = json.Unmarshal(w.Body.Bytes(), &createdStudent)

	assert.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), uuid.Nil, createdStudent.ID, "Created student should have a valid ID")
	assert.Equal(suite.T(), newStudent.FirstName, createdStudent.FirstName, "First name should match")
	assert.Equal(suite.T(), newStudent.LastName, createdStudent.LastName, "Last name should match")
	assert.Equal(suite.T(), newStudent.Email, createdStudent.Email, "Email should match")
	assert.Equal(suite.T(), newStudent.HasUniversityAccount, createdStudent.HasUniversityAccount, "HasUniversityAccount should match")
	assert.Equal(suite.T(), newStudent.MatriculationNumber, createdStudent.MatriculationNumber, "MatriculationNumber should match")
	assert.Equal(suite.T(), newStudent.UniversityLogin, createdStudent.UniversityLogin, "UniversityLogin should match")
	assert.Equal(suite.T(), newStudent.Gender, createdStudent.Gender, "Gender should match")
	assert.Equal(suite.T(), newStudent.Nationality, createdStudent.Nationality, "Nationality should match")
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}
