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

-- name: AddTutorToTeam :execrows
-- Adds one team to a tutor. A login already on the team inserts nothing, so the caller sees 0 rows;
-- the id already leading the team under another login still fails on the primary key.
INSERT INTO tutor (course_phase_id, course_participation_id, first_name, last_name, team_id, university_login)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (course_phase_id, university_login, team_id) WHERE university_login IS NOT NULL DO NOTHING;

-- name: RemoveTutorFromTeam :execrows
-- Removes one team from a tutor. A tutor left without a team is no longer a tutor of the phase.
DELETE
FROM tutor
WHERE course_phase_id = $1
  AND team_id = $2
  AND course_participation_id = $3;
