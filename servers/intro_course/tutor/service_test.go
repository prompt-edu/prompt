package tutor

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/intro_course/testutils"
	"github.com/prompt-edu/prompt/servers/intro_course/tutor/tutorDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TutorServiceTestSuite struct {
	suite.Suite
	ctx           context.Context
	cleanup       func()
	coursePhaseID uuid.UUID
	service       TutorService
}

func (suite *TutorServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(suite.ctx, "../database_dumps/intro_course.sql")
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.coursePhaseID = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")

	suite.service = TutorService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	TutorServiceSingleton = &suite.service
}

func (suite *TutorServiceTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func TestTutorServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TutorServiceTestSuite))
}

func (suite *TutorServiceTestSuite) TestGetTutors() {
	tutors, err := GetTutors(suite.ctx, suite.coursePhaseID)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), tutors, 2)
}

func (suite *TutorServiceTestSuite) TestImportTutors() {
	newCoursePhase := uuid.New()
	newTutors := []tutorDTO.Tutor{
		{
			ID:                  uuid.New(),
			FirstName:           "Jane",
			LastName:            "Doe",
			Email:               "jane.doe@example.com",
			MatriculationNumber: "200001",
			UniversityLogin:     "janedoe",
		},
	}

	err := ImportTutors(suite.ctx, newCoursePhase, newTutors)
	assert.NoError(suite.T(), err)

	tutors, err := TutorServiceSingleton.queries.GetAllTutors(suite.ctx, newCoursePhase)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), tutors, 1)
	assert.Equal(suite.T(), "Jane", tutors[0].FirstName)
}

func (suite *TutorServiceTestSuite) TestUpdateGitLabUsername() {
	tutorID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	update := tutorDTO.UpdateTutor{
		GitlabUsername: "updated-gitlab",
	}

	err := UpdateGitLabUsername(suite.ctx, suite.coursePhaseID, tutorID, update)
	assert.NoError(suite.T(), err)

	// Verify the update by getting all tutors and finding the one with the updated ID
	tutors, err := TutorServiceSingleton.queries.GetAllTutors(suite.ctx, suite.coursePhaseID)
	assert.NoError(suite.T(), err)
	found := false
	for _, tutor := range tutors {
		if tutor.ID == tutorID {
			assert.Equal(suite.T(), update.GitlabUsername, tutor.GitlabUsername.String)
			found = true
			break
		}
	}
	assert.True(suite.T(), found, "Updated tutor not found")
}
