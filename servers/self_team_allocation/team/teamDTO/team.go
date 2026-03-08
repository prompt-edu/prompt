package teamDTO

import (
	"encoding/json"

	"github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
)

func GetTeamDTOFromDBModel(dbTeam db.GetTeamsWithStudentNamesRow) (promptTypes.Team, error) {
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

func GetTeamWithFullNamesByIdDTOFromDBModel(dbTeam db.GetTeamWithStudentNamesByTeamIDRow) (promptTypes.Team, error) {
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

func GetTeamWithFullNameDTOsFromDBModels(dbTeams []db.GetTeamsWithStudentNamesRow) ([]promptTypes.Team, error) {
	teams := make([]promptTypes.Team, 0, len(dbTeams))
	for _, dbTeam := range dbTeams {
		t, err := GetTeamDTOFromDBModel(dbTeam)
		if err != nil {
			// handle or log error; skip or abort as you preferable
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}
