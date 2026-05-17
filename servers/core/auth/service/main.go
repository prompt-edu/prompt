package service

import (
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)


func InitAuthService(queries db.Queries, conn *pgxpool.Pool) {
	AuthServiceSingleton = &AuthService{
		queries: queries,
		conn:    conn,
	}
}

type AuthService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var AuthServiceSingleton *AuthService
