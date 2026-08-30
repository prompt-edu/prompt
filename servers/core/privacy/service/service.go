package service

import (
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type PrivacyService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

func NewPrivacyService(queries db.Queries, conn *pgxpool.Pool) *PrivacyService {
	return &PrivacyService{
		queries: queries,
		conn:    conn,
	}
}
