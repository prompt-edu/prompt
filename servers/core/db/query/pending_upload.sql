-- name: CreatePendingUpload :exec
INSERT INTO pending_upload (storage_key)
VALUES ($1);

-- name: DeletePendingUpload :exec
DELETE FROM pending_upload
WHERE storage_key = $1;

-- name: ListExpiredPendingUploads :many
SELECT storage_key
FROM pending_upload
WHERE created_at < $1
ORDER BY created_at
LIMIT $2;
