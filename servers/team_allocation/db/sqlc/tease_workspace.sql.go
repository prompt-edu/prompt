// Hand-written in the style of sqlc-generated code.
// Regenerate-compatible: if `sqlc generate` is run with the matching
// query file at db/query/tease_workspace.sql, the output should match
// (package layout, parameter struct names, pgtype usage, json tags).

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type TeaseWorkspace struct {
	CoursePhaseID    uuid.UUID        `json:"course_phase_id"`
	Constraints      []byte           `json:"constraints"`
	LockedStudents   []byte           `json:"locked_students"`
	AllocationsDraft []byte           `json:"allocations_draft"`
	AlgorithmType    pgtype.Text      `json:"algorithm_type"`
	UpdatedBy        pgtype.UUID      `json:"updated_by"`
	LastSavedAt      pgtype.Timestamp `json:"last_saved_at"`
	LastExportedAt   pgtype.Timestamp `json:"last_exported_at"`
}

const getTeaseWorkspace = `-- name: GetTeaseWorkspace :one
SELECT course_phase_id,
       constraints,
       locked_students,
       allocations_draft,
       algorithm_type,
       updated_by,
       last_saved_at,
       last_exported_at
FROM tease_workspace
WHERE course_phase_id = $1
`

func (q *Queries) GetTeaseWorkspace(ctx context.Context, coursePhaseID uuid.UUID) (TeaseWorkspace, error) {
	row := q.db.QueryRow(ctx, getTeaseWorkspace, coursePhaseID)
	var i TeaseWorkspace
	err := row.Scan(
		&i.CoursePhaseID,
		&i.Constraints,
		&i.LockedStudents,
		&i.AllocationsDraft,
		&i.AlgorithmType,
		&i.UpdatedBy,
		&i.LastSavedAt,
		&i.LastExportedAt,
	)
	return i, err
}

const upsertTeaseWorkspace = `-- name: UpsertTeaseWorkspace :one
INSERT INTO tease_workspace (
    course_phase_id,
    constraints,
    locked_students,
    allocations_draft,
    algorithm_type,
    updated_by,
    last_saved_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    CURRENT_TIMESTAMP
)
ON CONFLICT (course_phase_id)
DO UPDATE
SET constraints       = EXCLUDED.constraints,
    locked_students   = EXCLUDED.locked_students,
    allocations_draft = EXCLUDED.allocations_draft,
    algorithm_type    = EXCLUDED.algorithm_type,
    updated_by        = EXCLUDED.updated_by,
    last_saved_at     = CURRENT_TIMESTAMP
RETURNING course_phase_id,
          constraints,
          locked_students,
          allocations_draft,
          algorithm_type,
          updated_by,
          last_saved_at,
          last_exported_at
`

type UpsertTeaseWorkspaceParams struct {
	CoursePhaseID    uuid.UUID   `json:"course_phase_id"`
	Constraints      []byte      `json:"constraints"`
	LockedStudents   []byte      `json:"locked_students"`
	AllocationsDraft []byte      `json:"allocations_draft"`
	AlgorithmType    pgtype.Text `json:"algorithm_type"`
	UpdatedBy        pgtype.UUID `json:"updated_by"`
}

func (q *Queries) UpsertTeaseWorkspace(ctx context.Context, arg UpsertTeaseWorkspaceParams) (TeaseWorkspace, error) {
	row := q.db.QueryRow(ctx, upsertTeaseWorkspace,
		arg.CoursePhaseID,
		arg.Constraints,
		arg.LockedStudents,
		arg.AllocationsDraft,
		arg.AlgorithmType,
		arg.UpdatedBy,
	)
	var i TeaseWorkspace
	err := row.Scan(
		&i.CoursePhaseID,
		&i.Constraints,
		&i.LockedStudents,
		&i.AllocationsDraft,
		&i.AlgorithmType,
		&i.UpdatedBy,
		&i.LastSavedAt,
		&i.LastExportedAt,
	)
	return i, err
}

const stampTeaseWorkspaceExportedAt = `-- name: StampTeaseWorkspaceExportedAt :exec
UPDATE tease_workspace
SET last_exported_at = CURRENT_TIMESTAMP
WHERE course_phase_id = $1
`

func (q *Queries) StampTeaseWorkspaceExportedAt(ctx context.Context, coursePhaseID uuid.UUID) error {
	_, err := q.db.Exec(ctx, stampTeaseWorkspaceExportedAt, coursePhaseID)
	return err
}
