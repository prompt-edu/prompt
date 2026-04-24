package surveyDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

type StudentSkillResponse struct {
	SkillID    uuid.UUID     `json:"skillID"`
	SkillLevel db.SkillLevel `json:"skillLevel"`
}

func GetStudentSkillResponsesDTOFromDBModel(skills []db.GetStudentSkillResponsesRow) []StudentSkillResponse {
	skillDTOs := make([]StudentSkillResponse, 0, len(skills))
	for _, skill := range skills {
		skillDTOs = append(skillDTOs, StudentSkillResponse{
			SkillID:    skill.SkillID,
			SkillLevel: skill.SkillLevel,
		})
	}
	return skillDTOs
}

func GetStudentSkillResponsesDTOFromPreferenceModeDBModel(skills []db.GetStudentSkillResponsesByPreferenceModeRow) []StudentSkillResponse {
	skillDTOs := make([]StudentSkillResponse, 0, len(skills))
	for _, skill := range skills {
		skillDTOs = append(skillDTOs, StudentSkillResponse{
			SkillID:    skill.SkillID,
			SkillLevel: skill.SkillLevel,
		})
	}
	return skillDTOs
}
