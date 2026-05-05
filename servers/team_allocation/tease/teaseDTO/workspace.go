package teaseDTO

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TeaseWorkspace is the persisted TEASE workspace for one course phase.
type TeaseWorkspace struct {
	CoursePhaseID    uuid.UUID       `json:"coursePhaseId"`
	Constraints      json.RawMessage `json:"constraints" swaggertype:"object"`
	LockedStudents   json.RawMessage `json:"lockedStudents" swaggertype:"object"`
	AllocationsDraft json.RawMessage `json:"allocationsDraft" swaggertype:"object"`
	AlgorithmType    *string         `json:"algorithmType"`
	LastSavedAt      *time.Time      `json:"lastSavedAt"`
	LastExportedAt   *time.Time      `json:"lastExportedAt"`
	UpdatedBy        *uuid.UUID      `json:"updatedBy"`
}

// TeaseWorkspaceRequest is the client-editable workspace snapshot.
type TeaseWorkspaceRequest struct {
	Constraints      json.RawMessage `json:"constraints" swaggertype:"object"`
	LockedStudents   json.RawMessage `json:"lockedStudents" swaggertype:"object"`
	AllocationsDraft json.RawMessage `json:"allocationsDraft" swaggertype:"object"`
	AlgorithmType    *string         `json:"algorithmType"`
}

// TeaseSaveRequest publishes a workspace and allocation set.
type TeaseSaveRequest struct {
	TeaseWorkspaceRequest
	Allocations []Allocation `json:"allocations"`
}
