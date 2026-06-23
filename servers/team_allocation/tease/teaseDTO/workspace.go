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

// TeaseWorkspaceSettings holds the non-allocation workspace fields shared by the
// draft-save and publish requests.
type TeaseWorkspaceSettings struct {
	Constraints    json.RawMessage `json:"constraints" swaggertype:"object"`
	LockedStudents json.RawMessage `json:"lockedStudents" swaggertype:"object"`
	AlgorithmType  *string         `json:"algorithmType"`
}

// TeaseWorkspaceRequest is the client-editable workspace snapshot.
type TeaseWorkspaceRequest struct {
	TeaseWorkspaceSettings
	AllocationsDraft json.RawMessage `json:"allocationsDraft" swaggertype:"object"`
}

// TeaseSaveRequest publishes a workspace and allocation set. The persisted draft
// is derived from Allocations, so it is not sent separately.
type TeaseSaveRequest struct {
	TeaseWorkspaceSettings
	Allocations []Allocation `json:"allocations"`
}
