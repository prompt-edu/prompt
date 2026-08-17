package student

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	ctx            context.Context
	cleanup        func()
	studentService StudentService
}

func (suite *ServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Set up PostgreSQL container
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
}

func (suite *ServiceTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *ServiceTestSuite) TestGetAllStudents() {
	students, err := GetAllStudents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(students), 0, "Expected at least one student in the initial data")
}

func (suite *ServiceTestSuite) TestGetStudentByID() {
	expectedID, _ := uuid.Parse("3a774200-39a7-4656-bafb-92b7210a93c1")

	student, err := GetStudentByID(suite.ctx, expectedID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedID, student.ID, "Student ID should match the expected ID")
}

func (suite *ServiceTestSuite) TestCreateStudent() {
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
	assert.NotEqual(suite.T(), uuid.Nil, createdStudent.ID, "Created student should have a valid ID")
	assert.Equal(suite.T(), newStudent.FirstName, createdStudent.FirstName, "First name should match")
	assert.Equal(suite.T(), newStudent.LastName, createdStudent.LastName, "Last name should match")
	assert.Equal(suite.T(), newStudent.Email, createdStudent.Email, "Email should match")
	assert.Equal(suite.T(), newStudent.HasUniversityAccount, createdStudent.HasUniversityAccount, "HasUniversityAccount should match")
	assert.Equal(suite.T(), newStudent.MatriculationNumber, createdStudent.MatriculationNumber, "MatriculationNumber should match")
	assert.Equal(suite.T(), newStudent.UniversityLogin, createdStudent.UniversityLogin, "UniversityLogin should match")
	assert.Equal(suite.T(), newStudent.Gender, createdStudent.Gender, "Gender should match")
	assert.Equal(suite.T(), newStudent.Nationality, createdStudent.Nationality, "Nationality should match")
	assert.Equal(suite.T(), newStudent.CurrentSemester.Int32, createdStudent.CurrentSemester.Int32, "CurrentSemester should match")
	assert.Equal(suite.T(), newStudent.StudyProgram, createdStudent.StudyProgram, "StudyProgram should match")
	assert.Equal(suite.T(), newStudent.StudyDegree, createdStudent.StudyDegree, "StudyDegree should match")

	// Verify it exists in the database
	fetchedStudent, err := GetStudentByID(suite.ctx, createdStudent.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), createdStudent.ID, fetchedStudent.ID, "Fetched student ID should match created student ID")
	assert.Equal(suite.T(), createdStudent.FirstName, fetchedStudent.FirstName, "Fetched student First name should match created student First name")
	assert.Equal(suite.T(), createdStudent.LastName, fetchedStudent.LastName, "Fetched student Last name should match created student Last name")
	assert.Equal(suite.T(), createdStudent.Email, fetchedStudent.Email, "Fetched student Email should match created student Email")
	assert.Equal(suite.T(), createdStudent.HasUniversityAccount, fetchedStudent.HasUniversityAccount, "Fetched student HasUniversityAccount should match created student HasUniversityAccount")
	assert.Equal(suite.T(), createdStudent.MatriculationNumber, fetchedStudent.MatriculationNumber, "Fetched student MatriculationNumber should match created student MatriculationNumber")
	assert.Equal(suite.T(), createdStudent.UniversityLogin, fetchedStudent.UniversityLogin, "Fetched student UniversityLogin should match created student UniversityLogin")
	assert.Equal(suite.T(), createdStudent.Gender, fetchedStudent.Gender, "Fetched student Gender should match created student Gender")
	assert.Equal(suite.T(), createdStudent.Nationality, fetchedStudent.Nationality, "Nationality should match")
	assert.Equal(suite.T(), createdStudent.CurrentSemester.Int32, fetchedStudent.CurrentSemester.Int32, "CurrentSemester should match")
	assert.Equal(suite.T(), createdStudent.StudyProgram, fetchedStudent.StudyProgram, "StudyProgram should match")
	assert.Equal(suite.T(), createdStudent.StudyDegree, fetchedStudent.StudyDegree, "StudyDegree should match")
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func TestFilterEnrollmentsToAccessibleCourses(t *testing.T) {
	enrollments := studentDTO.StudentEnrollmentsDTO{
		Courses: []studentDTO.CourseEnrollmentDTO{
			{SemesterTag: "ss25", Name: "ios"},
			{SemesterTag: "ws24", Name: "android"},
		},
	}

	tests := []struct {
		name          string
		userRoles     map[string]bool
		expectedNames []string
	}{
		{
			name:          "prompt admin sees all courses",
			userRoles:     map[string]bool{permissionValidation.PromptAdmin: true},
			expectedNames: []string{"ios", "android"},
		},
		{
			name:          "prompt lecturer sees all courses",
			userRoles:     map[string]bool{permissionValidation.PromptLecturer: true},
			expectedNames: []string{"ios", "android"},
		},
		{
			name:          "course lecturer sees only their course",
			userRoles:     map[string]bool{"ss25-ios-" + permissionValidation.CourseLecturer: true},
			expectedNames: []string{"ios"},
		},
		{
			name:          "course editor sees only their course",
			userRoles:     map[string]bool{"ws24-android-" + permissionValidation.CourseEditor: true},
			expectedNames: []string{"android"},
		},
		{
			name:          "course student role grants no access",
			userRoles:     map[string]bool{"ss25-ios-" + permissionValidation.CourseStudent: true},
			expectedNames: []string{},
		},
		{
			name:          "no relevant roles sees nothing",
			userRoles:     map[string]bool{},
			expectedNames: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filtered := FilterEnrollmentsToAccessibleCourses(enrollments, tc.userRoles)

			actualNames := make([]string, 0, len(filtered.Courses))
			for _, course := range filtered.Courses {
				actualNames = append(actualNames, course.Name)
			}
			assert.ElementsMatch(t, tc.expectedNames, actualNames)
		})
	}
}
