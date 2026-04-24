package teaseDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

type StudentSkillResponse struct {
	ID          uuid.UUID `json:"id"`
	Proficiency string    `json:"proficiency"`
}

func GetTeaseStudentSkillResponseFromDBModel(skillResponses []db.GetStudentSkillResponsesRow) []StudentSkillResponse {
	skills := make([]StudentSkillResponse, 0, len(skillResponses))
	for _, skillResponse := range skillResponses {
		skills = append(skills, StudentSkillResponse{
			ID:          skillResponse.SkillID,
			Proficiency: getTeaseSkillLevel(skillResponse.SkillLevel),
		})
	}
	return skills
}

func GetTeaseStudentSkillResponseFromPreferenceModeDBModel(skillResponses []db.GetStudentSkillResponsesByPreferenceModeRow) []StudentSkillResponse {
	skills := make([]StudentSkillResponse, 0, len(skillResponses))
	for _, skillResponse := range skillResponses {
		skills = append(skills, StudentSkillResponse{
			ID:          skillResponse.SkillID,
			Proficiency: getTeaseSkillLevel(skillResponse.SkillLevel),
		})
	}
	return skills
}

// Tease defines the Skill Levels upper case, while the DB (and Assessment) defines them lower case
func getTeaseSkillLevel(skillLevel db.SkillLevel) string {
	switch skillLevel {
	case db.SkillLevelNovice:
		return "Novice"
	case db.SkillLevelIntermediate:
		return "Intermediate"
	case db.SkillLevelAdvanced:
		return "Advanced"
	case db.SkillLevelExpert:
		return "Expert"
	default:
		return "Unknown"
	}
}
