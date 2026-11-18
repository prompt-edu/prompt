-- name: GetTutorIDByLogin :one
-- The id a tutor already carries in this phase, so a re-import keeps the evaluations attached to it.
SELECT course_participation_id
FROM tutor
WHERE course_phase_id = $1
  AND university_login = $2
ORDER BY course_participation_id
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

-- name: GetTutorAssignmentsByCoursePhase :many
-- One row per tutor and team, with the login that identifies the tutor across import sources.
SELECT course_participation_id, university_login, team_id
FROM tutor
WHERE course_phase_id = $1
ORDER BY university_login, course_participation_id, team_id;

-- name: GetTutorByUniversityLogin :one
SELECT t.course_participation_id, t.first_name, t.last_name
FROM tutor t
WHERE t.course_phase_id = $1
  AND t.university_login = $2
ORDER BY t.course_participation_id
LIMIT 1;
