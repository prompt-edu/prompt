package allocationDTO

import "github.com/google/uuid"

type Allocation struct {
	TeamAllocation uuid.UUID `json:"teamAllocation"`
}
