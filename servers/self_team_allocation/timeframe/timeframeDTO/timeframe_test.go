package timeframeDTO

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/stretchr/testify/require"
)

func TestGetTimeframeDTOFromDBModel(t *testing.T) {
	start := time.Now()
	end := start.Add(24 * time.Hour)

	row := db.GetTimeframeRow{
		Starttime: pgtype.Timestamp{Time: start, Valid: true},
		Endtime:   pgtype.Timestamp{Time: end, Valid: true},
	}

	dto := GetTimeframeDTOFromDBModel(row)

	require.True(t, dto.TimeframeSet)
	require.Equal(t, start, dto.StartTime)
	require.Equal(t, end, dto.EndTime)
}
