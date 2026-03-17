-- name: CreateFile :one
INSERT INTO files (
    filename,
    original_filename,
    content_type,
    size_bytes,
    storage_key,
    storage_provider,
    uploaded_by_user_id,
    uploaded_by_email,
    course_phase_id,
    description,
    tags
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetFileByID :one
SELECT * FROM files
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetFileByStorageKey :one
SELECT * FROM files
WHERE storage_key = $1 AND deleted_at IS NULL;

-- name: GetFilesByCoursePhaseID :many
SELECT * FROM files
WHERE course_phase_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetFilesByUploader :many
SELECT * FROM files
WHERE uploaded_by_user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetAllFiles :many
SELECT * FROM files
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateFileMetadata :one
UPDATE files
SET 
    description = COALESCE(sqlc.narg('description'), description),
    tags = COALESCE(sqlc.narg('tags'), tags),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteFile :exec
UPDATE files
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: HardDeleteFile :exec
DELETE FROM files
WHERE id = $1;

-- name: CountFilesByUploader :one
SELECT COUNT(*) FROM files
WHERE uploaded_by_user_id = $1 AND deleted_at IS NULL;

-- name: GetTotalFileSizeByUploader :one
SELECT COALESCE(SUM(size_bytes), 0) as total_size
FROM files
WHERE uploaded_by_user_id = $1 AND deleted_at IS NULL;

-- name: GetFilesByTags :many
SELECT * FROM files
WHERE tags && $1::VARCHAR[] AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
