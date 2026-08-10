package mailing

import (
	"context"
	"errors"
	"net/mail"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/mailing/mailingDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type StatusMailDedupTestSuite struct {
	suite.Suite
	ctx     context.Context
	cleanup func()
	queries *db.Queries
	conn    *pgxpool.Pool

	oldSendMailFn func(
		courseMailingSettings mailingDTO.CourseMailingSettings,
		recipientAddress, subject, htmlBody string,
	) error
	oldNowFn func() time.Time
}

// Known IDs from database_dumps/application_administration.sql
var (
	dedupCourseID      = uuid.MustParse("be780b32-a678-4b79-ae1c-80071771d254")
	dedupCoursePhaseID = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	dedupCP1           = uuid.MustParse("82d7efae-d545-4cc5-9b94-5d0ee1e50d25")
	dedupCP2           = uuid.MustParse("32aa070e-67c3-4a69-852a-ba3b5e849a4d")
)

const (
	dedupSentAt = "2026-01-01T00:00:00Z"
	dedupEmail1 = "alice@example.com"
	dedupEmail2 = "bob@example.com"
)

func (suite *StatusMailDedupTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/application_administration.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	suite.Require().NoError(err, "failed to set up test database")
	suite.cleanup = cleanup
	suite.queries = testDB.Queries
	suite.conn = testDB.Conn

	MailingServiceSingleton = &MailingService{
		senderEmail: mail.Address{Name: "Status Mail Test", Address: "noreply@example.com"},
		queries:     *testDB.Queries,
		conn:        testDB.Conn,
	}

	suite.oldSendMailFn = sendMailFn
	suite.oldNowFn = nowFn

	suite.configureTestData()
}

func (suite *StatusMailDedupTestSuite) TearDownSuite() {
	sendMailFn = suite.oldSendMailFn
	nowFn = suite.oldNowFn
	suite.cleanup()
}

func (suite *StatusMailDedupTestSuite) SetupTest() {
	sendMailFn = suite.oldSendMailFn
	nowFn = suite.oldNowFn
}

// configureTestData adds the reply-to and passed-status templates the trigger requires and gives the
// two seeded students distinct addresses, since the dump shares one email between them.
func (suite *StatusMailDedupTestSuite) configureTestData() {
	_, err := suite.conn.Exec(suite.ctx,
		`UPDATE course SET restricted_data = COALESCE(restricted_data, '{}'::jsonb) || $2::jsonb WHERE id = $1`,
		dedupCourseID, `{"mailingSettings": {"replyToEmail": "replyto@example.com", "replyToName": "Course Team"}}`)
	suite.Require().NoError(err)

	_, err = suite.conn.Exec(suite.ctx,
		`UPDATE course_phase SET restricted_data = COALESCE(restricted_data, '{}'::jsonb) || $2::jsonb WHERE id = $1`,
		dedupCoursePhaseID, `{"mailingSettings": {"passedMailSubject": "Welcome {{firstName}}", "passedMailContent": "Hi {{firstName}}, you passed."}}`)
	suite.Require().NoError(err)

	for cpID, email := range map[uuid.UUID]string{dedupCP1: dedupEmail1, dedupCP2: dedupEmail2} {
		_, err = suite.conn.Exec(suite.ctx,
			`UPDATE student SET email = $2 WHERE id = (SELECT student_id FROM course_participation WHERE id = $1)`,
			cpID, email)
		suite.Require().NoError(err)
	}
}

func (suite *StatusMailDedupTestSuite) setParticipation(cpID uuid.UUID, passStatus string, restrictedData string) {
	_, err := suite.conn.Exec(suite.ctx,
		`UPDATE course_phase_participation SET pass_status = $3::pass_status, restricted_data = $4::jsonb
		 WHERE course_phase_id = $1 AND course_participation_id = $2`,
		dedupCoursePhaseID, cpID, passStatus, restrictedData)
	assert.NoError(suite.T(), err)
}

func (suite *StatusMailDedupTestSuite) setParticipationNull(cpID uuid.UUID, passStatus string) {
	_, err := suite.conn.Exec(suite.ctx,
		`UPDATE course_phase_participation SET pass_status = $3::pass_status, restricted_data = NULL
		 WHERE course_phase_id = $1 AND course_participation_id = $2`,
		dedupCoursePhaseID, cpID, passStatus)
	assert.NoError(suite.T(), err)
}

func (suite *StatusMailDedupTestSuite) recipientIDs(status db.PassStatus) []uuid.UUID {
	rows, err := suite.queries.GetParticipantMailingInformation(suite.ctx, db.GetParticipantMailingInformationParams{
		ID:         dedupCoursePhaseID,
		PassStatus: db.NullPassStatus{PassStatus: status, Valid: true},
	})
	assert.NoError(suite.T(), err)
	ids := make([]uuid.UUID, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.CourseParticipationID)
	}
	return ids
}

func (suite *StatusMailDedupTestSuite) restrictedData(cpID uuid.UUID) string {
	var data string
	err := suite.conn.QueryRow(suite.ctx,
		`SELECT COALESCE(restricted_data::text, 'null') FROM course_phase_participation
		 WHERE course_phase_id = $1 AND course_participation_id = $2`,
		dedupCoursePhaseID, cpID).Scan(&data)
	assert.NoError(suite.T(), err)
	return data
}

// recordingSendMailFn installs a sendMailFn that records recipients and fails for failFor.
func (suite *StatusMailDedupTestSuite) recordingSendMailFn(failFor string) *[]string {
	sent := make([]string, 0)
	sendMailFn = func(
		courseMailingSettings mailingDTO.CourseMailingSettings,
		recipientAddress, subject, htmlBody string,
	) error {
		if recipientAddress == failFor {
			return errors.New("smtp rejected recipient")
		}
		sent = append(sent, recipientAddress)
		return nil
	}
	return &sent
}

