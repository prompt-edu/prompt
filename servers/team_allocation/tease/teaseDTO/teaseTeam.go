package teaseDTO

import (
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

type TeaseTeam struct {
	TeamID              string               `json:"id"`
	TeamName            string               `json:"name"`
	ImportedConstraints []ImportedConstraint `json:"importedConstraints,omitempty"`
}

func GetTeaseTeamResponseFromDBModel(teams []db.Team) []TeaseTeam {
	teaseTeams := make([]TeaseTeam, 0, len(teams))
	for _, team := range teams {
		teaseTeam := TeaseTeam{
			TeamID:   team.ID.String(),
			TeamName: team.Name,
		}
		if team.TeamSizeMin.Valid || team.TeamSizeMax.Valid {
			lowerBound := int32(0)
			upperBound := int32(999)
			if team.TeamSizeMin.Valid {
				lowerBound = team.TeamSizeMin.Int32
			}
			if team.TeamSizeMax.Valid {
				upperBound = team.TeamSizeMax.Int32
			}
			teaseTeam.ImportedConstraints = []ImportedConstraint{
				{
					Type:       "team_size",
					LowerBound: lowerBound,
					UpperBound: upperBound,
				},
			}
		}
		teaseTeams = append(teaseTeams, teaseTeam)
	}
	return teaseTeams
}
