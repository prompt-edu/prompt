package callDTO

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func textOf(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func intOf(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}
	return &value.Int32
}

func timeOf(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}
