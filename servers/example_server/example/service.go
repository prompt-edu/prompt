package example

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/example_server/db/sqlc"
)

// ExampleService holds everything the module's business logic needs. Collaborators
// arrive through the constructor, so a test can build the service on its own and
// `main.go` stays the only place that knows the wiring order.
type ExampleService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

func NewExampleService(queries db.Queries, conn *pgxpool.Pool) *ExampleService {
	return &ExampleService{
		queries: queries,
		conn:    conn,
	}
}

// GetExampleInfo is a placeholder for real business logic. Read the database
// through the receiver's fields (`s.queries`, `s.conn`), never through a global.
func (s *ExampleService) GetExampleInfo(ctx context.Context, coursePhaseID uuid.UUID) (string, error) {
	return "This is a message provided by the example service", nil
}
