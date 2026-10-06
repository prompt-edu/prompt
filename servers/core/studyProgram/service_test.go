package studyProgram

import (
	"context"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/studyProgram/studyProgramDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var (
	computerScienceID           = uuid.MustParse("a1000000-0000-0000-0000-000000000001")
	informationSystemsID        = uuid.MustParse("a1000000-0000-0000-0000-000000000002")
	gamesEngineeringID          = uuid.MustParse("a1000000-0000-0000-0000-000000000003")
	managementAndTechnologyID   = uuid.MustParse("a1000000-0000-0000-0000-000000000004")
	computerScienceStudentID    = uuid.MustParse("b2000000-0000-0000-0000-000000000001")
	paddedComputerScienceID     = uuid.MustParse("b2000000-0000-0000-0000-000000000002")
	informationSystemsStudentID = uuid.MustParse("b2000000-0000-0000-0000-000000000003")
	gamesEngineeringStudentID   = uuid.MustParse("b2000000-0000-0000-0000-000000000004")
	freeTextStudentID           = uuid.MustParse("b2000000-0000-0000-0000-000000000005")
)

type ServiceTestSuite struct {
	suite.Suite
	ctx     context.Context
	cleanup func()
	conn    *pgxpool.Pool
	service *StudyProgramService
}

func (suite *ServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/study_program_test.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.conn = testDB.Conn
	suite.service = NewStudyProgramService(*testDB.Queries, testDB.Conn)
}

func (suite *ServiceTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *ServiceTestSuite) studyProgramOf(studentID uuid.UUID) pgtype.Text {
	var studyProgram pgtype.Text
	err := suite.conn.QueryRow(suite.ctx, "SELECT study_program FROM student WHERE id = $1", studentID).Scan(&studyProgram)
	require.NoError(suite.T(), err)
	return studyProgram
}

func (suite *ServiceTestSuite) countFor(studyProgramID uuid.UUID) int64 {
	counts, err := suite.service.GetStudentCounts(suite.ctx)
	require.NoError(suite.T(), err)
	for _, count := range counts {
		if count.StudyProgramID == studyProgramID {
			return count.StudentCount
		}
	}
	suite.T().Fatalf("no count for study program %s", studyProgramID)
	return 0
}

func (suite *ServiceTestSuite) TestListStudyProgramsIsSortedByName() {
	studyPrograms, err := suite.service.ListStudyPrograms(suite.ctx)
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), studyPrograms)

	for i := 1; i < len(studyPrograms); i++ {
		assert.LessOrEqual(suite.T(), studyPrograms[i-1].Name, studyPrograms[i].Name)
	}
}

func (suite *ServiceTestSuite) TestCreateStudyProgramTrimsInput() {
	created, err := suite.service.CreateStudyProgram(suite.ctx, studyProgramDTO.CreateStudyProgram{
		Name:      "  Robotics, Cognition, Intelligence  ",
		ShortName: " RCI ",
	})
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Robotics, Cognition, Intelligence", created.Name)
	assert.Equal(suite.T(), pgtype.Text{String: "RCI", Valid: true}, created.ShortName)
}

func (suite *ServiceTestSuite) TestCreateStudyProgramStoresBlankShortNameAsNull() {
	created, err := suite.service.CreateStudyProgram(suite.ctx, studyProgramDTO.CreateStudyProgram{
		Name:      "Data Engineering and Analytics",
		ShortName: "   ",
	})
	require.NoError(suite.T(), err)

	assert.False(suite.T(), created.ShortName.Valid)
}

func (suite *ServiceTestSuite) TestCreateStudyProgramRejectsDuplicateIgnoringCaseAndSpaces() {
	_, err := suite.service.CreateStudyProgram(suite.ctx, studyProgramDTO.CreateStudyProgram{
		Name: "  computer SCIENCE ",
	})

	assert.ErrorIs(suite.T(), err, ErrDuplicateStudyProgram)
}

