package applicationDTO

import (
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
)

type PutAssessment struct {
	Score          pgtype.Int4    `json:"score" swaggertype:"integer"`
	RestrictedData meta.MetaData  `json:"restrictedData"`
	PassStatus     *db.PassStatus `json:"passStatus"`
}
