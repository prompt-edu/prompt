package surveyDTO

import db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"

type StudentSurveyResponse struct {
	TeamPreferences  []StudentTeamPreferenceResponse `json:"teamPreferences,omitempty"`
	FieldPreferences []StudentTeamPreferenceResponse `json:"fieldPreferences,omitempty"`
	SkillResponses   []StudentSkillResponse          `json:"skillResponses"`
	PreferenceMode   string                          `json:"preferenceMode,omitempty"`
}

func GetStudentSurveyResponseDTOFromDBModels(teamPreferences []db.GetStudentTeamPreferencesRow, skillResponses []db.GetStudentSkillResponsesRow) StudentSurveyResponse {
	teamPreferencesDTO := GetStudentTeamPreferenceRepsonseDTOsFromDBModel(teamPreferences)
	skillResponsesDTO := GetStudentSkillResponsesDTOFromDBModel(skillResponses)

	return StudentSurveyResponse{
		TeamPreferences: teamPreferencesDTO,
		SkillResponses:  skillResponsesDTO,
	}
}

func GetStudentSurveyResponseDTOFromTypedDBModels(preferences []db.GetStudentPreferencesByTeamTypeRow, skillResponses []db.GetStudentSkillResponsesByPreferenceModeRow, preferenceMode string) StudentSurveyResponse {
	preferencesDTO := GetStudentPreferenceResponseDTOsFromTypedDBModel(preferences)
	skillResponsesDTO := GetStudentSkillResponsesDTOFromPreferenceModeDBModel(skillResponses)

	response := StudentSurveyResponse{
		SkillResponses: skillResponsesDTO,
		PreferenceMode: preferenceMode,
	}
	if preferenceMode == "fields" {
		response.FieldPreferences = preferencesDTO
	} else {
		response.TeamPreferences = preferencesDTO
	}
	return response
}
