package example

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/example_server/db/sqlc"
)

type ExampleService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var ExampleServiceSingleton *ExampleService

func GetExampleInfo(ctx context.Context, coursePhaseID uuid.UUID) (string, error) {
	return "This is a message provided by the example service", nil
}
