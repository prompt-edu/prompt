package phaseconfig

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/infrastructure_setup/db/sqlc"
	"github.com/prompt-edu/prompt/servers/infrastructure_setup/phaseconfig/phaseconfigDTO"
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
	if got.SemesterTag != "" {
		t.Fatalf("default config = %+v, want empty semester tag", got)
	}
}

func TestUpsertCreatesAndUpdatesConfig(t *testing.T) {
	testDB, cleanup := setupPhaseConfigTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	service := NewService(testDB.Conn)

	created, err := service.Upsert(context.Background(), coursePhaseID, phaseconfigDTO.UpsertRequest{
		SemesterTag: "ios26",
	})
	if err != nil {
		t.Fatalf("Upsert create returned error: %v", err)
	}
	if created.SemesterTag != "ios26" {
		t.Fatalf("semester tag = %q, want ios26", created.SemesterTag)
	}

	updated, err := service.Upsert(context.Background(), coursePhaseID, phaseconfigDTO.UpsertRequest{
		SemesterTag: "ss27",
	})
	if err != nil {
		t.Fatalf("Upsert update returned error: %v", err)
	}
	if updated.SemesterTag != "ss27" {
		t.Fatalf("updated semester tag = %q, want ss27", updated.SemesterTag)
	}
}

// The tag is interpolated into provider resource names, where a stray space would
// become a separator: validation trims it, so storage has to keep the trimmed form.
func TestUpsertStoresTheTrimmedSemesterTag(t *testing.T) {
	testDB, cleanup := setupPhaseConfigTestDB(t)
	defer cleanup()

	coursePhaseID := uuid.New()
	service := NewService(testDB.Conn)

	created, err := service.Upsert(context.Background(), coursePhaseID, phaseconfigDTO.UpsertRequest{
		SemesterTag: "  ios26  ",
	})
	if err != nil {
		t.Fatalf("Upsert returned error: %v", err)
	}
	if created.SemesterTag != "ios26" {
		t.Fatalf("stored semester tag = %q, want ios26", created.SemesterTag)
	}

	reread, err := service.Get(context.Background(), coursePhaseID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if reread.SemesterTag != "ios26" {
		t.Fatalf("re-read semester tag = %q, want ios26", reread.SemesterTag)
	}
}

func TestUpsertRejectsAnInvalidSemesterTagAsValidation(t *testing.T) {
	testDB, cleanup := setupPhaseConfigTestDB(t)
	defer cleanup()

	service := NewService(testDB.Conn)

	_, err := service.Upsert(context.Background(), uuid.New(), phaseconfigDTO.UpsertRequest{
		SemesterTag: "WS 25/26",
	})
	if err == nil {
		t.Fatal("Upsert accepted a semester tag with spaces and slashes")
	}
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("error = %v, want it to wrap ErrValidation so the router answers 400", err)
	}
}
