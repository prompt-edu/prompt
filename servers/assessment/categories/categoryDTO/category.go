package categoryDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type Category struct {
	ID                   uuid.UUID `json:"id"`
	Name                 string    `json:"name"`
	ShortName            string    `json:"shortName"`
	Description          string    `json:"description"`
	Weight               int32     `json:"weight"`
	AssessmentSchemaID uuid.UUID `json:"assessmentSchemaID"`
}

func GetCategoryDTOsFromDBModels(dbCategories []db.Category) []Category {
	categories := make([]Category, 0, len(dbCategories))
	for _, c := range dbCategories {
		categories = append(categories, Category{
			ID:                   c.ID,
			Name:                 c.Name,
			ShortName:            c.ShortName.String,
			Description:          c.Description.String,
			Weight:               c.Weight,
			AssessmentSchemaID: c.AssessmentSchemaID,
		})
	}
	return categories
}
