-- name: GetAssignmentsByParticipationIDs :many
SELECT *
FROM assignments
WHERE course_participation_id = ANY(@course_participation_ids::uuid[])
ORDER BY created_at ASC;
