package teamDTO

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/stretchr/testify/require"
)

func sampleTeamRow(t *testing.T) db.GetTeamsWithStudentNamesRow {
	t.Helper()

	membersJSON, err := json.Marshal([]promptTypes.Person{
		{ID: uuid.New(), FirstName: "Alice", LastName: "Anderson"},
	})
	require.NoError(t, err)

	tutorsJSON, err := json.Marshal([]promptTypes.Person{
		{ID: uuid.New(), FirstName: "Terry", LastName: "Tutor"},
	})
	require.NoError(t, err)

	return db.GetTeamsWithStudentNamesRow{
		ID:          uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Name:        "Alpha",
		TeamMembers: membersJSON,
		TeamTutors:  tutorsJSON,
	}
}

func TestGetTeamDTOFromDBModel(t *testing.T) {
	row := sampleTeamRow(t)

	team, err := GetTeamDTOFromDBModel(row)
	require.NoError(t, err)

	require.Equal(t, "Alpha", team.Name)
	require.Len(t, team.Members, 1)
	require.Equal(t, "Alice", team.Members[0].FirstName)
	require.Len(t, team.Tutors, 1)
}

func TestGetTeamWithFullNamesByIdDTOFromDBModel(t *testing.T) {
	source := sampleTeamRow(t)
	row := db.GetTeamWithStudentNamesByTeamIDRow(source)

	team, err := GetTeamWithFullNamesByIdDTOFromDBModel(row)
	require.NoError(t, err)
	require.Equal(t, source.ID, team.ID)
	require.Len(t, team.Tutors, 1)
}

func TestGetTeamWithFullNameDTOsFromDBModels(t *testing.T) {
	row := sampleTeamRow(t)

	teams, err := GetTeamWithFullNameDTOsFromDBModels([]db.GetTeamsWithStudentNamesRow{row})
	require.NoError(t, err)
	require.Len(t, teams, 1)
}
