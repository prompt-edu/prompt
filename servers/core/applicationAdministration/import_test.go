package applicationAdministration

import (
	"context"
	"log"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/applicationAdministration/applicationDTO"
	"github.com/prompt-edu/prompt/servers/core/course/courseParticipation"
	"github.com/prompt-edu/prompt/servers/core/coursePhase"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseParticipation"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/student"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// applicationPhaseID is the seeded Application phase from the test dump.
var importApplicationPhaseID = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")

// ApplicationImportTestSuite runs the CSV import tests against a dedicated database container so
// the questions and participations it creates do not leak into the other application suites.
type ApplicationImportTestSuite struct {
	suite.Suite
	router                  *gin.Engine
	ctx                     context.Context
	cleanup                 func()
	applicationAdminService ApplicationService
}

func (suite *ApplicationImportTestSuite) SetupSuite() {
	suite.ctx = context.Background()

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
	student.InitStudentModule(suite.router.Group("/api"), *testDB.Queries, testDB.Conn)
	coursePhase.InitCoursePhaseModule(suite.router.Group("/api"), *testDB.Queries, testDB.Conn)
	courseParticipation.InitCourseParticipationModule(suite.router.Group("/api"), *testDB.Queries, testDB.Conn)
	coursePhaseParticipation.InitCoursePhaseParticipationModule(suite.router.Group("/api"), *testDB.Queries, testDB.Conn)
}

func (suite *ApplicationImportTestSuite) TearDownSuite() {
	suite.cleanup()
}

func TestApplicationImportTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationImportTestSuite))
}

// setApplicationMode sets (or, with an empty mode, clears) restricted_data.applicationMode for a
// phase so a single test can toggle between apply and import mode.
func (suite *ApplicationImportTestSuite) setApplicationMode(phaseID uuid.UUID, mode string) {
	var err error
	if mode == "" {
		_, err = suite.applicationAdminService.conn.Exec(suite.ctx,
			`UPDATE course_phase SET restricted_data = restricted_data - 'applicationMode' WHERE id = $1`, phaseID)
	} else {
		_, err = suite.applicationAdminService.conn.Exec(suite.ctx,
			`UPDATE course_phase SET restricted_data = restricted_data || jsonb_build_object('applicationMode', $2::text) WHERE id = $1`,
			phaseID, mode)
	}
	assert.NoError(suite.T(), err)
}

func (suite *ApplicationImportTestSuite) passStatusForParticipation(courseParticipationID uuid.UUID, phaseID uuid.UUID) string {
	var status string
	err := suite.applicationAdminService.conn.QueryRow(suite.ctx,
		`SELECT pass_status FROM course_phase_participation WHERE course_participation_id = $1 AND course_phase_id = $2`,
		courseParticipationID, phaseID).Scan(&status)
	assert.NoError(suite.T(), err)
	return status
}

func (suite *ApplicationImportTestSuite) TestImportApplications_Success() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	req := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		NewQuestions: []applicationDTO.NewImportQuestion{
			{ColumnKey: "team", Title: "Preferred Team", AllowedLength: 100},
		},
		Rows: []applicationDTO.ImportRow{
			{
				Student: studentDTO.CreateStudent{
					FirstName: "Import", LastName: "One", Email: "import.one@example.com",
					UniversityLogin: "IM01ABC", MatriculationNumber: "01000001",
				},
				Answers: []applicationDTO.ImportAnswer{{ColumnKey: "team", Answer: "Team Rocket"}},
			},
			{
				Student: studentDTO.CreateStudent{
					FirstName: "Import", LastName: "Two", Email: "import.two@example.com",
					UniversityLogin: "im02abc",
				},
				// Empty answer must not create an answer row.
				Answers: []applicationDTO.ImportAnswer{{ColumnKey: "team", Answer: ""}},
			},
		},
	}

	err := validateApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.NoError(suite.T(), err)

	result, err := PostApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, result.Created)
	assert.Equal(suite.T(), 0, result.Updated)
	assert.Len(suite.T(), result.Rows, 2)

	q := suite.applicationAdminService.queries

	// University login is normalized to lowercase and enum fields are defaulted.
	createdStudent, err := q.GetStudentByUniversityLogin(suite.ctx, pgtype.Text{String: "im01abc", Valid: true})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), db.GenderPreferNotToSay, createdStudent.Gender)
	assert.Equal(suite.T(), db.StudyDegreeBachelor, createdStudent.StudyDegree)
	assert.True(suite.T(), createdStudent.HasUniversityAccount.Bool)

	// The imported column became a text question.
	questions, err := q.GetApplicationQuestionsTextForCoursePhase(suite.ctx, importApplicationPhaseID)
	assert.NoError(suite.T(), err)
	teamQuestions := 0
	for _, ques := range questions {
		if ques.Title.String == "Preferred Team" {
			teamQuestions++
		}
	}
	assert.Equal(suite.T(), 1, teamQuestions)

	// Only the non-empty answer was stored.
	var answerCount int
	err = suite.applicationAdminService.conn.QueryRow(suite.ctx,
		`SELECT count(*) FROM application_answer_text aat
		 JOIN application_question_text aqt ON aat.application_question_id = aqt.id
		 WHERE aqt.course_phase_id = $1 AND aqt.title = 'Preferred Team'`, importApplicationPhaseID).Scan(&answerCount)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, answerCount)

	// Chosen pass status is applied.
	assert.Equal(suite.T(), "passed", suite.passStatusForParticipation(*result.Rows[0].CourseParticipationID, importApplicationPhaseID))
}

