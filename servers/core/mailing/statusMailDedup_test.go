package mailing

import (
	"context"
	"errors"
	"net/mail"
	"sort"
	"sync"
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

// claimIDs claims every participant that is still eligible for the given status and returns them.
func (suite *StatusMailDedupTestSuite) claimIDs(status db.PassStatus) []uuid.UUID {
	rows, err := suite.queries.ClaimStatusMailRecipients(suite.ctx, db.ClaimStatusMailRecipientsParams{
		CoursePhaseID: dedupCoursePhaseID,
		Status:        string(status),
		SentAt:        dedupSentAt,
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

func (suite *StatusMailDedupTestSuite) TestStatusMailClaim() {
	t := suite.T()

	// Both accepted; cp1 with a pre-existing restricted_data key, cp2 with NULL restricted_data.
	suite.setParticipationNull(dedupCP2, "passed")
	suite.setParticipation(dedupCP1, "passed", `{"foo": "bar"}`)

	// The first claim covers both, including the one with NULL restricted_data.
	initial := suite.claimIDs(db.PassStatusPassed)
	assert.Contains(t, initial, dedupCP1)
	assert.Contains(t, initial, dedupCP2)

	// The pre-existing restricted_data key is preserved and the marker records the passed timestamp.
	cp1Data := suite.restrictedData(dedupCP1)
	assert.Contains(t, cp1Data, `"foo": "bar"`)
	assert.Contains(t, cp1Data, dedupSentAt)

	// A second claim finds nobody left.
	assert.Empty(t, suite.claimIDs(db.PassStatusPassed))

	// Opposite-status non-interference: a failed participant that already has statusMailSentAt.passed
	// is still claimed for the failed status mail.
	suite.setParticipation(dedupCP1, "failed", `{"statusMailSentAt": {"passed": "2026-01-01T00:00:00Z"}}`)
	assert.Contains(t, suite.claimIDs(db.PassStatusFailed), dedupCP1)
}

// TestStatusMailClaimNormalizesNonObjectRestrictedData covers restricted_data values that are not
// objects, where a plain `||` merge would not produce a readable statusMailSentAt marker.
func (suite *StatusMailDedupTestSuite) TestStatusMailClaimNormalizesNonObjectRestrictedData() {
	t := suite.T()

	for _, restrictedData := range []string{`null`, `"scalar"`, `[1, 2]`, `{"statusMailSentAt": "scalar"}`} {
		suite.setParticipation(dedupCP1, "passed", restrictedData)
		suite.setParticipation(dedupCP2, "failed", `{}`)

		assert.Contains(t, suite.claimIDs(db.PassStatusPassed), dedupCP1, restrictedData)
		assert.Contains(t, suite.restrictedData(dedupCP1), dedupSentAt, restrictedData)
		assert.Empty(t, suite.claimIDs(db.PassStatusPassed), restrictedData)
	}
}

// TestStatusMailClaimRestrictedToSelectedRecipients covers the recipient-list variant of the claim.
func (suite *StatusMailDedupTestSuite) TestStatusMailClaimRestrictedToSelectedRecipients() {
	t := suite.T()

	suite.setParticipation(dedupCP1, "passed", `{}`)
	suite.setParticipation(dedupCP2, "passed", `{}`)

	claimed, err := suite.queries.ClaimStatusMailRecipients(suite.ctx, db.ClaimStatusMailRecipientsParams{
		CoursePhaseID:          dedupCoursePhaseID,
		Status:                 string(db.PassStatusPassed),
		SentAt:                 dedupSentAt,
		CourseParticipationIds: []uuid.UUID{dedupCP1},
	})
	suite.Require().NoError(err)
	suite.Require().Len(claimed, 1)
	assert.Equal(t, dedupCP1, claimed[0].CourseParticipationID)

	// cp2 was left untouched and is still claimable.
	assert.Equal(t, []uuid.UUID{dedupCP2}, suite.claimIDs(db.PassStatusPassed))
}

// TestSendStatusMailConcurrentTriggersMailOnce covers two triggers for the same phase and status that
// overlap: the second one runs while the first is still sending, so it must find no recipients left.
func (suite *StatusMailDedupTestSuite) TestSendStatusMailConcurrentTriggersMailOnce() {
	t := suite.T()

	suite.setParticipation(dedupCP1, "passed", `{}`)
	suite.setParticipation(dedupCP2, "passed", `{}`)

	var mu sync.Mutex
	sent := make([]string, 0)
	sending := make(chan struct{})
	secondTriggerDone := make(chan struct{})

	sendMailFn = func(
		courseMailingSettings mailingDTO.CourseMailingSettings,
		recipientAddress, subject, htmlBody string,
	) error {
		mu.Lock()
		sent = append(sent, recipientAddress)
		isFirstSend := len(sent) == 1
		mu.Unlock()

		if isFirstSend {
			close(sending)
			<-secondTriggerDone
		}
		return nil
	}

	firstReports := make(chan mailingDTO.MailingReport, 1)
	go func() {
		report, err := SendStatusMailManualTrigger(suite.ctx, dedupCoursePhaseID, db.PassStatusPassed, nil)
		assert.NoError(t, err)
		firstReports <- report
	}()

	<-sending
	secondReport, err := SendStatusMailManualTrigger(suite.ctx, dedupCoursePhaseID, db.PassStatusPassed, nil)
	suite.Require().NoError(err)
	close(secondTriggerDone)

	assert.Empty(t, secondReport.SuccessfulEmails)
	assert.Len(t, (<-firstReports).SuccessfulEmails, 2)

	mu.Lock()
	defer mu.Unlock()
	sort.Strings(sent)
	assert.Equal(t, []string{dedupEmail1, dedupEmail2}, sent)
}

// TestSendStatusMailOnlyOncePerParticipant covers the claim-and-send loop: a repeated trigger for the
// same status must not mail anybody again.
func (suite *StatusMailDedupTestSuite) TestSendStatusMailOnlyOncePerParticipant() {
	t := suite.T()

	suite.setParticipation(dedupCP1, "passed", `{}`)
	suite.setParticipation(dedupCP2, "passed", `{}`)
	sent := suite.recordingSendMailFn("")

	report, err := SendStatusMailManualTrigger(suite.ctx, dedupCoursePhaseID, db.PassStatusPassed, nil)
	suite.Require().NoError(err)

	sortedSent := append([]string(nil), *sent...)
	sort.Strings(sortedSent)
	assert.Equal(t, []string{dedupEmail1, dedupEmail2}, sortedSent)
	assert.Len(t, report.SuccessfulEmails, 2)
	assert.Empty(t, report.FailedEmails)

	// The claims are committed per trigger, so the second trigger has no recipients left.
	secondSent := suite.recordingSendMailFn("")
	secondReport, err := SendStatusMailManualTrigger(suite.ctx, dedupCoursePhaseID, db.PassStatusPassed, nil)
	suite.Require().NoError(err)
	assert.Empty(t, *secondSent)
	assert.Empty(t, secondReport.SuccessfulEmails)
}

// TestSendStatusMailRetriesFailedRecipients ensures a failed send releases its claim again, so the
// participant is picked up by the next trigger while the successful one is not re-mailed.
func (suite *StatusMailDedupTestSuite) TestSendStatusMailRetriesFailedRecipients() {
	t := suite.T()

	suite.setParticipation(dedupCP1, "passed", `{"foo": "bar"}`)
	suite.setParticipation(dedupCP2, "passed", `{}`)
	failingRecipient := dedupEmail1

	sent := suite.recordingSendMailFn(failingRecipient)
	report, err := SendStatusMailManualTrigger(suite.ctx, dedupCoursePhaseID, db.PassStatusPassed, nil)
	suite.Require().NoError(err)
	assert.Equal(t, []string{dedupEmail2}, *sent)
	assert.Equal(t, []string{failingRecipient}, report.FailedEmails)

	// Releasing the claim removes only the marker for this status, not the other restricted_data keys.
	cp1Data := suite.restrictedData(dedupCP1)
	assert.NotContains(t, cp1Data, `"passed"`)
	assert.Contains(t, cp1Data, `"foo": "bar"`)

	retried := suite.recordingSendMailFn("")
	retryReport, err := SendStatusMailManualTrigger(suite.ctx, dedupCoursePhaseID, db.PassStatusPassed, nil)
	suite.Require().NoError(err)
	assert.Equal(t, []string{failingRecipient}, *retried)
	assert.Equal(t, []string{failingRecipient}, retryReport.SuccessfulEmails)
}

func TestStatusMailDedupTestSuite(t *testing.T) {
	suite.Run(t, new(StatusMailDedupTestSuite))
}
