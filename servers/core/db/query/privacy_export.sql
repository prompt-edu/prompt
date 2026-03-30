-- name: GetExportRecordByID :one
SELECT * FROM privacy_export WHERE id = $1 LIMIT 1;

-- name: GetExportRecordByIDWithDocs :one
SELECT * FROM privacy_export_with_docs WHERE id = $1;

-- name: GetLatestExportForUserWithDocs :one
SELECT * FROM privacy_export_with_docs WHERE user_id = $1 ORDER BY date_created DESC LIMIT 1;

-- name: CreateNewExport :one
INSERT INTO privacy_export ( id, user_id, student_id, status, valid_until )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: CreateNewExportDoc :one
INSERT INTO privacy_export_document ( id, export_id, source_name, object_key, status )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: SetExportStatus :one
UPDATE privacy_export SET status = $2 WHERE id = $1 RETURNING *;

-- name: SetExportDocStatus :one
UPDATE privacy_export_document SET status = $2 WHERE id = $1 RETURNING *;

-- name: SetExportDocFileSize :one
UPDATE privacy_export_document SET file_size = $2 WHERE id = $1 RETURNING *;

-- name: GetExportDocObjectKey :one
SELECT object_key FROM privacy_export_document WHERE id = $1;

-- name: SetExportDocDownloadedAt :exec
UPDATE privacy_export_document SET downloaded_at = now() WHERE id = $1 AND downloaded_at IS NULL;

-- name: GetLatestExportForUserForUpdate :one
SELECT * FROM privacy_export WHERE user_id = $1 ORDER BY date_created DESC LIMIT 1 FOR UPDATE;

-- name: GetAllExports :many
SELECT
  e.*,
  COUNT(CASE WHEN ed.status = 'complete' THEN 1 END)::int AS total_docs,
  COUNT(CASE WHEN ed.downloaded_at IS NOT NULL THEN 1 END)::int AS downloaded_docs,
  MAX(ed.downloaded_at)::timestamptz AS last_downloaded_at
FROM privacy_export e
LEFT JOIN privacy_export_document ed ON ed.export_id = e.id
GROUP BY e.id
ORDER BY e.date_created DESC;

-- name: GetInvalidExports :many
SELECT * FROM privacy_export WHERE now() >= valid_until AND status != 'archived';

-- name: GetExportDocObjectKeysByExportID :many
SELECT object_key FROM privacy_export_document WHERE export_id = $1;

-- name: SetExportDocStatusByExportID :exec
UPDATE privacy_export_document SET status = $2 WHERE export_id = $1;

