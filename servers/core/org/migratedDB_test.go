package org

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

// setupMigratedTestDB starts a test database from the real up migrations instead of a
// hand-written dump. The org table's constraints (the slug check, the slug uniqueness,
// and the restricting foreign keys) are part of the behavior under test, so the schema
// must not drift from what ships. The first migration is loaded as the SDK's dump, the
// rest are applied in numbered order.
func setupMigratedTestDB(ctx context.Context) (*sdkTestUtils.TestDB[*db.Queries], func(), error) {
	migrations, err := filepath.Glob(filepath.Join("..", "db", "migration", "*.up.sql"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list migrations: %w", err)
	}
	if len(migrations) == 0 {
		return nil, nil, errors.New("no migrations found")
	}
	sort.Strings(migrations)

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(ctx, migrations[0], func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		return nil, nil, err
	}

	for _, path := range migrations[1:] {
		statements, err := os.ReadFile(path)
		if err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("failed to read %s: %w", path, err)
		}
		// Argument-less Exec uses the simple protocol, so a migration's multiple
		// statements (including its BEGIN/COMMIT) run as one batch.
		if _, err := testDB.Conn.Exec(ctx, string(statements)); err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("failed to apply %s: %w", path, err)
		}
	}
	return testDB, cleanup, nil
}
