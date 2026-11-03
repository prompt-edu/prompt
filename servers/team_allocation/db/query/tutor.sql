-- name: GetTutorIDByLogin :one
-- The id a tutor already carries in this phase, so a re-import keeps the evaluations attached to it.
SELECT course_participation_id
FROM tutor
WHERE course_phase_id = $1
  AND university_login = $2
LIMIT 1;

-- name: DeleteTutorAssignments :exec
-- Removes every team assignment of one tutor, matched by login or by id, so an import replaces them.
DELETE
FROM tutor
WHERE course_phase_id = @course_phase_id
  AND (university_login = @university_login OR course_participation_id = @course_participation_id);

-- name: InsertTutor :exec
INSERT INTO tutor (course_phase_id, course_participation_id, first_name, last_name, team_id, university_login)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetTutorsByCourseParticipationID :many
SELECT t.*
FROM tutor t
WHERE t.course_participation_id = $1
  AND t.course_phase_id = $2
ORDER BY t.team_id;
