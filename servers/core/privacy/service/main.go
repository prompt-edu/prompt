package service

import (
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type PrivacyService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var PrivacyServiceSingleton *PrivacyService

func InitPrivacyService(queries db.Queries, conn *pgxpool.Pool) {

	PrivacyServiceSingleton = &PrivacyService{
		queries: queries,
		conn:    conn,
	}

}
