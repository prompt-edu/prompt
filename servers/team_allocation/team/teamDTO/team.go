package teamDTO

import (
	"encoding/json"

	"github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

func GetTeamDTOFromDBModel(dbTeam db.Team) promptTypes.Team {
	return promptTypes.Team{
		ID:   dbTeam.ID,
		Name: dbTeam.Name,
	}
}

func GetTeamDTOsFromDBModels(dbTeams []db.Team) []promptTypes.Team {
	teams := make([]promptTypes.Team, 0, len(dbTeams))
	for _, dbTeam := range dbTeams {
		teams = append(teams, GetTeamDTOFromDBModel(dbTeam))
	}
	return teams
}

func GetTeamsWithMembersDTOFromDBModel(dbTeams []db.GetTeamsWithMembersRow) ([]promptTypes.Team, error) {
	teams := make([]promptTypes.Team, 0, len(dbTeams))
	for _, dbTeam := range dbTeams {
		t, err := GetTeamWithMembersDTOFromDBModel(dbTeam)
		if err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}

func GetTeamWithMembersDTOFromDBModel(dbTeam db.GetTeamsWithMembersRow) (promptTypes.Team, error) {
	var members []promptTypes.Person
	var tutors []promptTypes.Person
	// unmarshal the JSON blob into your slice of structs
	if err := json.Unmarshal(dbTeam.TeamMembers, &members); err != nil {
		return promptTypes.Team{}, err
	}
	if err := json.Unmarshal(dbTeam.TeamTutors, &tutors); err != nil {
		return promptTypes.Team{}, err
	}

	return promptTypes.Team{
		ID:      dbTeam.ID,
		Name:    dbTeam.Name,
		Members: members,
		Tutors:  tutors,
	}, nil
}
