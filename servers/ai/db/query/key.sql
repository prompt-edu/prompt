-- name: GetPhaseKey :one
SELECT *
FROM ai_phase_key
WHERE course_phase_id = $1;

-- name: UpsertPhaseKey :one
INSERT INTO ai_phase_key (course_phase_id, phase_type, encrypted_key, last4, set_by)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (course_phase_id) DO UPDATE
SET phase_type    = EXCLUDED.phase_type,
    encrypted_key = EXCLUDED.encrypted_key,
    last4         = EXCLUDED.last4,
    set_by        = EXCLUDED.set_by,
    set_at        = now()
RETURNING *;

-- name: DeletePhaseKey :exec
DELETE
FROM ai_phase_key
WHERE course_phase_id = $1;
