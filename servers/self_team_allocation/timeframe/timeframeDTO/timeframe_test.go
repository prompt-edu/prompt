package timeframeDTO

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/stretchr/testify/require"
)

func TestGetTimeframeDTOFromDBModel(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	require.NoError(t, err)

	start := time.Date(2026, 1, 15, 12, 34, 56, 123456000, berlin)
	end := start.Add(24 * time.Hour)

	row := db.GetTimeframeRow{
		Starttime: pgtype.Timestamptz{Time: start, Valid: true},
		Endtime:   pgtype.Timestamptz{Time: end, Valid: true},
	}

	dto := GetTimeframeDTOFromDBModel(row)

	require.True(t, dto.TimeframeSet)
	require.True(t, start.Equal(dto.StartTime))
	require.True(t, end.Equal(dto.EndTime))
	require.Equal(t, time.UTC, dto.StartTime.Location())
	require.Equal(t, time.UTC, dto.EndTime.Location())
}
