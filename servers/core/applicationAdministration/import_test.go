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
	applicationAdminService *ApplicationService
}

func (suite *ApplicationImportTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/application_administration.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}

	suite.cleanup = cleanup
	suite.applicationAdminService = NewApplicationService(*testDB.Queries, testDB.Conn)
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

	err := suite.applicationAdminService.validateApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.NoError(suite.T(), err)

	result, err := suite.applicationAdminService.PostApplicationImport(suite.ctx, importApplicationPhaseID, req)
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

	result, err := suite.applicationAdminService.PostApplicationImport(suite.ctx, importApplicationPhaseID, req)
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

	first, err := suite.applicationAdminService.PostApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, first.Created)

	// Change the answer to verify overwrite (not duplicate) on re-import.
	req.Rows[0].Answers[0].Answer = "second"
	second, err := suite.applicationAdminService.PostApplicationImport(suite.ctx, importApplicationPhaseID, req)
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

// TestImportApplications_ReImportPreservesOmittedAttributes verifies that re-importing a student
// with a CSV that omits optional attributes keeps the previously stored values instead of resetting
// them to defaults.
func (suite *ApplicationImportTestSuite) TestImportApplications_ReImportPreservesOmittedAttributes() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	full := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		Rows: []applicationDTO.ImportRow{
			{Student: studentDTO.CreateStudent{
				FirstName: "Keep", LastName: "Attrs", Email: "keep.attrs@example.com",
				UniversityLogin: "ka01abc",
				Gender:          db.GenderMale,
				Nationality:     "DE",
				StudyDegree:     db.StudyDegreeMaster,
				StudyProgram:    "Informatics",
				CurrentSemester: pgtype.Int4{Int32: 4, Valid: true},
			}},
		},
	}
	_, err := suite.applicationAdminService.PostApplicationImport(suite.ctx, importApplicationPhaseID, full)
	assert.NoError(suite.T(), err)

	// Re-import the same student with only the required fields; optional attributes are omitted.
	partial := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		Rows: []applicationDTO.ImportRow{
			{Student: studentDTO.CreateStudent{
				FirstName: "Keep", LastName: "Attrs", Email: "keep.attrs@example.com",
				UniversityLogin: "ka01abc",
			}},
		},
	}
	_, err = suite.applicationAdminService.PostApplicationImport(suite.ctx, importApplicationPhaseID, partial)
	assert.NoError(suite.T(), err)

	stored, err := suite.applicationAdminService.queries.GetStudentByUniversityLogin(suite.ctx, pgtype.Text{String: "ka01abc", Valid: true})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), db.GenderMale, stored.Gender)
	assert.Equal(suite.T(), "DE", stored.Nationality.String)
	assert.Equal(suite.T(), db.StudyDegreeMaster, stored.StudyDegree)
	assert.Equal(suite.T(), "Informatics", stored.StudyProgram.String)
	assert.Equal(suite.T(), int32(4), stored.CurrentSemester.Int32)
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

	_, err := suite.applicationAdminService.PostApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.NoError(suite.T(), err)

	// The token at login carries no matriculation number.
	roles, err := suite.applicationAdminService.queries.GetStudentRoleStrings(suite.ctx, db.GetStudentRoleStringsParams{
		MatriculationNumber: pgtype.Text{String: "", Valid: true},
		UniversityLogin:     pgtype.Text{String: "ro01abc", Valid: true},
	})
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), roles)

	// The token later carries a real matriculation number, while the imported student row still has
	// an empty one. This is the production login path and must still resolve the course role.
	rolesWithMatriculation, err := suite.applicationAdminService.queries.GetStudentRoleStrings(suite.ctx, db.GetStudentRoleStringsParams{
		MatriculationNumber: pgtype.Text{String: "01900001", Valid: true},
		UniversityLogin:     pgtype.Text{String: "ro01abc", Valid: true},
	})
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), rolesWithMatriculation)
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

	err := suite.applicationAdminService.validateApplicationImport(suite.ctx, importApplicationPhaseID, req)
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

	err := suite.applicationAdminService.validateApplicationImport(suite.ctx, importApplicationPhaseID, req)
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

	err := suite.applicationAdminService.validateApplicationImport(suite.ctx, importApplicationPhaseID, req)
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

	err := suite.applicationAdminService.validateApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "does not map to an import question")
}

func (suite *ApplicationImportTestSuite) TestImportApplications_DuplicateQuestionTitleRejected() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	// Two columns with distinct keys but the same title would collapse onto one question.
	req := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		NewQuestions: []applicationDTO.NewImportQuestion{
			{ColumnKey: "team1", Title: "Team", AllowedLength: 50},
			{ColumnKey: "team2", Title: "Team", AllowedLength: 50},
		},
		Rows: []applicationDTO.ImportRow{
			{Student: studentDTO.CreateStudent{FirstName: "Dup", LastName: "Title", Email: "dup.title@example.com", UniversityLogin: "dt01abc"}},
		},
	}

	err := suite.applicationAdminService.validateApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "duplicate question title")
}

