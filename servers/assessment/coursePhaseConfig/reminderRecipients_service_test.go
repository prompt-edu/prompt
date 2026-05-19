package coursePhaseConfig

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/coursePhaseConfig/coursePhaseConfigDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const reminderTestPhaseID = "88888888-8888-8888-8888-888888888888"

var (
	authorOneID = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1")
	authorTwoID = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2")
	authorThrID = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3")
	teamOneID   = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb1")
	tutorOneID  = uuid.MustParse("cccccccc-cccc-cccc-cccc-ccccccccccc1")
)

type evaluationCompletionSeed struct {
	targetID uuid.UUID
	authorID uuid.UUID
	tpe      assessmentType.AssessmentType
	done     bool
}

type ReminderRecipientsServiceTestSuite struct {
	suite.Suite
	ctx                     context.Context
	cleanup                 func()
	testPhaseID             uuid.UUID
	oldGetParticipationsFn  func(context.Context, string, uuid.UUID) ([]coursePhaseConfigDTO.AssessmentParticipationWithStudent, error)
	oldGetTeamsFn           func(context.Context, string, uuid.UUID) ([]promptTypes.Team, error)
	coursePhaseConfigServer CoursePhaseConfigService
}

func (suite *ReminderRecipientsServiceTestSuite) SetupSuite() {
	if testing.Short() {
		suite.T().Skip("skipping db-backed reminder recipient tests in short mode")
	}
	defer func() {
		if r := recover(); r != nil {
			suite.T().Skipf("skipping db-backed reminder recipient tests: %v", r)
		}
	}()

	suite.ctx = context.Background()
	suite.testPhaseID = uuid.MustParse(reminderTestPhaseID)

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(
		suite.ctx,
		"../database_dumps/reminder_recipients_test.sql",
		func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) },
	)
	if err != nil {
		suite.T().Skipf("skipping db-backed reminder recipient tests: %v", err)
	}
	suite.cleanup = cleanup

	suite.coursePhaseConfigServer = CoursePhaseConfigService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	CoursePhaseConfigSingleton = &suite.coursePhaseConfigServer

	suite.oldGetParticipationsFn = getParticipationsForCoursePhaseFn
	suite.oldGetTeamsFn = getTeamsForCoursePhaseFn
}

func (suite *ReminderRecipientsServiceTestSuite) TearDownSuite() {
	getParticipationsForCoursePhaseFn = suite.oldGetParticipationsFn
	getTeamsForCoursePhaseFn = suite.oldGetTeamsFn
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *ReminderRecipientsServiceTestSuite) SetupTest() {
	getParticipationsForCoursePhaseFn = func(
		ctx context.Context,
		authHeader string,
		coursePhaseID uuid.UUID,
	) ([]coursePhaseConfigDTO.AssessmentParticipationWithStudent, error) {
		return []coursePhaseConfigDTO.AssessmentParticipationWithStudent{
			{
				CoursePhaseParticipationWithStudent: promptTypes.CoursePhaseParticipationWithStudent{
					CourseParticipationID: authorOneID,
				},
				TeamID: &teamOneID,
			},
			{
				CoursePhaseParticipationWithStudent: promptTypes.CoursePhaseParticipationWithStudent{
					CourseParticipationID: authorTwoID,
				},
				TeamID: &teamOneID,
			},
			{
				CoursePhaseParticipationWithStudent: promptTypes.CoursePhaseParticipationWithStudent{
					CourseParticipationID: authorThrID,
				},
				TeamID: nil,
			},
		}, nil
	}

	getTeamsForCoursePhaseFn = func(
		ctx context.Context,
		authHeader string,
		coursePhaseID uuid.UUID,
	) ([]promptTypes.Team, error) {
		return []promptTypes.Team{
			{
				ID:   teamOneID,
				Name: "Team One",
				Members: []promptTypes.Person{
					{ID: authorOneID, FirstName: "Alice", LastName: "One"},
					{ID: authorTwoID, FirstName: "Bob", LastName: "Two"},
				},
				Tutors: []promptTypes.Person{
					{ID: tutorOneID, FirstName: "Tina", LastName: "Tutor"},
				},
			},
		}, nil
	}

	suite.resetConfig()
	suite.setEvaluationCompletions(nil)
}

