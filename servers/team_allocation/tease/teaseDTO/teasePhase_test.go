package teaseDTO

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestTeasePhaseMarshalsDeadlineWithoutZone(t *testing.T) {
	phase := TeasePhase{
		CoursePhaseID: uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02"),
		SemesterName:  "ios2426-iPraktikum-Team Allocation",
		KickoffSubmissionPeriodEnd: pgtype.Timestamp{
			Time:  time.Date(2026, 1, 15, 12, 34, 56, 123456000, time.UTC),
			Valid: true,
		},
	}

	encoded, err := json.Marshal(phase)
	require.NoError(t, err)

	require.JSONEq(t, `{
		"id": "4179d58a-d00d-4fa7-94a5-397bc69fab02",
		"semesterName": "ios2426-iPraktikum-Team Allocation",
		"kickoffSubmissionPeriodEnd": "2026-01-15T12:34:56.123456"
	}`, string(encoded))
}
