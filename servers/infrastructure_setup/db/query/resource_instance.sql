-- name: CreateResourceInstance :one
INSERT INTO resource_instance (
    id,
    resource_config_id,
    course_phase_id,
    team_id,
    course_participation_id
)
VALUES (gen_random_uuid(), $1, $2, $3, $4)
ON CONFLICT DO NOTHING
RETURNING id, resource_config_id, course_phase_id, team_id, course_participation_id, status, external_id, external_url, error_message, created_at, updated_at;

-- name: GetResourceInstance :one
SELECT id, resource_config_id, course_phase_id, team_id, course_participation_id, status, external_id, external_url, error_message, created_at, updated_at
FROM resource_instance
WHERE id = $1 AND course_phase_id = $2;

-- name: ListResourceInstances :many
SELECT id, resource_config_id, course_phase_id, team_id, course_participation_id, status, external_id, external_url, error_message, created_at, updated_at
FROM resource_instance
WHERE course_phase_id = $1
ORDER BY created_at DESC;

-- name: UpdateResourceInstanceStatus :exec
UPDATE resource_instance
SET status = $2,
    external_id = $3,
    external_url = $4,
    error_message = $5,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteResourceInstance :exec
DELETE FROM resource_instance
WHERE id = $1 AND course_phase_id = $2;

-- name: CountNonTerminalInstances :one
SELECT COUNT(*)
FROM resource_instance
WHERE course_phase_id = $1 AND status IN ('pending', 'in_progress');

-- name: ListPendingInstances :many
SELECT id, resource_config_id, course_phase_id, team_id, course_participation_id, status, external_id, external_url, error_message, created_at, updated_at
FROM resource_instance
WHERE course_phase_id = $1 AND status = 'pending';

-- name: ResetInProgressToPending :exec
UPDATE resource_instance
SET status = 'pending',
    updated_at = NOW()
WHERE status = 'in_progress';

-- name: ResetFailedInstanceToPending :exec
UPDATE resource_instance
SET status = 'pending',
    error_message = NULL,
    updated_at = NOW()
WHERE id = $1 AND course_phase_id = $2 AND status = 'failed';
