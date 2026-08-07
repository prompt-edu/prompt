package auditLog

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/audit"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/auditLog/auditLogDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	seededPhaseID  = "11111111-1111-1111-1111-111111111111"
	seededCourseID = "22222222-2222-2222-2222-222222222222"
	otherCourseID  = "33333333-3333-3333-3333-333333333333"
)

type AuditLogTestSuite struct {
	suite.Suite
	ctx     context.Context
	cleanup func()
	conn    *pgxpool.Pool
	queries *db.Queries
	sink    *DBSink
}

func TestAuditLogTestSuite(t *testing.T) {
	suite.Run(t, new(AuditLogTestSuite))
}

func (s *AuditLogTestSuite) SetupSuite() {
	os.Setenv("AUDIT_ENABLED", "true")
	s.ctx = context.Background()

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(s.ctx, "../database_dumps/audit_log_test.sql",
		func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	require.NoError(s.T(), err)

	s.cleanup = cleanup
	s.conn = testDB.Conn
	s.queries = testDB.Queries
	s.sink = NewDBSink(*testDB.Queries)
	AuditLogServiceSingleton = &AuditLogService{queries: *testDB.Queries, conn: testDB.Conn}
}

func (s *AuditLogTestSuite) TearDownSuite() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *AuditLogTestSuite) SetupTest() {
	_, err := s.conn.Exec(s.ctx, "TRUNCATE audit_log")
	require.NoError(s.T(), err)
}

func (s *AuditLogTestSuite) countRows() int {
	var n int
	require.NoError(s.T(), s.conn.QueryRow(s.ctx, "SELECT count(*) FROM audit_log").Scan(&n))
	return n
}

// insertRaw inserts a row directly, controlling created_at (which the sink never
// sets) — used to seed ordering/retention scenarios.
func (s *AuditLogTestSuite) insertRaw(createdAt time.Time, courseID, actorRole, outcome, action string) {
	_, err := s.conn.Exec(s.ctx,
		`INSERT INTO audit_log (created_at, course_id, actor_role, outcome, action, source_service)
		 VALUES ($1, $2, $3, $4, $5, 'core')`,
		createdAt, pgUUID(courseID), actorRole, outcome, action)
	require.NoError(s.T(), err)
}

func (s *AuditLogTestSuite) TestDBSink_ResolvesCourseAndSnapshotsEntity() {
	err := s.sink.Record(s.ctx, audit.Event{
		ActorID:       "44444444-4444-4444-4444-444444444444",
		ActorName:     "Ada Lovelace",
		Action:        "Published grades",
		CoursePhaseID: seededPhaseID, // no CourseID -> must be resolved
		EntityType:    "grade",
		EntityID:      "g1",
		EntityName:    "team Alpha",
	})
	require.NoError(s.T(), err)

	var courseID pgtype.UUID
	var entityName pgtype.Text
	require.NoError(s.T(), s.conn.QueryRow(s.ctx,
		"SELECT course_id, entity_name FROM audit_log LIMIT 1").Scan(&courseID, &entityName))
	require.Equal(s.T(), seededCourseID, uuidToString(courseID))
	require.Equal(s.T(), "team Alpha", entityName.String)
}

func (s *AuditLogTestSuite) TestAntiUpdateTriggerRejectsUpdate() {
	require.NoError(s.T(), s.sink.Record(s.ctx, audit.Event{Action: "Created slot"}))

	_, err := s.conn.Exec(s.ctx, "UPDATE audit_log SET action = 'tampered'")
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "append-only")
}

func (s *AuditLogTestSuite) TestRecordTx_RollsBackWithTransaction() {
	tx, err := s.conn.Begin(s.ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), RecordTx(s.ctx, s.queries.WithTx(tx), audit.Event{Action: "Rolled back"}))
	require.NoError(s.T(), tx.Rollback(s.ctx))
	require.Equal(s.T(), 0, s.countRows())

	tx2, err := s.conn.Begin(s.ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), RecordTx(s.ctx, s.queries.WithTx(tx2), audit.Event{Action: "Committed"}))
	require.NoError(s.T(), tx2.Commit(s.ctx))
	require.Equal(s.T(), 1, s.countRows())
}

func (s *AuditLogTestSuite) TestDeleteExpiredAuditEntries() {
	now := time.Now()
	s.insertRaw(now.AddDate(0, 0, -40), seededCourseID, "Lecturer", "success", "old")
	s.insertRaw(now.AddDate(0, 0, -1), seededCourseID, "Lecturer", "success", "recent")

	cutoff := now.AddDate(0, 0, -30)
	require.NoError(s.T(), s.queries.DeleteExpiredAuditEntries(s.ctx, pgtype.Timestamptz{Time: cutoff, Valid: true}))
	require.Equal(s.T(), 1, s.countRows())
}

func (s *AuditLogTestSuite) TestListAuditLog_CourseScopeAndOutcomeFilter() {
	now := time.Now()
	s.insertRaw(now.Add(-3*time.Minute), seededCourseID, "Lecturer", "success", "a")
	s.insertRaw(now.Add(-2*time.Minute), seededCourseID, "Student", "denied", "b")
	s.insertRaw(now.Add(-1*time.Minute), otherCourseID, "Lecturer", "success", "c")

	// Course scope: only the seeded course's rows.
	page, err := AuditLogServiceSingleton.ListAuditLog(s.ctx, auditLogDTO.ListFilters{CourseID: seededCourseID})
	require.NoError(s.T(), err)
	require.Len(s.T(), page.Entries, 2)
	for _, e := range page.Entries {
		require.Equal(s.T(), seededCourseID, e.CourseID)
	}

	// Outcome filter within the course.
	denied, err := AuditLogServiceSingleton.ListAuditLog(s.ctx, auditLogDTO.ListFilters{CourseID: seededCourseID, Outcome: "denied"})
	require.NoError(s.T(), err)
	require.Len(s.T(), denied.Entries, 1)
	require.Equal(s.T(), "b", denied.Entries[0].Action)

	// Global (no course filter) sees everything.
	global, err := AuditLogServiceSingleton.ListAuditLog(s.ctx, auditLogDTO.ListFilters{})
	require.NoError(s.T(), err)
	require.Len(s.T(), global.Entries, 3)
}

func (s *AuditLogTestSuite) TestListAuditLog_KeysetPagination() {
	now := time.Now()
	for i := 0; i < 5; i++ {
		s.insertRaw(now.Add(time.Duration(i)*time.Minute), seededCourseID, "Lecturer", "success", "row")
	}

	seen := map[string]bool{}
	filters := auditLogDTO.ListFilters{CourseID: seededCourseID, Limit: 2}
	pages := 0
	var prev *time.Time
	for {
		page, err := AuditLogServiceSingleton.ListAuditLog(s.ctx, filters)
		require.NoError(s.T(), err)
		for _, e := range page.Entries {
			require.False(s.T(), seen[e.ID], "entry appeared on two pages")
			seen[e.ID] = true
			// Strictly descending by created_at across pages.
			if prev != nil {
				require.False(s.T(), e.CreatedAt.After(*prev))
			}
			ts := e.CreatedAt
			prev = &ts
		}
		pages++
		if page.NextCursor == nil {
			break
		}
		filters.CursorCreatedAt = &page.NextCursor.CreatedAt
		filters.CursorID = page.NextCursor.ID
		require.LessOrEqual(s.T(), pages, 5)
	}
	require.Len(s.T(), seen, 5)
	require.Equal(s.T(), 3, pages) // 2 + 2 + 1
}
