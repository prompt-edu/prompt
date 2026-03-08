package categoryDTO

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/competencies/competencyDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type CategoryWithCompetencies struct {
	ID           uuid.UUID                                        `json:"id"`
	Name         string                                           `json:"name"`
	ShortName    string                                           `json:"shortName"`
	Description  string                                           `json:"description"`
	Weight       int32                                            `json:"weight"`
	Competencies []competencyDTO.CompetencyWithMappedCompetencies `json:"competencies"`
}

func MapToCategoryWithCompetenciesDTO(rows []db.GetCategoriesWithCompetenciesRow) []CategoryWithCompetencies {
	var result = make([]CategoryWithCompetencies, 0, len(rows))

	for _, row := range rows {
		var competencies []competencyDTO.CompetencyWithMappedCompetencies
		if row.Competencies != nil {
			if err := json.Unmarshal(row.Competencies, &competencies); err != nil {
				log.Printf("Error unmarshalling competencies for category %s: %v", row.ID, err)
			}
		}

		category := CategoryWithCompetencies{
			ID:           row.ID,
			Name:         row.Name,
			ShortName:    row.ShortName.String,
			Description:  row.Description.String,
			Weight:       row.Weight,
			Competencies: competencies,
		}
		result = append(result, category)
	}

	return result
}
