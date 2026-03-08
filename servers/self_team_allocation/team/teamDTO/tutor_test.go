package teamDTO

import (
	"testing"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/stretchr/testify/require"
)

func TestGetTutorDTOFromDBModel(t *testing.T) {
	dbTutor := db.Tutor{
		CoursePhaseID:         uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		CourseParticipationID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		FirstName:             "Tara",
		LastName:              "Tutor",
		TeamID:                uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
	}

	dto := GetTutorDTOFromDBModel(dbTutor)

	require.Equal(t, "Tara", dto.FirstName)
	require.Equal(t, dbTutor.CourseParticipationID, dto.CourseParticipationID)
}

func TestGetTutorDTOsFromDBModels(t *testing.T) {
	dbTutors := []db.Tutor{
		{
			CoursePhaseID:         uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			CourseParticipationID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			FirstName:             "Tim",
			LastName:              "Tutor",
			TeamID:                uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		},
	}

	tutors := GetTutorDTOsFromDBModels(dbTutors)

	require.Len(t, tutors, 1)
	require.Equal(t, "Tim", tutors[0].FirstName)
}
