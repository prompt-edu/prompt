package surveyDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

type StudentTeamPreferenceResponse struct {
	TeamID     uuid.UUID `json:"teamID"`
	Preference int32     `json:"preference"`
}

func GetStudentTeamPreferenceRepsonseDTOsFromDBModel(preferences []db.GetStudentTeamPreferencesRow) []StudentTeamPreferenceResponse {
	preferenceDTOs := make([]StudentTeamPreferenceResponse, 0, len(preferences))
	for _, preference := range preferences {
		preferenceDTOs = append(preferenceDTOs, StudentTeamPreferenceResponse{
			TeamID:     preference.TeamID,
			Preference: preference.Preference,
		})
	}
	return preferenceDTOs
}

func GetStudentPreferenceResponseDTOsFromTypedDBModel(preferences []db.GetStudentPreferencesByTeamTypeRow) []StudentTeamPreferenceResponse {
	preferenceDTOs := make([]StudentTeamPreferenceResponse, 0, len(preferences))
	for _, preference := range preferences {
		preferenceDTOs = append(preferenceDTOs, StudentTeamPreferenceResponse{
			TeamID:     preference.TeamID,
			Preference: preference.Preference,
		})
	}
	return preferenceDTOs
}
