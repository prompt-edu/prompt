-- name: CreateTutor :exec
INSERT INTO tutor (course_phase_id, course_participation_id, first_name, last_name, team_id)
VALUES ($1, $2, $3, $4, $5);

-- name: GetTutorIDByLogin :one
-- The id a tutor already carries in this phase, so a re-import keeps the evaluations attached to it.
SELECT course_participation_id
FROM tutor
WHERE course_phase_id = $1
  AND university_login = $2
ORDER BY course_participation_id
LIMIT 1;

-- name: LockTutorImport :exec
SELECT pg_advisory_xact_lock(hashtextextended(CAST(sqlc.arg(course_phase_id) AS uuid)::text, 0));

-- name: DeleteTutorAssignments :exec
-- Removes every team assignment of one tutor, matched by login or, for rows stored without one, by id.
DELETE
FROM tutor
WHERE course_phase_id = @course_phase_id
  AND (university_login = @university_login
    OR (university_login IS NULL AND course_participation_id = @course_participation_id));

-- name: InsertTutor :exec
INSERT INTO tutor (course_phase_id, course_participation_id, first_name, last_name, team_id, university_login)
VALUES ($1, $2, $3, $4, $5, $6);

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

-- name: DeleteTutor :exec
DELETE FROM tutor
WHERE course_participation_id = $1
  AND course_phase_id = $2;
