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

-- name: CountLiveInstancesForConfig :one
-- Counts the instances of one config that still describe a live external resource.
-- A failed instance never created anything, so it does not pin the config's identity.
SELECT COUNT(*)
FROM resource_instance
WHERE resource_config_id = $1 AND status != 'failed';

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

-- name: FailStaleInProgressInstances :execrows
-- Recovers instances a crashed process left claimed. They are marked failed rather
-- than pending: nothing picks a pending row up on its own (the worker needs the
-- lecturer's auth header to resolve targets), while a failed one is terminal, visible
-- in the UI and retryable. The cutoff keeps the sweep off work another replica is
-- still doing - a worker's context is capped well below it.
-- The cutoff is computed from the database clock, not the caller's: updated_at is a
-- timestamp without a time zone, so a client-side cutoff would be off by the server's
-- UTC offset.
UPDATE resource_instance
SET status = 'failed',
    error_message = sqlc.arg(error_message),
    updated_at = NOW()
WHERE status = 'in_progress'
  AND updated_at < CURRENT_TIMESTAMP - make_interval(secs => sqlc.arg(max_claim_age_seconds)::double precision);

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

-- name: GetResourceInstancesByCourseParticipationIDs :many
-- Everything the service stores about one subject: the resources provisioned for them
-- personally. The config is joined in so the export names what was created rather than
-- an opaque config id. Team-scoped instances are course data, not subject data.
SELECT instance.id,
       instance.course_phase_id,
       instance.status,
       instance.external_id,
       instance.external_url,
       instance.error_message,
       instance.created_at,
       instance.updated_at,
       config.provider_type,
       config.resource_type,
       config.name_template
FROM resource_instance AS instance
    JOIN resource_config AS config ON config.id = instance.resource_config_id
WHERE instance.course_participation_id = ANY(sqlc.arg(course_participation_ids)::uuid[])
ORDER BY instance.created_at;

-- name: DeleteResourceInstancesByCourseParticipationIDs :exec
DELETE FROM resource_instance
WHERE course_participation_id = ANY(sqlc.arg(course_participation_ids)::uuid[]);
