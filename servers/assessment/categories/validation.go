package categories

import (
	"errors"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/categories/categoryDTO"
)

var (
	ErrIncompleteSchemaOrder   = errors.New("the order must list every category and competency of the schema exactly once")
	ErrDuplicateCompetencyName = errors.New("a category cannot contain two competencies with the same name")
)

// validateSchemaOrder checks that the order lists exactly the categories and competencies of the schema
// and that no category ends up with two competencies of the same name.
func validateSchemaOrder(current []categoryDTO.CategoryWithCompetencies, order []categoryDTO.CategoryOrder) error {
	categoryIDs := make(map[uuid.UUID]struct{}, len(current))
	competencyNames := make(map[uuid.UUID]string)
	for _, category := range current {
		categoryIDs[category.ID] = struct{}{}
		for _, competency := range category.Competencies {
			competencyNames[competency.ID] = competency.Name
		}
	}

	if len(order) != len(categoryIDs) {
		return ErrIncompleteSchemaOrder
	}

	seenCategories := make(map[uuid.UUID]struct{}, len(order))
	seenCompetencies := make(map[uuid.UUID]struct{}, len(competencyNames))
	for _, category := range order {
		if _, ok := categoryIDs[category.ID]; !ok {
			return ErrIncompleteSchemaOrder
		}
		if _, duplicate := seenCategories[category.ID]; duplicate {
			return ErrIncompleteSchemaOrder
		}
		seenCategories[category.ID] = struct{}{}

		namesInCategory := make(map[string]struct{}, len(category.CompetencyIDs))
		for _, competencyID := range category.CompetencyIDs {
			name, ok := competencyNames[competencyID]
			if !ok {
				return ErrIncompleteSchemaOrder
			}
			if _, duplicate := seenCompetencies[competencyID]; duplicate {
				return ErrIncompleteSchemaOrder
			}
			seenCompetencies[competencyID] = struct{}{}

			if _, duplicate := namesInCategory[name]; duplicate {
				return ErrDuplicateCompetencyName
			}
			namesInCategory[name] = struct{}{}
		}
	}

	if len(seenCompetencies) != len(competencyNames) {
		return ErrIncompleteSchemaOrder
	}

	return nil
}
