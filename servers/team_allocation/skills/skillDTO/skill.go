package skillDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

// Skill represents a simplified view of the skill record.
type Skill struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// GetSkillDTOsFromDBModels converts a slice of database skill models into a slice of Skill DTOs.
func GetSkillDTOsFromDBModels(dbSkills []db.Skill) []Skill {
	skills := make([]Skill, len(dbSkills))
	for i, s := range dbSkills {
		skills[i] = Skill{
			ID:   s.ID,
			Name: s.Name,
		}
	}
	return skills
}
