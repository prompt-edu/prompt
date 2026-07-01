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

// Maps the 5-level survey scale to TEASE's 4-level proficiency scale.
// The two lowest levels both collapse to Novice.
func getTeaseSkillLevel(skillLevel db.SkillLevel) string {
	switch skillLevel {
	case db.SkillLevelVeryBad:
		return "Novice"
	case db.SkillLevelBad:
		return "Novice"
	case db.SkillLevelOk:
		return "Intermediate"
	case db.SkillLevelGood:
		return "Advanced"
	case db.SkillLevelVeryGood:
		return "Expert"
	default:
		return "Unknown"
	}
}
