package assessmentSchemaDTO

import (
	"time"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type AssessmentSchema struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	// IsOwnedByCurrentPhase is only populated by phase-scoped listings; nil elsewhere.
	IsOwnedByCurrentPhase *bool `json:"isOwnedByCurrentPhase,omitempty"`
}

func MapDBAssessmentSchemaToDTOAssessmentSchema(dbSchema db.AssessmentSchema) AssessmentSchema {
	return AssessmentSchema{
		ID:          dbSchema.ID,
		Name:        dbSchema.Name,
		Description: dbSchema.Description.String,
		CreatedAt:   dbSchema.CreatedAt.Time,
		UpdatedAt:   dbSchema.UpdatedAt.Time,
	}
}
