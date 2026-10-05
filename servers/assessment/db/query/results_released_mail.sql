-- name: ClaimResultsReleasedMailRecipients :many
INSERT INTO results_released_mail (course_phase_id, course_participation_id)
SELECT @course_phase_id::uuid, unnest(@course_participation_ids::uuid[])
ON CONFLICT DO NOTHING
RETURNING course_participation_id;

-- name: ReleaseResultsReleasedMailClaims :exec
DELETE
FROM results_released_mail
WHERE course_phase_id = @course_phase_id
  AND course_participation_id = ANY (@course_participation_ids::uuid[]);
