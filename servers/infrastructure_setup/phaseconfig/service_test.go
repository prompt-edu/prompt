package phaseconfig

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
)

func setupPhaseConfigTestDB(t *testing.T) (*sdkTestUtils.TestDB[*db.Queries], func()) {
	t.Helper()

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(context.Background(), "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries {
		return db.New(conn)
	})
	if err != nil {
		t.Fatalf("setup test db: %v", err)
	}
	return testDB, cleanup
}

func TestGetReturnsDefaultWhenConfigMissing(t *testing.T) {
	testDB, cleanup := setupPhaseConfigTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	service := NewService(testDB.Conn)

	got, err := service.Get(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.CoursePhaseID != coursePhaseID {
		t.Fatalf("CoursePhaseID = %s, want %s", got.CoursePhaseID, coursePhaseID)
	}
	if got.TeamSourceCoursePhaseID != nil || got.StudentSourceCoursePhaseID != nil || got.SemesterTag != "" {
		t.Fatalf("default config = %+v, want empty sources and semester tag", got)
	}
}

func TestUpsertCreatesAndUpdatesConfig(t *testing.T) {
	testDB, cleanup := setupPhaseConfigTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	teamSourceID := uuid.New()
	studentSourceID := uuid.New()
	service := NewService(testDB.Conn)

	created, err := service.Upsert(context.Background(), coursePhaseID, UpsertRequest{
		TeamSourceCoursePhaseID:    &teamSourceID,
		StudentSourceCoursePhaseID: &studentSourceID,
		SemesterTag:                "ios26",
	})
	if err != nil {
		t.Fatalf("Upsert create returned error: %v", err)
	}
	if created.TeamSourceCoursePhaseID == nil || *created.TeamSourceCoursePhaseID != teamSourceID {
		t.Fatalf("team source = %v, want %s", created.TeamSourceCoursePhaseID, teamSourceID)
	}
	if created.StudentSourceCoursePhaseID == nil || *created.StudentSourceCoursePhaseID != studentSourceID {
		t.Fatalf("student source = %v, want %s", created.StudentSourceCoursePhaseID, studentSourceID)
	}
	if created.SemesterTag != "ios26" {
		t.Fatalf("semester tag = %q, want ios26", created.SemesterTag)
	}

	updated, err := service.Upsert(context.Background(), coursePhaseID, UpsertRequest{
		TeamSourceCoursePhaseID: nil,
		SemesterTag:             "ss27",
	})
	if err != nil {
		t.Fatalf("Upsert update returned error: %v", err)
	}
	if updated.TeamSourceCoursePhaseID != nil || updated.StudentSourceCoursePhaseID != nil {
		t.Fatalf("updated sources = team %v student %v, want nil", updated.TeamSourceCoursePhaseID, updated.StudentSourceCoursePhaseID)
	}
	if updated.SemesterTag != "ss27" {
		t.Fatalf("updated semester tag = %q, want ss27", updated.SemesterTag)
	}
}