func (suite *StatusMailDedupTestSuite) markStatusMailSent(cpID uuid.UUID, status string) {
	err := suite.queries.MarkStatusMailSent(suite.ctx, db.MarkStatusMailSentParams{
		CoursePhaseID:         dedupCoursePhaseID,
		CourseParticipationID: cpID,
		Status:                status,
		SentAt:                dedupSentAt,
	})
	assert.NoError(suite.T(), err)
}

func (suite *StatusMailDedupTestSuite) TestStatusMailDedup() {
	t := suite.T()

	// Both accepted; cp1 with a pre-existing restricted_data key, cp2 with NULL restricted_data.
	suite.setParticipationNull(dedupCP2, "passed")
	suite.setParticipation(dedupCP1, "passed", `{"foo": "bar"}`)

	// Initially both are recipients for the passed status.
	initial := suite.recipientIDs(db.PassStatusPassed)
	assert.Contains(t, initial, dedupCP1)
	assert.Contains(t, initial, dedupCP2)

	// Mark cp1 as mailed for "passed".
	suite.markStatusMailSent(dedupCP1, "passed")

	// cp1 is now excluded; cp2 (unmarked, previously NULL restricted_data) is still included.
	afterMark := suite.recipientIDs(db.PassStatusPassed)
	assert.NotContains(t, afterMark, dedupCP1)
	assert.Contains(t, afterMark, dedupCP2)

	// The pre-existing restricted_data key is preserved and the marker records the passed timestamp.
	cp1Data := suite.restrictedData(dedupCP1)
	assert.Contains(t, cp1Data, `"foo": "bar"`)
	assert.Contains(t, cp1Data, dedupSentAt)

	// Marking cp2 (NULL restricted_data) upgrades it to an object and excludes it.
	suite.markStatusMailSent(dedupCP2, "passed")
	assert.NotContains(t, suite.recipientIDs(db.PassStatusPassed), dedupCP2)

	// Idempotency: marking cp1 again keeps it excluded (single valid marker).
	suite.markStatusMailSent(dedupCP1, "passed")
	assert.NotContains(t, suite.recipientIDs(db.PassStatusPassed), dedupCP1)

	// Opposite-status non-interference: a failed participant that already has statusMailSentAt.passed
	// must still receive the failed status mail.
	suite.setParticipation(dedupCP1, "failed", `{"statusMailSentAt": {"passed": "2026-01-01T00:00:00Z"}}`)
	failedRecipients := suite.recipientIDs(db.PassStatusFailed)
	assert.Contains(t, failedRecipients, dedupCP1)
}

// TestSendStatusMailOnlyOncePerParticipant covers the send-and-mark loop: a repeated trigger for the
// same status must not mail anybody again.
func (suite *StatusMailDedupTestSuite) TestSendStatusMailOnlyOncePerParticipant() {
	t := suite.T()

	suite.setParticipation(dedupCP1, "passed", `{}`)
	suite.setParticipation(dedupCP2, "passed", `{}`)
	sent := suite.recordingSendMailFn("")

	report, err := SendStatusMailManualTrigger(suite.ctx, dedupCoursePhaseID, db.PassStatusPassed)
	suite.Require().NoError(err)

	sortedSent := append([]string(nil), *sent...)
	sort.Strings(sortedSent)
	assert.Equal(t, []string{dedupEmail1, dedupEmail2}, sortedSent)
	assert.Len(t, report.SuccessfulEmails, 2)
	assert.Empty(t, report.FailedEmails)
	assert.Empty(t, report.MarkFailures)

	// The markers are committed per participant, so the second trigger has no recipients left.
	secondSent := suite.recordingSendMailFn("")
	secondReport, err := SendStatusMailManualTrigger(suite.ctx, dedupCoursePhaseID, db.PassStatusPassed)
	suite.Require().NoError(err)
	assert.Empty(t, *secondSent)
	assert.Empty(t, secondReport.SuccessfulEmails)
}

// TestSendStatusMailRetriesFailedRecipients ensures a failed send leaves no marker behind, so the
// participant is picked up again by the next trigger while the successful one is not re-mailed.
func (suite *StatusMailDedupTestSuite) TestSendStatusMailRetriesFailedRecipients() {
	t := suite.T()

	suite.setParticipation(dedupCP1, "passed", `{}`)
	suite.setParticipation(dedupCP2, "passed", `{}`)
	failingRecipient := dedupEmail1

	sent := suite.recordingSendMailFn(failingRecipient)
	report, err := SendStatusMailManualTrigger(suite.ctx, dedupCoursePhaseID, db.PassStatusPassed)
	suite.Require().NoError(err)
	assert.Equal(t, []string{dedupEmail2}, *sent)
	assert.Equal(t, []string{failingRecipient}, report.FailedEmails)
	assert.Empty(t, report.MarkFailures)

	retried := suite.recordingSendMailFn("")
	retryReport, err := SendStatusMailManualTrigger(suite.ctx, dedupCoursePhaseID, db.PassStatusPassed)
	suite.Require().NoError(err)
	assert.Equal(t, []string{failingRecipient}, *retried)
	assert.Equal(t, []string{failingRecipient}, retryReport.SuccessfulEmails)
}

func TestStatusMailDedupTestSuite(t *testing.T) {
	suite.Run(t, new(StatusMailDedupTestSuite))
}
