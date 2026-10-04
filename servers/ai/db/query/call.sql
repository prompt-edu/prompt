-- name: CreateCall :one
INSERT INTO ai_call (course_phase_id, actor_id, actor_role, feature, template, template_version,
                     requested_model, provider, params, context_hash, outcome, http_status, error_code,
                     streamed, server_version, completed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING id;

-- name: CreateCallContent :exec
INSERT INTO ai_call_content (call_id, request)
VALUES ($1, $2);

-- name: CreateCallSubjects :exec
INSERT INTO ai_call_subject (call_id, course_participation_id)
SELECT $1, unnest(@course_participation_ids::uuid[]);

-- name: CompleteCall :execrows
UPDATE ai_call
SET served_model      = $2,
    response_hash     = $3,
    outcome           = $4,
    http_status       = $5,
    finish_reason     = $6,
    error_code        = $7,
    prompt_tokens     = $8,
    completion_tokens = $9,
    first_token_at    = $10,
    completed_at      = now()
WHERE id = $1
  AND completed_at IS NULL;

-- name: SetCallResponse :exec
UPDATE ai_call_content
SET response = $2
WHERE call_id = $1
  AND response IS NULL;

-- name: ListCalls :many
SELECT *
FROM ai_call
WHERE course_phase_id = @course_phase_id
  AND (sqlc.narg(cursor_requested_at)::timestamptz IS NULL
    OR (requested_at, id) < (sqlc.narg(cursor_requested_at)::timestamptz, sqlc.narg(cursor_id)::uuid))
ORDER BY requested_at DESC, id DESC
LIMIT @page_size;

-- name: GetCall :one
SELECT sqlc.embed(ai_call),
       content.request,
       content.response,
       (restriction.call_id IS NOT NULL)::boolean AS restricted
FROM ai_call
         LEFT JOIN ai_call_content content ON content.call_id = ai_call.id
         LEFT JOIN ai_call_content_restriction restriction ON restriction.call_id = ai_call.id
WHERE ai_call.id = $1
  AND ai_call.course_phase_id = $2;

-- name: ListCallSubjects :many
SELECT course_participation_id
FROM ai_call_subject
WHERE call_id = $1
ORDER BY course_participation_id;

-- name: ListCallEvents :many
SELECT *
FROM ai_call_event
WHERE call_id = $1
ORDER BY created_at, id;

-- name: CreateCallEvent :one
INSERT INTO ai_call_event (call_id, actor_id, type, data)
SELECT ai_call.id, @actor_id, @type, @data
FROM ai_call
WHERE ai_call.id = @call_id
  AND ai_call.course_phase_id = @course_phase_id
  AND ai_call.outcome <> 'denied'
  AND (@by_admin::boolean OR ai_call.actor_id = @actor_id)
RETURNING *;
