package permissionValidation

import (
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type ValidationService struct {
	queries db.Queries
}

func NewValidationService(queries db.Queries) *ValidationService {
	return &ValidationService{
		queries: queries,
	}
}