func (suite *ApplicationImportTestSuite) TestImportApplications_InvalidAllowedLengthRejected() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	req := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		NewQuestions: []applicationDTO.NewImportQuestion{
			{ColumnKey: "note", Title: "Note", AllowedLength: 0},
		},
		Rows: []applicationDTO.ImportRow{
			{Student: studentDTO.CreateStudent{FirstName: "Bad", LastName: "Length", Email: "bad.length@example.com", UniversityLogin: "bl01abc"}},
		},
	}

	err := suite.applicationAdminService.validateApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "allowed length")
}

func (suite *ApplicationImportTestSuite) TestImportApplications_AnswerExceedingAllowedLengthRejected() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	req := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		NewQuestions: []applicationDTO.NewImportQuestion{
			{ColumnKey: "short", Title: "Short Answer", AllowedLength: 5},
		},
		Rows: []applicationDTO.ImportRow{
			{
				Student: studentDTO.CreateStudent{FirstName: "Long", LastName: "Answer", Email: "long.answer@example.com", UniversityLogin: "la01abc"},
				Answers: []applicationDTO.ImportAnswer{{ColumnKey: "short", Answer: "way too long for five"}},
			},
		},
	}

	// validateApplicationImport rejects it before the transaction, so the handler maps it to a 400
	// (client error) instead of a 500.
	err := suite.applicationAdminService.validateApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "exceeds the allowed length")

	_, err = suite.applicationAdminService.PostApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "exceeds the allowed length")
	assert.ErrorIs(suite.T(), err, ErrImportAnswerTooLong)
}

// TestImportApplications_ReuseUsesPersistedLength verifies that when a re-import reuses a question by
// title, validation enforces the persisted allowed length rather than the (possibly larger) length
// sent in the request.
func (suite *ApplicationImportTestSuite) TestImportApplications_ReuseUsesPersistedLength() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	// Create the question with a small persisted limit.
	create := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		NewQuestions: []applicationDTO.NewImportQuestion{
			{ColumnKey: "note", Title: "Persisted Length Note", AllowedLength: 10},
		},
		Rows: []applicationDTO.ImportRow{
			{
				Student: studentDTO.CreateStudent{FirstName: "Persist", LastName: "Len", Email: "persist.len@example.com", UniversityLogin: "pl01abc"},
				Answers: []applicationDTO.ImportAnswer{{ColumnKey: "note", Answer: "tiny"}},
			},
		},
	}
	assert.NoError(suite.T(), suite.applicationAdminService.validateApplicationImport(suite.ctx, importApplicationPhaseID, create))
	_, err := suite.applicationAdminService.PostApplicationImport(suite.ctx, importApplicationPhaseID, create)
	assert.NoError(suite.T(), err)

	// Re-import the same title with a larger requested limit and an answer that fits the request but
	// exceeds the persisted limit. Validation must follow the persisted limit and reject it.
	reuse := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		NewQuestions: []applicationDTO.NewImportQuestion{
			{ColumnKey: "note", Title: "Persisted Length Note", AllowedLength: 100},
		},
		Rows: []applicationDTO.ImportRow{
			{
				Student: studentDTO.CreateStudent{FirstName: "Persist", LastName: "Len", Email: "persist.len@example.com", UniversityLogin: "pl01abc"},
				Answers: []applicationDTO.ImportAnswer{{ColumnKey: "note", Answer: "this answer is definitely longer than ten"}},
			},
		},
	}
	err = suite.applicationAdminService.validateApplicationImport(suite.ctx, importApplicationPhaseID, reuse)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "exceeds the allowed length of 10")
}

// TestImportApplications_ReImportKeepsManualPassStatus verifies a re-import does not overwrite a
// pass status a lecturer set by hand on a participation the import did not just create.
func (suite *ApplicationImportTestSuite) TestImportApplications_ReImportKeepsManualPassStatus() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	req := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusNotAssessed,
		Rows: []applicationDTO.ImportRow{
			{Student: studentDTO.CreateStudent{
				FirstName: "Manual", LastName: "Status", Email: "manual.status@example.com",
				UniversityLogin: "ms01abc",
			}},
		},
	}

	first, err := suite.applicationAdminService.PostApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, first.Created)
	participationID := *first.Rows[0].CourseParticipationID

	// The lecturer rejects the student by hand.
	_, err = suite.applicationAdminService.conn.Exec(suite.ctx,
		`UPDATE course_phase_participation SET pass_status = 'failed' WHERE course_participation_id = $1 AND course_phase_id = $2`,
		participationID, importApplicationPhaseID)
	assert.NoError(suite.T(), err)

	// Re-importing the same roster with the default status must not flip the manual decision.
	req.PassStatus = db.PassStatusPassed
	second, err := suite.applicationAdminService.PostApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, second.Created)
	assert.Equal(suite.T(), 1, second.Updated)
	assert.Equal(suite.T(), "failed", suite.passStatusForParticipation(participationID, importApplicationPhaseID))
}

