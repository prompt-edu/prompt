-- name: ListActiveAnnouncements :many
SELECT * FROM announcement
WHERE enabled
  AND (starts_at IS NULL OR starts_at <= now())
  AND (expires_at IS NULL OR expires_at > now())
ORDER BY severity DESC, COALESCE(starts_at, created_at) ASC;

-- name: ListAnnouncements :many
SELECT * FROM announcement
WHERE sqlc.arg(include_expired)::boolean OR expires_at IS NULL OR expires_at > now()
ORDER BY created_at DESC;

-- name: CreateAnnouncement :one
INSERT INTO announcement (
    severity, title, message, link_url, link_label, starts_at, expires_at, enabled
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdateAnnouncement :one
UPDATE announcement
SET severity = $2,
    title = $3,
    message = $4,
    link_url = $5,
    link_label = $6,
    starts_at = $7,
    expires_at = $8,
    enabled = $9,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteAnnouncement :execrows
DELETE FROM announcement WHERE id = $1;
