-- name: GetTeamsByCoursePhase :many
SELECT *
FROM team
WHERE course_phase_id = $1
ORDER BY name;

-- name: GetTeamsWithStudentNames :many
SELECT
  t.id,
  t.name,
  COALESCE(members.team_members, '[]'::jsonb) AS team_members,
  COALESCE(tutors.team_tutors, '[]'::jsonb) AS team_tutors
FROM
  team t
LEFT JOIN LATERAL (
  SELECT jsonb_agg(
    jsonb_build_object(
      'id', a.course_participation_id,
      'firstName', a.student_first_name,
      'lastName', a.student_last_name
    )
    ORDER BY a.student_first_name, a.student_last_name
  ) AS team_members
  FROM assignments a
  WHERE a.team_id = t.id
) members ON TRUE
LEFT JOIN LATERAL (
  SELECT jsonb_agg(
    jsonb_build_object(
      'id', tu.course_participation_id,
      'firstName', tu.first_name,
      'lastName', tu.last_name
    )
    ORDER BY tu.first_name, tu.last_name
  ) AS team_tutors
  FROM tutor tu
  WHERE tu.team_id = t.id
) tutors ON TRUE
WHERE
  t.course_phase_id = $1
ORDER BY
  t.name;

-- name: GetTeamWithStudentNamesByTeamID :one
SELECT
  t.id,
  t.name,
  COALESCE(members.team_members, '[]'::jsonb) AS team_members,
  COALESCE(tutors.team_tutors, '[]'::jsonb) AS team_tutors
FROM
  team t
LEFT JOIN LATERAL (
  SELECT jsonb_agg(
    jsonb_build_object(
      'id', a.course_participation_id,
      'firstName', a.student_first_name,
      'lastName', a.student_last_name
    )
    ORDER BY a.student_first_name, a.student_last_name
  ) AS team_members
  FROM assignments a
  WHERE a.team_id = t.id
) members ON TRUE
LEFT JOIN LATERAL (
  SELECT jsonb_agg(
    jsonb_build_object(
      'id', tu.course_participation_id,
      'firstName', tu.first_name,
      'lastName', tu.last_name
    )
    ORDER BY tu.first_name, tu.last_name
  ) AS team_tutors
  FROM tutor tu
  WHERE tu.team_id = t.id
) tutors ON TRUE
WHERE
  t.course_phase_id = $1
  AND t.id = $2;


-- name: CreateTeam :exec
INSERT INTO team (id, name, course_phase_id, created_by_course_participation_id)
VALUES ($1, $2, $3, sqlc.narg(created_by_course_participation_id));

-- name: LockTeamCreation :exec
-- Serializes team creation within a course phase so the creation limits below cannot be
-- bypassed by concurrent requests. Released automatically at the end of the transaction.
SELECT pg_advisory_xact_lock(hashtextextended((sqlc.arg(course_phase_id)::uuid)::text, 0));

-- name: CountTeamsByCoursePhase :one
SELECT COUNT(*)
FROM team
WHERE course_phase_id = $1;

-- name: CountTeamsCreatedByParticipation :one
SELECT COUNT(*)
FROM team
WHERE course_phase_id = $1
  AND created_by_course_participation_id = $2;

-- name: UpdateTeam :exec
UPDATE team
SET name = $3
WHERE id = $1
AND course_phase_id = $2; -- ensuring to update only teams in the authenticated course_phase

-- name: DeleteTeam :exec
DELETE FROM team
WHERE id = $1
AND course_phase_id = $2; -- ensuring to delete only teams in the authenticated course_phase