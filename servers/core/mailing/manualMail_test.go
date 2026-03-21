package mailing

import (
	"context"
	"errors"
	"net/mail"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/mailing/mailingDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	testCoursePhaseID = "33333333-3333-3333-3333-333333333333"
	testRecipientOne  = "44444444-4444-4444-4444-444444444444"
	testRecipientTwo  = "55555555-5555-5555-5555-555555555555"
)

type sentManualMail struct {
	Recipient string
	Subject   string
	Content   string
}

type ManualMailServiceTestSuite struct {
	suite.Suite
	ctx        context.Context
	cleanup    func()
	phaseID    uuid.UUID
	recipient1 uuid.UUID
	recipient2 uuid.UUID

	oldSendMailFn func(
		courseMailingSettings mailingDTO.CourseMailingSettings,
		recipientAddress, subject, htmlBody string,
	) error
	oldNowFn func() time.Time
}

func (suite *ManualMailServiceTestSuite) SetupSuite() {
	if testing.Short() {
		suite.T().Skip("skipping db-backed manual mail tests in short mode")
	}
	defer func() {
		if r := recover(); r != nil {
			suite.T().Skipf("skipping db-backed manual mail tests: %v", r)
		}
	}()

	suite.ctx = context.Background()
	suite.phaseID = uuid.MustParse(testCoursePhaseID)
	suite.recipient1 = uuid.MustParse(testRecipientOne)
	suite.recipient2 = uuid.MustParse(testRecipientTwo)

	testDB, cleanup, err := testutils.SetupTestDB(
		suite.ctx,
		"../database_dumps/manual_mail_test.sql",
		func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) },
	)
	if err != nil {
		suite.T().Skipf("skipping db-backed manual mail tests: %v", err)
	}

	suite.cleanup = cleanup

	MailingServiceSingleton = &MailingService{
		senderEmail: mail.Address{
			Name:    "Reminder Test",
			Address: "noreply@example.com",
		},
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}

	suite.oldSendMailFn = sendMailFn
	suite.oldNowFn = nowFn
}

func (suite *ManualMailServiceTestSuite) TearDownSuite() {
	sendMailFn = suite.oldSendMailFn
	nowFn = suite.oldNowFn
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *ManualMailServiceTestSuite) SetupTest() {
	sendMailFn = suite.oldSendMailFn
	nowFn = suite.oldNowFn
}

func (suite *ManualMailServiceTestSuite) TestSendManualMailHappyPath() {
	fixedNow := time.Date(2026, time.January, 10, 8, 30, 0, 0, time.UTC)
	nowFn = func() time.Time { return fixedNow }

	sentMails := make([]sentManualMail, 0)
	sendMailFn = func(
		courseMailingSettings mailingDTO.CourseMailingSettings,
		recipientAddress, subject, htmlBody string,
	) error {
		sentMails = append(sentMails, sentManualMail{
			Recipient: recipientAddress,
			Subject:   subject,
			Content:   htmlBody,
		})
		return nil
	}

	report, err := SendManualMailToParticipants(
		suite.ctx,
		suite.phaseID,
		mailingDTO.SendManualMailRequest{
			Subject:                         "Reminder {{customType}}",
			Content:                         "Hi {{firstName}}, phase={{phaseName}} deadline={{dueDate}}.",
			RecipientCourseParticipationIDs: []uuid.UUID{suite.recipient1, suite.recipient2},
			AdditionalPlaceholders: map[string]string{
				"customType": "Custom Mail",
				"phaseName":  "Assessment Phase",
				"dueDate":    "09.01.2026 15:00",
			},
		},
	)
	suite.Require().NoError(err)

	assert.Equal(suite.T(), 2, report.RequestedRecipients)
	suite.Require().Len(report.SuccessfulEmails, 2)
	assert.Len(suite.T(), report.FailedEmails, 0)
	assert.Equal(suite.T(), fixedNow, report.SentAt)

	suite.Require().Len(sentMails, 2)
	sentRecipients := []string{sentMails[0].Recipient, sentMails[1].Recipient}
	sort.Strings(sentRecipients)
	assert.Equal(suite.T(), []string{"alice@example.com", "bob@example.com"}, sentRecipients)

	for _, sentMail := range sentMails {
		assert.Contains(suite.T(), sentMail.Subject, "Custom Mail")
		assert.Contains(suite.T(), sentMail.Content, "Assessment Phase")
		assert.Contains(suite.T(), sentMail.Content, "09.01.2026 15:00")
		assert.False(suite.T(), strings.Contains(sentMail.Subject, "{{"))
		assert.False(suite.T(), strings.Contains(sentMail.Content, "{{"))
	}
}

