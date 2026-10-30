package testutils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/ai/db/sqlc"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDB struct {
	Conn    *pgxpool.Pool
	Queries *db.Queries
}

// Resolved against this file, so suites in any package can call SetupTestDB.
func migrationPaths() ([]string, error) {
	_, thisFile, _, _ := runtime.Caller(0)
	pattern := filepath.Join(filepath.Dir(thisFile), "..", "db", "migration", "*.up.sql")
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("list migrations: %w", err)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no migrations found at %s", pattern)
	}
	sort.Strings(paths)
	return paths, nil
}

func SetupTestDB(ctx context.Context, seedPaths ...string) (*TestDB, func(), error) {
	request := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "prompt",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("start postgres container: %w", err)
	}
	terminate := func() { _ = container.Terminate(ctx) }

	host, err := container.Host(ctx)
	if err != nil {
		terminate()
		return nil, nil, fmt.Errorf("get container host: %w", err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		terminate()
		return nil, nil, fmt.Errorf("get container port: %w", err)
	}

	conn, err := connect(ctx, fmt.Sprintf(
		"postgres://testuser:testpass@%s:%s/prompt?sslmode=disable", host, port.Port(),
	))
	if err != nil {
		terminate()
		return nil, nil, err
	}

	migrations, err := migrationPaths()
	if err != nil {
		conn.Close()
		terminate()
		return nil, nil, err
	}
	for _, path := range append(migrations, seedPaths...) {
		if err := runSQLFile(ctx, conn, path); err != nil {
			conn.Close()
			terminate()
			return nil, nil, err
		}
	}

	cleanup := func() {
		conn.Close()
		terminate()
	}
	return &TestDB{Conn: conn, Queries: db.New(conn)}, cleanup, nil
}

func connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	var lastErr error
	for range 5 {
		conn, err := pgxpool.New(ctx, url)
		if err != nil {
			lastErr = err
		} else if lastErr = conn.Ping(ctx); lastErr == nil {
			return conn, nil
		} else {
			conn.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil, fmt.Errorf("connect to test database: %w", lastErr)
}

func runSQLFile(ctx context.Context, conn *pgxpool.Pool, path string) error {
	statements, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	// Argument-less Exec uses the simple protocol, which runs the file's statements as one batch.
	if _, err := conn.Exec(ctx, string(statements)); err != nil {
		return fmt.Errorf("apply %s: %w", path, err)
	}
	return nil
}
