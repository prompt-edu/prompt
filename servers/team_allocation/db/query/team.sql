-- name: GetTeamsByCoursePhase :many
SELECT *
FROM team
WHERE course_phase_id = $1
ORDER BY name;

-- name: GetSurveyTeamsByCoursePhase :many
SELECT *
FROM team
WHERE course_phase_id = $1
  AND team_type IN ('standard', 'field_bucket')
ORDER BY name;

-- name: GetStandardTeamsByCoursePhase :many
SELECT *
FROM team
WHERE course_phase_id = $1
  AND team_type = 'standard'
ORDER BY name;

-- name: GetFieldBucketTeamsByCoursePhase :many
SELECT *
FROM team
WHERE course_phase_id = $1
  AND team_type = 'field_bucket'
ORDER BY name;

-- name: GetTeaseTeamsByCoursePhase :many
SELECT *
FROM team
WHERE course_phase_id = $1
  AND team_type IN ('standard', 'company_project')
ORDER BY name;

-- name: GetTeamByCoursePhaseAndTeamID :one
-- ensuring to get only teams in the authenticated course_phase
SELECT *
FROM team
WHERE id = $1
  AND course_phase_id = $2;

-- name: CreateTeam :exec
INSERT INTO team (id, name, course_phase_id, team_type, team_size_min, team_size_max)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (course_phase_id, name)
DO UPDATE SET team_type = EXCLUDED.team_type,
              team_size_min = EXCLUDED.team_size_min,
              team_size_max = EXCLUDED.team_size_max;

-- name: GetAllocationProfile :one
SELECT profile
FROM allocation_profile
WHERE course_phase_id = $1;

-- name: UpsertAllocationProfile :exec
INSERT INTO allocation_profile (course_phase_id, profile)
VALUES ($1, $2)
ON CONFLICT (course_phase_id)
DO UPDATE SET profile = EXCLUDED.profile;

-- name: UpdateTeam :exec
-- ensuring to update only teams in the authenticated course_phase
UPDATE team
SET name = $3
WHERE id = $1
  AND course_phase_id = $2;

-- name: DeleteTeam :exec
-- ensuring to delete only teams in the authenticated course_phase
DELETE
FROM team
WHERE id = $1
  AND course_phase_id = $2;

-- name: GetTeamsWithMembers :many
SELECT t.id,
       t.name,
       COALESCE(members.team_members, '[]'::jsonb) AS team_members,
       COALESCE(tutors.team_tutors, '[]'::jsonb)   AS team_tutors
FROM team t
         LEFT JOIN LATERAL (
    SELECT jsonb_agg(
                   jsonb_build_object(
                           'id', a.course_participation_id,
                           'firstName', a.student_first_name,
                           'lastName', a.student_last_name
                   )
                   ORDER BY a.student_first_name
           ) AS team_members
    FROM allocations a
    WHERE a.team_id = t.id
    ) members ON TRUE
         LEFT JOIN LATERAL (
    SELECT jsonb_agg(
                   jsonb_build_object(
                           'id', tu.course_participation_id,
                           'firstName', tu.first_name,
                           'lastName', tu.last_name
                   )
                   ORDER BY tu.first_name
           ) AS team_tutors
    FROM tutor tu
    WHERE tu.team_id = t.id
    ) tutors ON TRUE
WHERE t.course_phase_id = $1
ORDER BY t.name;