func (suite *ReminderRecipientsServiceTestSuite) resetConfig() {
	_, err := suite.coursePhaseConfigServer.conn.Exec(
		suite.ctx,
		`UPDATE course_phase_config
		 SET self_evaluation_enabled = true,
		     peer_evaluation_enabled = true,
		     tutor_evaluation_enabled = true,
		     self_evaluation_deadline = $1,
		     peer_evaluation_deadline = $1,
		     tutor_evaluation_deadline = $1
		 WHERE course_phase_id = $2`,
		time.Now().Add(-4*time.Hour),
		suite.testPhaseID,
	)
	suite.Require().NoError(err)
}

func (suite *ReminderRecipientsServiceTestSuite) setEvaluationCompletions(completions []evaluationCompletionSeed) {
	_, err := suite.coursePhaseConfigServer.conn.Exec(
		suite.ctx,
		`DELETE FROM evaluation_completion WHERE course_phase_id = $1`,
		suite.testPhaseID,
	)
	suite.Require().NoError(err)

	for _, completion := range completions {
		_, err := suite.coursePhaseConfigServer.conn.Exec(
			suite.ctx,
			`INSERT INTO evaluation_completion (
				id,
				course_participation_id,
				course_phase_id,
				author_course_participation_id,
				completed_at,
				completed,
				type
			) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New(),
			completion.targetID,
			suite.testPhaseID,
			completion.authorID,
			time.Now(),
			completion.done,
			string(completion.tpe),
		)
		suite.Require().NoError(err)
	}
}

func (suite *ReminderRecipientsServiceTestSuite) TestGetEvaluationReminderRecipientsSelf() {
	suite.setEvaluationCompletions([]evaluationCompletionSeed{
		{targetID: authorOneID, authorID: authorOneID, tpe: assessmentType.Self, done: true},
		{targetID: authorThrID, authorID: authorThrID, tpe: assessmentType.Self, done: true},
	})

	response, err := GetEvaluationReminderRecipients(
		suite.ctx,
		"Bearer test",
		suite.testPhaseID,
		assessmentType.Self,
	)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), assessmentType.Self, response.EvaluationType)
	assert.Equal(suite.T(), 3, response.TotalAuthors)
	assert.Equal(suite.T(), 2, response.CompletedAuthors)
	assert.Equal(suite.T(), []uuid.UUID{authorTwoID}, response.IncompleteAuthorCourseParticipationIDs)
}

func (suite *ReminderRecipientsServiceTestSuite) TestGetEvaluationReminderRecipientsPeer() {
	suite.setEvaluationCompletions([]evaluationCompletionSeed{
		{targetID: authorTwoID, authorID: authorOneID, tpe: assessmentType.Peer, done: true},
	})

	response, err := GetEvaluationReminderRecipients(
		suite.ctx,
		"Bearer test",
		suite.testPhaseID,
		assessmentType.Peer,
	)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), assessmentType.Peer, response.EvaluationType)
	assert.Equal(suite.T(), 3, response.TotalAuthors)
	assert.Equal(suite.T(), 1, response.CompletedAuthors)
	assert.Equal(suite.T(), []uuid.UUID{authorTwoID, authorThrID}, response.IncompleteAuthorCourseParticipationIDs)
}

func (suite *ReminderRecipientsServiceTestSuite) TestGetEvaluationReminderRecipientsTutor() {
	suite.setEvaluationCompletions([]evaluationCompletionSeed{
		{targetID: tutorOneID, authorID: authorOneID, tpe: assessmentType.Tutor, done: true},
	})

	response, err := GetEvaluationReminderRecipients(
		suite.ctx,
		"Bearer test",
		suite.testPhaseID,
		assessmentType.Tutor,
	)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), assessmentType.Tutor, response.EvaluationType)
	assert.Equal(suite.T(), 3, response.TotalAuthors)
	assert.Equal(suite.T(), 1, response.CompletedAuthors)
	assert.Equal(suite.T(), []uuid.UUID{authorTwoID, authorThrID}, response.IncompleteAuthorCourseParticipationIDs)
}

func (suite *ReminderRecipientsServiceTestSuite) TestGetEvaluationReminderRecipientsDisabledType() {
	_, err := suite.coursePhaseConfigServer.conn.Exec(
		suite.ctx,
		`UPDATE course_phase_config
		 SET peer_evaluation_enabled = false
		 WHERE course_phase_id = $1`,
		suite.testPhaseID,
	)
	suite.Require().NoError(err)

	response, err := GetEvaluationReminderRecipients(
		suite.ctx,
		"Bearer test",
		suite.testPhaseID,
		assessmentType.Peer,
	)
	suite.Require().NoError(err)

	assert.False(suite.T(), response.EvaluationEnabled)
	assert.Equal(suite.T(), 3, response.CompletedAuthors)
	assert.Empty(suite.T(), response.IncompleteAuthorCourseParticipationIDs)
}

func (suite *ReminderRecipientsServiceTestSuite) TestGetEvaluationReminderRecipientsDeadlineFlag() {
	pastDeadline := time.Now().Add(-2 * time.Hour)
	futureDeadline := time.Now().Add(24 * time.Hour)

	_, err := suite.coursePhaseConfigServer.conn.Exec(
		suite.ctx,
		`UPDATE course_phase_config
		 SET self_evaluation_deadline = $1
		 WHERE course_phase_id = $2`,
		pastDeadline,
		suite.testPhaseID,
	)
	suite.Require().NoError(err)

	pastResponse, err := GetEvaluationReminderRecipients(
		suite.ctx,
		"Bearer test",
		suite.testPhaseID,
		assessmentType.Self,
	)
	suite.Require().NoError(err)
	assert.True(suite.T(), pastResponse.DeadlinePassed)
	suite.Require().NotNil(pastResponse.Deadline)

	_, err = suite.coursePhaseConfigServer.conn.Exec(
		suite.ctx,
		`UPDATE course_phase_config
		 SET self_evaluation_deadline = $1
		 WHERE course_phase_id = $2`,
		futureDeadline,
		suite.testPhaseID,
	)
	suite.Require().NoError(err)

	futureResponse, err := GetEvaluationReminderRecipients(
		suite.ctx,
		"Bearer test",
		suite.testPhaseID,
		assessmentType.Self,
	)
	suite.Require().NoError(err)
	assert.False(suite.T(), futureResponse.DeadlinePassed)
	suite.Require().NotNil(futureResponse.Deadline)
}

func (suite *ReminderRecipientsServiceTestSuite) TestGetEvaluationReminderRecipientsDeadlineBoundaryIsPassed() {
	boundaryDeadline := time.Now()

	_, err := suite.coursePhaseConfigServer.conn.Exec(
		suite.ctx,
		`UPDATE course_phase_config
		 SET self_evaluation_deadline = $1
		 WHERE course_phase_id = $2`,
		boundaryDeadline,
		suite.testPhaseID,
	)
	suite.Require().NoError(err)

	response, err := GetEvaluationReminderRecipients(
		suite.ctx,
		"Bearer test",
		suite.testPhaseID,
		assessmentType.Self,
	)
	suite.Require().NoError(err)
	assert.True(suite.T(), response.DeadlinePassed)
	suite.Require().NotNil(response.Deadline)
}

func TestReminderRecipientsServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ReminderRecipientsServiceTestSuite))
}
