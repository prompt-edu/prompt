-- name: GetAssignmentsByParticipationIDs :many
SELECT
    a.id,
    a.course_participation_id,
    a.course_phase_id,
    a.student_first_name,
    a.student_last_name,
    a.created_at,
    a.updated_at,
    a.team_id,
    t.name           AS team_name,
    t.created_at     AS team_created_at
FROM assignments a, team t
WHERE a.course_participation_id = ANY(@course_participation_ids::uuid[])
  AND t.id = a.team_id
ORDER BY a.created_at ASC;

-- name: GetTutorsByCourseParticipationIDs :many
SELECT *
FROM tutor
WHERE course_participation_id = ANY(@course_participation_ids::uuid[]);