func (suite *ApplicationImportTestSuite) TestImportApplications_NotAssessedStatus() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	req := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusNotAssessed,
		Rows: []applicationDTO.ImportRow{
			{Student: studentDTO.CreateStudent{
				FirstName: "Not", LastName: "Assessed", Email: "not.assessed@example.com",
				UniversityLogin: "na01abc",
			}},
		},
	}

	result, err := PostApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, result.Created)
	assert.Equal(suite.T(), "not_assessed", suite.passStatusForParticipation(*result.Rows[0].CourseParticipationID, importApplicationPhaseID))
}

func (suite *ApplicationImportTestSuite) TestImportApplications_ReImportIdempotent() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	req := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		NewQuestions: []applicationDTO.NewImportQuestion{
			{ColumnKey: "note", Title: "Reimport Note", AllowedLength: 50},
		},
		Rows: []applicationDTO.ImportRow{
			{
				Student: studentDTO.CreateStudent{
					FirstName: "Re", LastName: "Import", Email: "re.import@example.com",
					UniversityLogin: "ri01abc",
				},
				Answers: []applicationDTO.ImportAnswer{{ColumnKey: "note", Answer: "first"}},
			},
		},
	}

	first, err := PostApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, first.Created)

	// Change the answer to verify overwrite (not duplicate) on re-import.
	req.Rows[0].Answers[0].Answer = "second"
	second, err := PostApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, second.Created)
	assert.Equal(suite.T(), 1, second.Updated)

	var studentCount, questionCount, answerCount int
	err = suite.applicationAdminService.conn.QueryRow(suite.ctx,
		`SELECT count(*) FROM student WHERE university_login = 'ri01abc'`).Scan(&studentCount)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, studentCount)

	err = suite.applicationAdminService.conn.QueryRow(suite.ctx,
		`SELECT count(*) FROM application_question_text WHERE course_phase_id = $1 AND title = 'Reimport Note'`,
		importApplicationPhaseID).Scan(&questionCount)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, questionCount)

	err = suite.applicationAdminService.conn.QueryRow(suite.ctx,
		`SELECT count(*) FROM application_answer_text aat
		 JOIN application_question_text aqt ON aat.application_question_id = aqt.id
		 WHERE aqt.course_phase_id = $1 AND aqt.title = 'Reimport Note'`, importApplicationPhaseID).Scan(&answerCount)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, answerCount)
}

// TestImport_StudentRoleResolvedWithoutMatriculation verifies the relaxed GetStudentRoleStrings
// query resolves the course student role for an imported student that has no matriculation number.
func (suite *ApplicationImportTestSuite) TestImport_StudentRoleResolvedWithoutMatriculation() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	req := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		Rows: []applicationDTO.ImportRow{
			{Student: studentDTO.CreateStudent{
				FirstName: "Role", LastName: "Only", Email: "role.only@example.com",
				UniversityLogin: "ro01abc",
			}},
		},
	}

	_, err := PostApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.NoError(suite.T(), err)

	roles, err := suite.applicationAdminService.queries.GetStudentRoleStrings(suite.ctx, db.GetStudentRoleStringsParams{
		MatriculationNumber: pgtype.Text{String: "", Valid: true},
		UniversityLogin:     pgtype.Text{String: "ro01abc", Valid: true},
	})
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), roles)
}

func (suite *ApplicationImportTestSuite) TestImportApplications_RejectApplyMode() {
	suite.setApplicationMode(importApplicationPhaseID, "")

	req := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		Rows: []applicationDTO.ImportRow{
			{Student: studentDTO.CreateStudent{
				FirstName: "Apply", LastName: "Mode", Email: "apply.mode@example.com",
				UniversityLogin: "am01abc",
			}},
		},
	}

	err := validateApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "import mode")
}

func (suite *ApplicationImportTestSuite) TestImportApplications_DuplicateLoginRejected() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	req := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		Rows: []applicationDTO.ImportRow{
			{Student: studentDTO.CreateStudent{FirstName: "Dup", LastName: "A", Email: "dup.a@example.com", UniversityLogin: "du01abc"}},
			{Student: studentDTO.CreateStudent{FirstName: "Dup", LastName: "B", Email: "dup.b@example.com", UniversityLogin: "DU01ABC"}},
		},
	}

	err := validateApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "duplicate university login")
}

func (suite *ApplicationImportTestSuite) TestImportApplications_InvalidEmailRejected() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	req := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		Rows: []applicationDTO.ImportRow{
			{Student: studentDTO.CreateStudent{FirstName: "Bad", LastName: "Email", Email: "not-an-email", UniversityLogin: "be01abc"}},
		},
	}

	err := validateApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.Error(suite.T(), err)
}

func (suite *ApplicationImportTestSuite) TestImportApplications_UnknownAnswerColumnRejected() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	req := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		Rows: []applicationDTO.ImportRow{
			{
				Student: studentDTO.CreateStudent{FirstName: "Unknown", LastName: "Column", Email: "unknown.column@example.com", UniversityLogin: "uc01abc"},
				Answers: []applicationDTO.ImportAnswer{{ColumnKey: "missing", Answer: "value"}},
			},
		},
	}

	err := validateApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "does not map to an import question")
}
