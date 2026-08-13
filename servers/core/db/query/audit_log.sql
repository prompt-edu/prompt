-- name: CreateAuditEntry :one
INSERT INTO audit_log (
  actor_id, actor_name, actor_email, actor_roles, actor_role,
  action, action_key, outcome,
  entity_type, entity_id, entity_name,
  course_id, course_phase_id, source_service,
  http_method, http_path, http_status, metadata
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
)
RETURNING *;

-- name: DeleteExpiredAuditEntries :exec
DELETE FROM audit_log WHERE created_at < $1;

-- name: ListAuditLog :many
SELECT * FROM audit_log
WHERE (sqlc.narg('course_id')::uuid IS NULL OR course_id = sqlc.narg('course_id'))
  AND (sqlc.narg('actor_role')::text IS NULL OR actor_role = sqlc.narg('actor_role'))
  AND (sqlc.narg('outcome')::text IS NULL OR outcome = sqlc.narg('outcome'))
  AND (sqlc.narg('action_key')::text IS NULL OR action_key = sqlc.narg('action_key'))
  AND (sqlc.narg('entity_type')::text IS NULL OR entity_type = sqlc.narg('entity_type'))
  AND (sqlc.narg('source_service')::text IS NULL OR source_service = sqlc.narg('source_service'))
  AND (sqlc.narg('course_phase_id')::uuid IS NULL OR course_phase_id = sqlc.narg('course_phase_id'))
  AND (sqlc.narg('from_time')::timestamptz IS NULL OR created_at >= sqlc.narg('from_time'))
  AND (sqlc.narg('to_time')::timestamptz IS NULL OR created_at <= sqlc.narg('to_time'))
  AND (sqlc.narg('search')::text IS NULL OR (
        actor_name ILIKE '%' || sqlc.narg('search') || '%' ESCAPE '\' OR
        actor_email ILIKE '%' || sqlc.narg('search') || '%' ESCAPE '\' OR
        action ILIKE '%' || sqlc.narg('search') || '%' ESCAPE '\' OR
        entity_name ILIKE '%' || sqlc.narg('search') || '%' ESCAPE '\' OR
        entity_id ILIKE '%' || sqlc.narg('search') || '%' ESCAPE '\'))
  AND (sqlc.narg('cursor_ts')::timestamptz IS NULL OR
        (created_at, id) < (sqlc.narg('cursor_ts')::timestamptz, sqlc.narg('cursor_id')::uuid))
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('page_limit');