// TestImportApplications_QuestionDefaultsPersisted verifies imported questions store the column
// defaults for accessible_for_other_phases and access_key instead of NULL, so the Questions editor
// (which validates a boolean) is not broken by an import.
func (suite *ApplicationImportTestSuite) TestImportApplications_QuestionDefaultsPersisted() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	req := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		NewQuestions: []applicationDTO.NewImportQuestion{
			{ColumnKey: "defaults", Title: "Defaults Question", AllowedLength: 50},
		},
		Rows: []applicationDTO.ImportRow{
			{Student: studentDTO.CreateStudent{FirstName: "Def", LastName: "Aults", Email: "def.aults@example.com", UniversityLogin: "da01abc"}},
		},
	}
	_, err := suite.applicationAdminService.PostApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.NoError(suite.T(), err)

	var accessible pgtype.Bool
	var accessKey pgtype.Text
	err = suite.applicationAdminService.conn.QueryRow(suite.ctx,
		`SELECT accessible_for_other_phases, access_key FROM application_question_text
		 WHERE course_phase_id = $1 AND title = 'Defaults Question'`, importApplicationPhaseID).Scan(&accessible, &accessKey)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), accessible.Valid)
	assert.False(suite.T(), accessible.Bool)
	assert.True(suite.T(), accessKey.Valid)
	assert.Equal(suite.T(), "", accessKey.String)
}

// TestImport_RoleNotResolvedForStoredMatriculation verifies the tightened GetStudentRoleStrings
// query: a token that carries no matriculation claim resolves the role only for a student whose
// stored matriculation is also empty, never one that carries a number.
func (suite *ApplicationImportTestSuite) TestImport_RoleNotResolvedForStoredMatriculation() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	req := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		Rows: []applicationDTO.ImportRow{
			{Student: studentDTO.CreateStudent{
				FirstName: "Has", LastName: "Matriculation", Email: "has.matriculation@example.com",
				UniversityLogin: "hm01abc", MatriculationNumber: "01000009",
			}},
		},
	}
	_, err := suite.applicationAdminService.PostApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.NoError(suite.T(), err)

	// A token missing the matriculation claim must not match this student, who carries one.
	roles, err := suite.applicationAdminService.queries.GetStudentRoleStrings(suite.ctx, db.GetStudentRoleStringsParams{
		MatriculationNumber: pgtype.Text{String: "", Valid: true},
		UniversityLogin:     pgtype.Text{String: "hm01abc", Valid: true},
	})
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), roles)

	// The real matriculation number still resolves the role.
	matched, err := suite.applicationAdminService.queries.GetStudentRoleStrings(suite.ctx, db.GetStudentRoleStringsParams{
		MatriculationNumber: pgtype.Text{String: "01000009", Valid: true},
		UniversityLogin:     pgtype.Text{String: "hm01abc", Valid: true},
	})
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), matched)
}

func (suite *ApplicationImportTestSuite) TestImportApplications_DuplicateEmailRejected() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	req := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		Rows: []applicationDTO.ImportRow{
			{Student: studentDTO.CreateStudent{FirstName: "Dup", LastName: "MailA", Email: "shared@example.com", UniversityLogin: "de01abc"}},
			{Student: studentDTO.CreateStudent{FirstName: "Dup", LastName: "MailB", Email: "Shared@example.com", UniversityLogin: "de02abc"}},
		},
	}

	err := suite.applicationAdminService.validateApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "duplicate email")
}

func (suite *ApplicationImportTestSuite) TestImportApplications_EmptyRowsRejected() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	req := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		NewQuestions: []applicationDTO.NewImportQuestion{
			{ColumnKey: "note", Title: "Empty Rows Note", AllowedLength: 50},
		},
		Rows: []applicationDTO.ImportRow{},
	}

	err := suite.applicationAdminService.validateApplicationImport(suite.ctx, importApplicationPhaseID, req)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "no rows")
}

func (suite *ApplicationImportTestSuite) TestImportApplications_InvalidOptionalFieldsRejected() {
	suite.setApplicationMode(importApplicationPhaseID, "import")
	defer suite.setApplicationMode(importApplicationPhaseID, "")

	// A semester value the CSV does provide must be valid, since it is written to the shared student row.
	semesterReq := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		Rows: []applicationDTO.ImportRow{
			{Student: studentDTO.CreateStudent{
				FirstName: "Zero", LastName: "Semester", Email: "zero.semester@example.com",
				UniversityLogin: "zs01abc", CurrentSemester: pgtype.Int4{Int32: 0, Valid: true},
			}},
		},
	}
	err := suite.applicationAdminService.validateApplicationImport(suite.ctx, importApplicationPhaseID, semesterReq)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "semester is invalid")

	// A study degree outside the enum must be rejected rather than silently persisted.
	degreeReq := applicationDTO.ImportApplicationRequest{
		PassStatus: db.PassStatusPassed,
		Rows: []applicationDTO.ImportRow{
			{Student: studentDTO.CreateStudent{
				FirstName: "Bad", LastName: "Degree", Email: "bad.degree@example.com",
				UniversityLogin: "bd01abc", StudyDegree: db.StudyDegree("phd"),
			}},
		},
	}
	err = suite.applicationAdminService.validateApplicationImport(suite.ctx, importApplicationPhaseID, degreeReq)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "study degree is invalid")
}
