package allocationDTO

import "github.com/google/uuid"

type UpdateAllocationRequest struct {
	TeamID uuid.UUID `json:"teamID" binding:"required"`
}
