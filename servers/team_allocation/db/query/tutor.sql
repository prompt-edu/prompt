-- name: UpsertTutor :exec
INSERT INTO tutor (course_phase_id, course_participation_id, first_name, last_name, team_id, university_login)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (course_phase_id, course_participation_id) DO UPDATE
    SET team_id          = EXCLUDED.team_id,
        university_login = EXCLUDED.university_login,
        first_name       = EXCLUDED.first_name,
        last_name        = EXCLUDED.last_name;

-- name: GetTutorByCourseParticipationID :one
SELECT t.*
FROM tutor t
WHERE t.course_participation_id = $1
  AND t.course_phase_id = $2;

-- name: GetTutorByTeamID :one
SELECT t.*
FROM tutor t
WHERE t.team_id = $1
  AND t.course_phase_id = $2;

-- name: GetTutorTeamByUniversityLogin :one
SELECT team_id
FROM tutor
WHERE course_phase_id = $1
  AND university_login = $2;

-- name: UpdateTutorTeam :execrows
UPDATE tutor
SET team_id = $3
WHERE course_phase_id = $1
  AND university_login = $2;