func (suite *ServiceTestSuite) TestCreateStudyProgramRejectsShortNameOfAnotherProgram() {
	_, err := suite.service.CreateStudyProgram(suite.ctx, studyProgramDTO.CreateStudyProgram{
		Name:      "Electrical Engineering",
		ShortName: "EE",
	})
	require.NoError(suite.T(), err)

	_, err = suite.service.CreateStudyProgram(suite.ctx, studyProgramDTO.CreateStudyProgram{
		Name:      "Electronics Engineering",
		ShortName: " ee ",
	})

	assert.ErrorIs(suite.T(), err, ErrDuplicateStudyProgramLabel)
}

func (suite *ServiceTestSuite) TestCreateStudyProgramRejectsShortNameEqualToAnotherProgramsName() {
	_, err := suite.service.CreateStudyProgram(suite.ctx, studyProgramDTO.CreateStudyProgram{
		Name: "Chemistry",
	})
	require.NoError(suite.T(), err)

	_, err = suite.service.CreateStudyProgram(suite.ctx, studyProgramDTO.CreateStudyProgram{
		Name:      "Chemical Engineering",
		ShortName: "chemistry",
	})

	assert.ErrorIs(suite.T(), err, ErrDuplicateStudyProgramLabel)
}

func (suite *ServiceTestSuite) TestCreateStudyProgramRejectsNameEqualToAnotherProgramsShortName() {
	_, err := suite.service.CreateStudyProgram(suite.ctx, studyProgramDTO.CreateStudyProgram{
		Name:      "Aerospace Engineering",
		ShortName: "AE",
	})
	require.NoError(suite.T(), err)

	_, err = suite.service.CreateStudyProgram(suite.ctx, studyProgramDTO.CreateStudyProgram{
		Name: "ae",
	})

	assert.ErrorIs(suite.T(), err, ErrDuplicateStudyProgramLabel)
}

func (suite *ServiceTestSuite) TestUpdateStudyProgramRenamesMatchingStudents() {
	updated, err := suite.service.UpdateStudyProgram(suite.ctx, computerScienceID, studyProgramDTO.UpdateStudyProgram{
		Name:      "Informatics",
		ShortName: "INF",
	})
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Informatics", updated.Name)
	assert.Equal(suite.T(), "Informatics", suite.studyProgramOf(computerScienceStudentID).String)
	assert.Equal(suite.T(), "Informatics", suite.studyProgramOf(paddedComputerScienceID).String)
	assert.Equal(suite.T(), "Information Systems", suite.studyProgramOf(informationSystemsStudentID).String)
	assert.Equal(suite.T(), "Robotics", suite.studyProgramOf(freeTextStudentID).String)
}

func (suite *ServiceTestSuite) TestUpdateStudyProgramCaseOnlyRenameUpdatesStudents() {
	created, err := suite.service.CreateStudyProgram(suite.ctx, studyProgramDTO.CreateStudyProgram{
		Name: "Mechanical Engineering",
	})
	require.NoError(suite.T(), err)
	studentID := uuid.New()
	_, err = suite.conn.Exec(suite.ctx, "INSERT INTO student (id, study_program) VALUES ($1, $2)", studentID, " Mechanical Engineering ")
	require.NoError(suite.T(), err)

	_, err = suite.service.UpdateStudyProgram(suite.ctx, created.ID, studyProgramDTO.UpdateStudyProgram{
		Name: "mechanical engineering",
	})
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "mechanical engineering", suite.studyProgramOf(studentID).String)
}

func (suite *ServiceTestSuite) TestUpdateStudyProgramShortNameOnlyLeavesStudentsUntouched() {
	updated, err := suite.service.UpdateStudyProgram(suite.ctx, informationSystemsID, studyProgramDTO.UpdateStudyProgram{
		Name:      "Information Systems",
		ShortName: "WI",
	})
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), pgtype.Text{String: "WI", Valid: true}, updated.ShortName)
	assert.Equal(suite.T(), "Information Systems", suite.studyProgramOf(informationSystemsStudentID).String)
}

