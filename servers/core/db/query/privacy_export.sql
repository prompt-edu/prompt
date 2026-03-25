-- name: GetExportRecordByID :one
SELECT * FROM privacy_export WHERE id = $1 LIMIT 1;

-- name: GetExportRecordByIDWithDocs :one
SELECT * FROM privacy_export_with_docs WHERE id = $1;

-- name: GetLatestExportForUserWithDocs :one
SELECT * FROM privacy_export_with_docs WHERE userID = $1 ORDER BY date_created DESC LIMIT 1;

-- name: CreateNewExport :one
INSERT INTO privacy_export ( id, userID, studentID, status, valid_until )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: CreateNewExportDoc :one
INSERT INTO privacy_export_document ( id, exportID, source_name, object_key, status )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: SetExportStatus :one
UPDATE privacy_export SET status = $2 WHERE id = $1 RETURNING *;

-- name: SetExportDocStatus :one
UPDATE privacy_export_document SET status = $2 WHERE id = $1 RETURNING *;

-- name: UpdateExportDocResult :one
UPDATE privacy_export_document SET status = $2, file_size = $3 WHERE id = $1 RETURNING *;

-- name: SetExportDocDownloadedAt :exec
UPDATE privacy_export_document SET downloaded_at = now() WHERE id = $1 AND downloaded_at IS NULL;
