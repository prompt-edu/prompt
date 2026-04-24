package teaseDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

type ProjectPreference struct {
	ProjectID uuid.UUID `json:"projectId"`
	Priority  int32     `json:"priority"`
}

func GetProjectPreferenceFromDBModel(teamPreferences []db.GetStudentTeamPreferencesRow) []ProjectPreference {
	projectPreferences := make([]ProjectPreference, 0, len(teamPreferences))
	for _, teamPreference := range teamPreferences {
		projectPreferences = append(projectPreferences, ProjectPreference{
			ProjectID: teamPreference.TeamID,
			Priority:  teamPreference.Preference - 1, // in tease 0 is the highest priority
		})
	}
	return projectPreferences
}

func GetProjectPreferenceFromTypedDBModel(teamPreferences []db.GetStudentPreferencesByTeamTypeRow) []ProjectPreference {
	projectPreferences := make([]ProjectPreference, 0, len(teamPreferences))
	for _, teamPreference := range teamPreferences {
		projectPreferences = append(projectPreferences, ProjectPreference{
			ProjectID: teamPreference.TeamID,
			Priority:  teamPreference.Preference - 1, // in tease 0 is the highest priority
		})
	}
	return projectPreferences
}