func (suite *ServiceTestSuite) TestUpdateStudyProgramToExistingNameRollsBack() {
	_, err := suite.service.UpdateStudyProgram(suite.ctx, managementAndTechnologyID, studyProgramDTO.UpdateStudyProgram{
		Name: "information systems",
	})
	assert.ErrorIs(suite.T(), err, ErrDuplicateStudyProgram)

	studyPrograms, err := suite.service.ListStudyPrograms(suite.ctx)
	require.NoError(suite.T(), err)
	names := make([]string, 0, len(studyPrograms))
	for _, studyProgram := range studyPrograms {
		names = append(names, studyProgram.Name)
	}
	assert.Contains(suite.T(), names, "Management and Technology")
}

func (suite *ServiceTestSuite) TestUpdateStudyProgramToShortNameOfAnotherProgramRollsBack() {
	_, err := suite.service.CreateStudyProgram(suite.ctx, studyProgramDTO.CreateStudyProgram{
		Name:      "Architecture",
		ShortName: "ARCH",
	})
	require.NoError(suite.T(), err)
	civilEngineering, err := suite.service.CreateStudyProgram(suite.ctx, studyProgramDTO.CreateStudyProgram{
		Name:      "Civil Engineering",
		ShortName: "CE",
	})
	require.NoError(suite.T(), err)

	_, err = suite.service.UpdateStudyProgram(suite.ctx, civilEngineering.ID, studyProgramDTO.UpdateStudyProgram{
		Name:      "Civil Engineering",
		ShortName: "arch",
	})
	assert.ErrorIs(suite.T(), err, ErrDuplicateStudyProgramLabel)

	studyPrograms, err := suite.service.ListStudyPrograms(suite.ctx)
	require.NoError(suite.T(), err)
	for _, studyProgram := range studyPrograms {
		if studyProgram.ID == civilEngineering.ID {
			assert.Equal(suite.T(), "CE", studyProgram.ShortName.String)
		}
	}
}

func (suite *ServiceTestSuite) TestRenameStudyProgramToLabelOfAnotherProgramRollsBack() {
	_, err := suite.service.CreateStudyProgram(suite.ctx, studyProgramDTO.CreateStudyProgram{
		Name:      "Biochemistry",
		ShortName: "BC",
	})
	require.NoError(suite.T(), err)
	bioinformatics, err := suite.service.CreateStudyProgram(suite.ctx, studyProgramDTO.CreateStudyProgram{
		Name: "Bioinformatics",
	})
	require.NoError(suite.T(), err)
	studentID := uuid.New()
	_, err = suite.conn.Exec(suite.ctx, "INSERT INTO student (id, study_program) VALUES ($1, $2)", studentID, "Bioinformatics")
	require.NoError(suite.T(), err)

	_, err = suite.service.UpdateStudyProgram(suite.ctx, bioinformatics.ID, studyProgramDTO.UpdateStudyProgram{
		Name: "bc",
	})
	assert.ErrorIs(suite.T(), err, ErrDuplicateStudyProgramLabel)

	assert.Equal(suite.T(), "Bioinformatics", suite.studyProgramOf(studentID).String)
}

func (suite *ServiceTestSuite) TestUpdateUnknownStudyProgram() {
	_, err := suite.service.UpdateStudyProgram(suite.ctx, uuid.New(), studyProgramDTO.UpdateStudyProgram{
		Name: "Physics",
	})

	assert.ErrorIs(suite.T(), err, ErrStudyProgramNotFound)
}

func (suite *ServiceTestSuite) TestDeleteStudyProgramKeepsStudentValues() {
	err := suite.service.DeleteStudyProgram(suite.ctx, gamesEngineeringID)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Games Engineering", suite.studyProgramOf(gamesEngineeringStudentID).String)
}

func (suite *ServiceTestSuite) TestDeleteUnknownStudyProgram() {
	err := suite.service.DeleteStudyProgram(suite.ctx, uuid.New())

	assert.ErrorIs(suite.T(), err, ErrStudyProgramNotFound)
}

func (suite *ServiceTestSuite) TestGetStudentCountsMatchesTrimmedNames() {
	assert.Equal(suite.T(), int64(2), suite.countFor(informationSystemsID))
	assert.Equal(suite.T(), int64(0), suite.countFor(managementAndTechnologyID))
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
