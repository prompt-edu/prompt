package teaseDTO

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TeaseWorkspace is the over-the-wire representation of a persisted
// Tease workspace. Fields mirror §4.2 of the Phase 1 plan.
//
// `Constraints`, `LockedStudents`, and `AllocationsDraft` are stored
// verbatim as JSON snapshots in the `tease_workspace` table; we never
// inspect their contents on the server side, so they are passed
// through as `json.RawMessage`.
type TeaseWorkspace struct {
	CoursePhaseID    uuid.UUID       `json:"coursePhaseId"`
	Constraints      json.RawMessage `json:"constraints"`
	LockedStudents   json.RawMessage `json:"lockedStudents"`
	AllocationsDraft json.RawMessage `json:"allocationsDraft"`
	AlgorithmType    *string         `json:"algorithmType"`
	LastSavedAt      *time.Time      `json:"lastSavedAt"`
	LastExportedAt   *time.Time      `json:"lastExportedAt"`
	UpdatedBy        *uuid.UUID      `json:"updatedBy"`
}

// TeaseWorkspaceRequest is the payload for PUT /workspace and the
// workspace portion of POST /save. Server-managed fields (timestamps,
// UpdatedBy) are intentionally omitted — UpdatedBy is derived from the
// authenticated token in the handler.
type TeaseWorkspaceRequest struct {
	Constraints      json.RawMessage `json:"constraints"`
	LockedStudents   json.RawMessage `json:"lockedStudents"`
	AllocationsDraft json.RawMessage `json:"allocationsDraft"`
	AlgorithmType    *string         `json:"algorithmType"`
}

// TeaseSaveRequest is the payload for POST /save. The client supplies
// the workspace snapshot plus the finalised allocations to write to
// the `allocations` table in the same transaction.
type TeaseSaveRequest struct {
	Constraints      json.RawMessage `json:"constraints"`
	LockedStudents   json.RawMessage `json:"lockedStudents"`
	AllocationsDraft json.RawMessage `json:"allocationsDraft"`
	AlgorithmType    *string         `json:"algorithmType"`
	Allocations      []Allocation    `json:"allocations"`
}
