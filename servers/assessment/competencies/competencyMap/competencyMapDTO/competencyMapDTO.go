package competencyMapDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type CompetencyMapping struct {
	FromCompetencyID uuid.UUID `json:"fromCompetencyId"`
	ToCompetencyID   uuid.UUID `json:"toCompetencyId"`
}

func GetCompetencyMappingFromDBModel(dbModel db.CompetencyMap) CompetencyMapping {
	return CompetencyMapping{
		FromCompetencyID: dbModel.FromCompetencyID,
		ToCompetencyID:   dbModel.ToCompetencyID,
	}
}

func GetCompetencyMappingsFromDBModels(dbModels []db.CompetencyMap) []CompetencyMapping {
	var mappings []CompetencyMapping
	for _, dbModel := range dbModels {
		mappings = append(mappings, GetCompetencyMappingFromDBModel(dbModel))
	}
	return mappings
}
