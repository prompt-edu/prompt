-- name: GetTeaseWorkspace :one
SELECT course_phase_id,
       constraints,
       locked_students,
       allocations_draft,
       algorithm_type,
       updated_by,
       last_saved_at,
       last_exported_at
FROM tease_workspace
WHERE course_phase_id = $1;

-- name: UpsertTeaseWorkspace :one
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
          last_exported_at;

-- name: StampTeaseWorkspaceExportedAt :exec
UPDATE tease_workspace
SET last_exported_at = CURRENT_TIMESTAMP
WHERE course_phase_id = $1;
