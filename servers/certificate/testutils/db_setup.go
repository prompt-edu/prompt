package testutils

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/ls1intum/prompt2/servers/certificate/db/sqlc"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDB struct {
	Conn    *pgxpool.Pool
	Queries *db.Queries
}

func SetupTestDB(ctx context.Context, sqlDumpPath string) (*TestDB, func(), error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
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
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not start container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, fmt.Errorf("could not get container host: %w", err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, fmt.Errorf("could not get container port: %w", err)
	}
	dbURL := fmt.Sprintf("postgres://testuser:testpass@%s:%s/prompt?sslmode=disable", host, port.Port())

	var conn *pgxpool.Pool
	var pingErr error
	for i := 0; i < 5; i++ {
		conn, err = pgxpool.New(ctx, dbURL)
		if err == nil {
			pingErr = conn.Ping(ctx)
			if pingErr == nil {
				break
			}
			conn.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to connect to the database after retries: %w", err)
	}
	if pingErr != nil {
		_ = container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to ping the database after retries: %w", pingErr)
	}

	if err := runSQLDump(conn, sqlDumpPath); err != nil {
		conn.Close()
		_ = container.Terminate(ctx)
		return nil, nil, fmt.Errorf("failed to run SQL dump: %w", err)
	}

	queries := db.New(conn)

	cleanup := func() {
		conn.Close()
		_ = container.Terminate(ctx)
	}

	return &TestDB{
		Conn:    conn,
		Queries: queries,
	}, cleanup, nil
}

func runSQLDump(conn *pgxpool.Pool, path string) error {
	dump, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read SQL dump file: %w", err)
	}

	_, err = conn.Exec(context.Background(), string(dump))
	if err != nil {
		return fmt.Errorf("could not execute SQL dump: %w", err)
	}

	return nil
}
