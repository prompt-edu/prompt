-- name: GetOpenDeletionRequestForUser :one
SELECT * FROM privacy_deletion_request
WHERE user_id = $1 AND status IN ('pending_approval', 'in_progress')
LIMIT 1;

-- name: GetLatestDeletionRequestForUserWithSubrequests :one
SELECT * FROM privacy_deletion_request_with_subrequests
WHERE user_id = $1
ORDER BY requested_at DESC
LIMIT 1;

-- name: GetDeletionRequestByID :one
SELECT * FROM privacy_deletion_request WHERE id = $1;

-- name: GetDeletionRequestByIDWithSubrequests :one
SELECT * FROM privacy_deletion_request_with_subrequests WHERE id = $1;

-- name: CreateNewDeletionRequest :one
INSERT INTO privacy_deletion_request (id, user_id, student_id, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: SetDeletionRequestAuditor :exec
UPDATE privacy_deletion_request
SET auditor_id           = $2,
    auditor_name         = $3,
    auditor_email        = $4,
    auditor_note         = $5,
    auditor_responded_at = now()
WHERE id = $1;

-- name: SetDeletionRequestStatus :one
UPDATE privacy_deletion_request
SET status       = $2,
    completed_at = CASE WHEN $2::privacy_deletion_request_status IN ('rejected', 'succeeded', 'failed') THEN now() ELSE completed_at END
WHERE id = $1
RETURNING *;

-- name: CreateNewDeletionSubrequest :one
INSERT INTO privacy_deletion_subrequest (id, deletion_request_id, source_name, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAllDeletionRequests :many
SELECT
  v.*,
  s.first_name AS student_first_name,
  s.last_name  AS student_last_name,
  s.email      AS student_email
FROM privacy_deletion_request_with_subrequests v
LEFT JOIN student s ON s.id = v.student_id
ORDER BY v.requested_at DESC;

-- name: SetDeletionSubrequestStatus :one
UPDATE privacy_deletion_subrequest
SET status        = $2,
    completed_at  = CASE WHEN $2::privacy_deletion_subrequest_status IN ('succeeded', 'failed') THEN now() ELSE completed_at END,
    error_message = $3
WHERE id = $1
RETURNING *;
