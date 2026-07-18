-- name: CreateResourceConfig :one
INSERT INTO resource_config (
    id,
    course_phase_id,
    provider_type,
    resource_type,
    scope,
    name_template,
    permission_mapping,
    resource_extra_config
)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
RETURNING id, course_phase_id, provider_type, resource_type, scope, name_template, permission_mapping, resource_extra_config, created_at;

-- name: GetResourceConfig :one
SELECT id, course_phase_id, provider_type, resource_type, scope, name_template, permission_mapping, resource_extra_config, created_at
FROM resource_config
WHERE id = $1 AND course_phase_id = $2;

-- name: ListResourceConfigs :many
SELECT id, course_phase_id, provider_type, resource_type, scope, name_template, permission_mapping, resource_extra_config, created_at
FROM resource_config
WHERE course_phase_id = $1
ORDER BY created_at;

-- name: UpdateResourceConfig :one
UPDATE resource_config
SET resource_type = $3,
    scope = $4,
    name_template = $5,
    permission_mapping = $6,
    resource_extra_config = $7
WHERE id = $1 AND course_phase_id = $2
RETURNING id, course_phase_id, provider_type, resource_type, scope, name_template, permission_mapping, resource_extra_config, created_at;

-- name: DeleteResourceConfig :exec
DELETE FROM resource_config
WHERE id = $1 AND course_phase_id = $2;

-- name: CopyResourceConfigs :exec
INSERT INTO resource_config (
    id,
    course_phase_id,
    provider_type,
    resource_type,
    scope,
    name_template,
    permission_mapping,
    resource_extra_config
)
SELECT gen_random_uuid(), sqlc.arg(target_course_phase_id), rc.provider_type, rc.resource_type, rc.scope, rc.name_template, rc.permission_mapping, rc.resource_extra_config
FROM resource_config AS rc
WHERE rc.course_phase_id = sqlc.arg(source_course_phase_id);
