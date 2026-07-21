package mailing

import (
	"context"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type StatusMailDedupTestSuite struct {
	suite.Suite
	ctx     context.Context
	cleanup func()
	queries *db.Queries
	conn    *pgxpool.Pool
}

// Known IDs from database_dumps/application_administration.sql
var (
	dedupCoursePhaseID = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	dedupCP1           = uuid.MustParse("82d7efae-d545-4cc5-9b94-5d0ee1e50d25")
	dedupCP2           = uuid.MustParse("32aa070e-67c3-4a69-852a-ba3b5e849a4d")
)

func (suite *StatusMailDedupTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/application_administration.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.queries = testDB.Queries
	suite.conn = testDB.Conn
}

func (suite *StatusMailDedupTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *StatusMailDedupTestSuite) setParticipation(cpID uuid.UUID, passStatus string, restrictedData string) {
	_, err := suite.conn.Exec(suite.ctx,
		`UPDATE course_phase_participation SET pass_status = $3::pass_status, restricted_data = $4::jsonb
		 WHERE course_phase_id = $1 AND course_participation_id = $2`,
		dedupCoursePhaseID, cpID, passStatus, restrictedData)
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
	err := suite.queries.MarkStatusMailSent(suite.ctx, db.MarkStatusMailSentParams{
		CoursePhaseID:         dedupCoursePhaseID,
		CourseParticipationID: dedupCP1,
		Status:                "passed",
	})
	assert.NoError(t, err)

	// cp1 is now excluded; cp2 (unmarked, previously NULL restricted_data) is still included.
	afterMark := suite.recipientIDs(db.PassStatusPassed)
	assert.NotContains(t, afterMark, dedupCP1)
	assert.Contains(t, afterMark, dedupCP2)

	// The pre-existing restricted_data key is preserved and the marker is recorded.
	cp1Data := suite.restrictedData(dedupCP1)
	assert.Contains(t, cp1Data, `"foo": "bar"`)
	assert.Contains(t, cp1Data, "statusMailSentAt")

	// Marking cp2 (NULL restricted_data) upgrades it to an object and excludes it.
	err = suite.queries.MarkStatusMailSent(suite.ctx, db.MarkStatusMailSentParams{
		CoursePhaseID:         dedupCoursePhaseID,
		CourseParticipationID: dedupCP2,
		Status:                "passed",
	})
	assert.NoError(t, err)
	assert.NotContains(t, suite.recipientIDs(db.PassStatusPassed), dedupCP2)

	// Idempotency: marking cp1 again keeps it excluded (single valid marker).
	err = suite.queries.MarkStatusMailSent(suite.ctx, db.MarkStatusMailSentParams{
		CoursePhaseID:         dedupCoursePhaseID,
		CourseParticipationID: dedupCP1,
		Status:                "passed",
	})
	assert.NoError(t, err)
	assert.NotContains(t, suite.recipientIDs(db.PassStatusPassed), dedupCP1)

	// Opposite-status non-interference: a failed participant that already has statusMailSentAt.passed
	// must still receive the failed status mail.
	suite.setParticipation(dedupCP1, "failed", `{"statusMailSentAt": {"passed": "2026-01-01T00:00:00Z"}}`)
	failedRecipients := suite.recipientIDs(db.PassStatusFailed)
	assert.Contains(t, failedRecipients, dedupCP1)
}

func (suite *StatusMailDedupTestSuite) setParticipationNull(cpID uuid.UUID, passStatus string) {
	_, err := suite.conn.Exec(suite.ctx,
		`UPDATE course_phase_participation SET pass_status = $3::pass_status, restricted_data = NULL
		 WHERE course_phase_id = $1 AND course_participation_id = $2`,
		dedupCoursePhaseID, cpID, passStatus)
	assert.NoError(suite.T(), err)
}

func TestStatusMailDedupTestSuite(t *testing.T) {
	suite.Run(t, new(StatusMailDedupTestSuite))
}
