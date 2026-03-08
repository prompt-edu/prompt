package allocationDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
)

type AllocationWithParticipation struct {
	CourseParticipationID uuid.UUID `json:"courseParticipationID"`
	TeamAllocation        uuid.UUID `json:"teamAllocation"`
}

// GetAllocationsFromDBModels converts a slice of db.Allocation to DTOs.
func GetAllocationsFromDBModels(dbAllocations []db.Allocation) []AllocationWithParticipation {
	allocations := make([]AllocationWithParticipation, 0, len(dbAllocations))
	for _, allocation := range dbAllocations {
		allocations = append(allocations, AllocationWithParticipation{
			CourseParticipationID: allocation.CourseParticipationID,
			TeamAllocation:        allocation.TeamID,
		})
	}
	return allocations
}
