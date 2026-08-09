package mailing

import (
	"context"
	"net/mail"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/mailing/mailingDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	passedParticipation      = "44444444-4444-4444-4444-444444444444"
	failedParticipation      = "55555555-5555-5555-5555-555555555555"
	notAssessedParticipation = "99999999-9999-9999-9999-999999999999"
)

type StatusMailServiceTestSuite struct {
	suite.Suite
	ctx         context.Context
	cleanup     func()
	phaseID     uuid.UUID
	passed      uuid.UUID
	failed      uuid.UUID
	notAssessed uuid.UUID

	oldSendMailFn func(
		courseMailingSettings mailingDTO.CourseMailingSettings,
		recipientAddress, subject, htmlBody string,
	) error
}

func (suite *StatusMailServiceTestSuite) SetupSuite() {
	if testing.Short() {
		suite.T().Skip("skipping db-backed status mail tests in short mode")
	}
	defer func() {
		if r := recover(); r != nil {
			suite.T().Skipf("skipping db-backed status mail tests: %v", r)
		}
	}()

	suite.ctx = context.Background()
	suite.phaseID = uuid.MustParse(testCoursePhaseID)
	suite.passed = uuid.MustParse(passedParticipation)
	suite.failed = uuid.MustParse(failedParticipation)
	suite.notAssessed = uuid.MustParse(notAssessedParticipation)

	testDB, cleanup, err := testutils.SetupTestDB(
		suite.ctx,
		"../database_dumps/mailing_test.sql",
		func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) },
	)
	if err != nil {
		suite.T().Skipf("skipping db-backed status mail tests: %v", err)
	}

	suite.cleanup = cleanup

	MailingServiceSingleton = &MailingService{
		senderEmail: mail.Address{
			Name:    "Status Mail Test",
			Address: "noreply@example.com",
		},
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}

	suite.oldSendMailFn = sendMailFn
}

func (suite *StatusMailServiceTestSuite) TearDownSuite() {
	sendMailFn = suite.oldSendMailFn
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *StatusMailServiceTestSuite) SetupTest() {
	sendMailFn = suite.oldSendMailFn
}

func (suite *StatusMailServiceTestSuite) recordSentRecipients() *[]string {
	sentRecipients := make([]string, 0)
	sendMailFn = func(
		courseMailingSettings mailingDTO.CourseMailingSettings,
		recipientAddress, subject, htmlBody string,
	) error {
		sentRecipients = append(sentRecipients, recipientAddress)
		return nil
	}
	return &sentRecipients
}

func (suite *StatusMailServiceTestSuite) TestSendStatusMailToAllParticipantsWithStatus() {
	sentRecipients := suite.recordSentRecipients()

	report, err := SendStatusMailManualTrigger(suite.ctx, suite.phaseID, db.PassStatusPassed, nil)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), []string{"alice@example.com"}, report.SuccessfulEmails)
	assert.Empty(suite.T(), report.FailedEmails)
	assert.Equal(suite.T(), []string{"alice@example.com"}, *sentRecipients)
}

func (suite *StatusMailServiceTestSuite) TestSendStatusMailToAllParticipantsUsesStatusTemplate() {
	sentRecipients := suite.recordSentRecipients()

	report, err := SendStatusMailManualTrigger(suite.ctx, suite.phaseID, db.PassStatusFailed, nil)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), []string{"bob@example.com"}, report.SuccessfulEmails)
	assert.Equal(suite.T(), []string{"bob@example.com"}, *sentRecipients)
}

func (suite *StatusMailServiceTestSuite) TestSendStatusMailToSelectedRecipients() {
	sentRecipients := suite.recordSentRecipients()

	report, err := SendStatusMailManualTrigger(
		suite.ctx,
		suite.phaseID,
		db.PassStatusPassed,
		[]uuid.UUID{suite.passed},
	)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), []string{"alice@example.com"}, report.SuccessfulEmails)
	assert.Empty(suite.T(), report.FailedEmails)
	assert.Equal(suite.T(), []string{"alice@example.com"}, *sentRecipients)
}

func (suite *StatusMailServiceTestSuite) TestSendStatusMailSkipsSelectedRecipientsWithOtherStatus() {
	sentRecipients := suite.recordSentRecipients()

	report, err := SendStatusMailManualTrigger(
		suite.ctx,
		suite.phaseID,
		db.PassStatusPassed,
		[]uuid.UUID{suite.passed, suite.failed, suite.notAssessed},
	)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), []string{"alice@example.com"}, report.SuccessfulEmails)
	assert.Equal(suite.T(), []string{"alice@example.com"}, *sentRecipients)
}

func (suite *StatusMailServiceTestSuite) TestSendStatusMailToEmptyRecipientSelection() {
	sentRecipients := suite.recordSentRecipients()

	report, err := SendStatusMailManualTrigger(
		suite.ctx,
		suite.phaseID,
		db.PassStatusPassed,
		[]uuid.UUID{},
	)
	suite.Require().NoError(err)

	assert.NotNil(suite.T(), report.SuccessfulEmails)
	assert.NotNil(suite.T(), report.FailedEmails)
	assert.Empty(suite.T(), report.SuccessfulEmails)
	assert.Empty(suite.T(), report.FailedEmails)
	assert.Empty(suite.T(), *sentRecipients)
}

func TestStatusMailServiceTestSuite(t *testing.T) {
	suite.Run(t, new(StatusMailServiceTestSuite))
}
