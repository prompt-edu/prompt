-- name: UpsertProviderConfig :one
INSERT INTO provider_config (id, course_phase_id, provider_type, credentials)
VALUES (gen_random_uuid(), $1, $2, $3)
ON CONFLICT (course_phase_id, provider_type)
DO UPDATE SET credentials = EXCLUDED.credentials
RETURNING id, course_phase_id, provider_type, credentials;

-- name: GetProviderConfig :one
SELECT id, course_phase_id, provider_type, credentials
FROM provider_config
WHERE course_phase_id = $1 AND provider_type = $2;

-- name: ListProviderConfigs :many
SELECT id, course_phase_id, provider_type, credentials
FROM provider_config
WHERE course_phase_id = $1
ORDER BY provider_type;

-- name: DeleteProviderConfig :exec
DELETE FROM provider_config
WHERE course_phase_id = $1 AND provider_type = $2;

-- name: CopyProviderConfigsWithEmptyCredentials :exec
INSERT INTO provider_config (id, course_phase_id, provider_type, credentials)
SELECT gen_random_uuid(), sqlc.arg(target_course_phase_id), pc.provider_type, ''::bytea
FROM provider_config AS pc
WHERE pc.course_phase_id = sqlc.arg(source_course_phase_id)
ON CONFLICT (course_phase_id, provider_type) DO NOTHING;