func (suite *ManualMailServiceTestSuite) TestSendManualMailNoRecipients() {
	fixedNow := time.Date(2026, time.January, 14, 8, 0, 0, 0, time.UTC)
	nowFn = func() time.Time { return fixedNow }

	sendCalls := 0
	sendMailFn = func(
		courseMailingSettings mailingDTO.CourseMailingSettings,
		recipientAddress, subject, htmlBody string,
	) error {
		sendCalls++
		return nil
	}

	report, err := SendManualMailToParticipants(
		suite.ctx,
		suite.phaseID,
		mailingDTO.SendManualMailRequest{
			Subject:                         "Reminder",
			Content:                         "Body",
			RecipientCourseParticipationIDs: []uuid.UUID{},
		},
	)
	suite.Require().NoError(err)
	assert.Equal(suite.T(), 0, report.RequestedRecipients)
	assert.Equal(suite.T(), fixedNow, report.SentAt)
	assert.Equal(suite.T(), 0, sendCalls)
}

func (suite *ManualMailServiceTestSuite) TestSendManualMailTemplateMissing() {
	sendWasCalled := false
	sendMailFn = func(
		courseMailingSettings mailingDTO.CourseMailingSettings,
		recipientAddress, subject, htmlBody string,
	) error {
		sendWasCalled = true
		return nil
	}

	_, err := SendManualMailToParticipants(
		suite.ctx,
		suite.phaseID,
		mailingDTO.SendManualMailRequest{
			Subject:                         "",
			Content:                         "",
			RecipientCourseParticipationIDs: []uuid.UUID{suite.recipient1},
		},
	)
	suite.Require().Error(err)
	assert.ErrorIs(suite.T(), err, ErrManualMailValidation)
	assert.Contains(suite.T(), err.Error(), "template incomplete")
	assert.False(suite.T(), sendWasCalled)
}

func (suite *ManualMailServiceTestSuite) TestSendManualMailPartialSendFailure() {
	sendMailFn = func(
		courseMailingSettings mailingDTO.CourseMailingSettings,
		recipientAddress, subject, htmlBody string,
	) error {
		if recipientAddress == "bob@example.com" {
			return errors.New("mail failed")
		}
		return nil
	}

	report, err := SendManualMailToParticipants(
		suite.ctx,
		suite.phaseID,
		mailingDTO.SendManualMailRequest{
			Subject:                         "Reminder",
			Content:                         "Body",
			RecipientCourseParticipationIDs: []uuid.UUID{suite.recipient1, suite.recipient2},
		},
	)
	suite.Require().NoError(err)
	assert.Equal(suite.T(), 2, report.RequestedRecipients)
	assert.Equal(suite.T(), 1, len(report.SuccessfulEmails))
	assert.Equal(suite.T(), 1, len(report.FailedEmails))
	assert.Equal(suite.T(), "bob@example.com", report.FailedEmails[0])
}

func TestManualMailServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ManualMailServiceTestSuite))
}

func TestDeduplicateUUIDList(t *testing.T) {
	idOne := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	idTwo := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	assert.Equal(t, []uuid.UUID{idOne, idTwo}, deduplicateUUIDList([]uuid.UUID{idOne, idTwo, idOne}))
}
