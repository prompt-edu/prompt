package teaseDTO

import (
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

type TeaseTeam struct {
	TeamID   string `json:"id"`
	TeamName string `json:"name"`
}

func GetTeaseTeamResponseFromDBModel(teams []db.Team) []TeaseTeam {
	teaseTeams := make([]TeaseTeam, 0, len(teams))
	for _, team := range teams {
		teaseTeam := TeaseTeam{
			TeamID:   team.ID.String(),
			TeamName: team.Name,
		}
		teaseTeams = append(teaseTeams, teaseTeam)
	}
	return teaseTeams
}
