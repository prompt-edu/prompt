package teamDTO

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/stretchr/testify/require"
)

func TestGetTutorAssignmentsFromDBModels(t *testing.T) {
	tutorID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	teamID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	assignments := GetTutorAssignmentsFromDBModels([]db.GetTutorAssignmentsByCoursePhaseRow{
		{CourseParticipationID: tutorID, UniversityLogin: pgtype.Text{String: "ab12cde", Valid: true}, TeamID: teamID},
		{CourseParticipationID: tutorID, TeamID: teamID},
	})

	require.Equal(t, []TutorAssignment{
		{ID: tutorID, UniversityLogin: "ab12cde", TeamID: teamID},
		{ID: tutorID, UniversityLogin: "", TeamID: teamID},
	}, assignments)
}

func TestUniversityLoginParamStoresNullForAnEmptyLogin(t *testing.T) {
	require.False(t, UniversityLoginParam("").Valid)
	require.Equal(t, pgtype.Text{String: "ab12cde", Valid: true}, UniversityLoginParam("ab12cde"))
}
