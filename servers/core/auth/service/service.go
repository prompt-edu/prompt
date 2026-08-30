package service

import (
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type AuthService struct {
	queries db.Queries
}

func NewAuthService(queries db.Queries) *AuthService {
	return &AuthService{queries: queries}
}
