package categories

import (
	"testing"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/categories/categoryDTO"
	"github.com/prompt-edu/prompt/servers/assessment/competencies/competencyDTO"
	"github.com/stretchr/testify/assert"
)

func TestValidateSchemaOrder(t *testing.T) {
	categoryA, categoryB := uuid.New(), uuid.New()
	competencyA1, competencyA2, competencyB1 := uuid.New(), uuid.New(), uuid.New()
	current := []categoryDTO.CategoryWithCompetencies{
		{ID: categoryA, Competencies: []competencyDTO.Competency{{ID: competencyA1, Name: "Shared"}, {ID: competencyA2, Name: "Only A"}}},
		{ID: categoryB, Competencies: []competencyDTO.Competency{{ID: competencyB1, Name: "Shared"}}},
	}

	tests := []struct {
		name  string
		order []categoryDTO.CategoryOrder
		want  error
	}{
		{
			name: "reordered and moved",
			order: []categoryDTO.CategoryOrder{
				{ID: categoryB, CompetencyIDs: []uuid.UUID{competencyB1, competencyA2}},
				{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyA1}},
			},
		},
		{
			name:  "missing category",
			order: []categoryDTO.CategoryOrder{{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyA1, competencyA2, competencyB1}}},
			want:  ErrIncompleteSchemaOrder,
		},
		{
			name: "unknown category",
			order: []categoryDTO.CategoryOrder{
				{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyA1, competencyA2}},
				{ID: uuid.New(), CompetencyIDs: []uuid.UUID{competencyB1}},
			},
			want: ErrIncompleteSchemaOrder,
		},
		{
			name: "duplicate category",
			order: []categoryDTO.CategoryOrder{
				{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyA1, competencyA2}},
				{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyB1}},
			},
			want: ErrIncompleteSchemaOrder,
		},
		{
			name: "missing competency",
			order: []categoryDTO.CategoryOrder{
				{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyA1}},
				{ID: categoryB, CompetencyIDs: []uuid.UUID{competencyB1}},
			},
			want: ErrIncompleteSchemaOrder,
		},
		{
			name: "competency listed twice",
			order: []categoryDTO.CategoryOrder{
				{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyA1, competencyA2}},
				{ID: categoryB, CompetencyIDs: []uuid.UUID{competencyB1, competencyA2}},
			},
			want: ErrIncompleteSchemaOrder,
		},
		{
			name: "unknown competency",
			order: []categoryDTO.CategoryOrder{
				{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyA1, competencyA2, uuid.New()}},
				{ID: categoryB, CompetencyIDs: []uuid.UUID{competencyB1}},
			},
			want: ErrIncompleteSchemaOrder,
		},
		{
			name: "duplicate name after move",
			order: []categoryDTO.CategoryOrder{
				{ID: categoryA, CompetencyIDs: []uuid.UUID{competencyA2}},
				{ID: categoryB, CompetencyIDs: []uuid.UUID{competencyB1, competencyA1}},
			},
			want: ErrDuplicateCompetencyName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSchemaOrder(current, tt.order)
			if tt.want == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.want)
			}
		})
	}
}
