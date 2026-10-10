package categoryDTO

import "github.com/google/uuid"

// CategoryOrder lists the competencies of one category in their display order.
type CategoryOrder struct {
	ID            uuid.UUID   `json:"id"`
	CompetencyIDs []uuid.UUID `json:"competencyIDs"`
}

// UpdateSchemaOrderRequest describes the full layout of a schema: the categories in their display order,
// each with its competencies in display order. A competency listed under another category moves there.
// The schema is the one the categories belong to.
type UpdateSchemaOrderRequest struct {
	Categories []CategoryOrder `json:"categories"`
}

// UpdateSchemaOrderResponse names the schema the order was written to. It differs from the schema the
// request's categories belong to when the write copied a shared schema for the course phase.
type UpdateSchemaOrderResponse struct {
	AssessmentSchemaID uuid.UUID `json:"assessmentSchemaID"`
}
