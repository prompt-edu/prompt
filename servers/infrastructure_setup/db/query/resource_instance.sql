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

-- name: MarkInstanceCreated :exec
UPDATE resource_instance
SET status = 'created',
    external_id = $2,
    external_url = $3,
    error_message = NULL,
    updated_at = NOW()
WHERE id = $1;

-- name: MarkInstancePartial :exec
UPDATE resource_instance
SET status = 'partial',
    external_id = $2,
    external_url = $3,
    error_message = $4,
    updated_at = NOW()
WHERE id = $1;

-- name: MarkInstanceFailed :exec
UPDATE resource_instance
SET status = 'failed',
    error_message = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteResourceInstance :exec
DELETE FROM resource_instance
WHERE id = $1 AND course_phase_id = $2;

-- name: CountNonTerminalInstances :one
SELECT COUNT(*)
FROM resource_instance
WHERE course_phase_id = $1 AND status IN ('pending', 'in_progress');

-- name: ClaimPendingInstances :many
-- Atomically takes ownership of every pending instance in one statement, so two
-- overlapping runs can never process the same row.
UPDATE resource_instance
SET status = 'in_progress',
    updated_at = NOW()
WHERE id IN (
    SELECT pending.id FROM resource_instance AS pending
    WHERE pending.course_phase_id = $1 AND pending.status = 'pending'
    FOR UPDATE SKIP LOCKED
)
RETURNING id, resource_config_id, course_phase_id, team_id, course_participation_id, status, external_id, external_url, error_message, created_at, updated_at;

-- name: ResetInProgressToPending :exec
UPDATE resource_instance
SET status = 'pending',
    updated_at = NOW()
WHERE status = 'in_progress';

-- name: ResetInstanceToPending :one
UPDATE resource_instance
SET status = 'pending',
    error_message = NULL,
    updated_at = NOW()
WHERE id = $1 AND course_phase_id = $2 AND status IN ('failed', 'partial')
RETURNING id, resource_config_id, course_phase_id, team_id, course_participation_id, status, external_id, external_url, error_message, created_at, updated_at;

-- name: TryLockPhaseExecution :one
-- Transaction-scoped advisory lock keyed on the course phase. Held only while the
-- trigger counts and inserts instances, never across provider calls.
SELECT pg_try_advisory_xact_lock(hashtextextended(sqlc.arg(course_phase_id)::text, 0));
