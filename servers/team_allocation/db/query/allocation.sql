-- name: GetAllocationsByCoursePhase :many
SELECT a.*
FROM allocations a
WHERE a.course_phase_id = $1;

-- name: CreateOrUpdateAllocation :exec
INSERT INTO allocations AS a (
  id,
  course_participation_id,
  team_id,
  course_phase_id,
  created_at,
  updated_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
ON CONFLICT ON CONSTRAINT allocations_participation_phase_uk
DO UPDATE
SET team_id = EXCLUDED.team_id,
    updated_at = CURRENT_TIMESTAMP;

-- name: DeleteAllocationsByPhase :exec
DELETE
FROM allocations a
    USING team t
WHERE a.team_id = t.id
  AND t.course_phase_id = $1;

-- name: GetAllocationForStudent :one
SELECT id,
       course_participation_id,
       student_first_name,
       student_last_name,
       team_id,
       course_phase_id,
       created_at,
       updated_at
FROM allocations
WHERE course_participation_id = $1
  AND course_phase_id = $2;

-- name: GetAggregatedAllocationsByCoursePhase :many
SELECT team_id,
       array_agg(course_participation_id ORDER BY course_participation_id)::uuid[] AS student_ids
FROM allocations
WHERE course_phase_id = $1
GROUP BY team_id
ORDER BY team_id;

-- name: GetTeamsWithStudentNames :many
SELECT
  t.id,
  t.name,
  COALESCE(
    jsonb_agg(
      jsonb_build_object(
        'courseParticipationID', a.course_participation_id,
        'studentFirstName',           a.student_first_name,
        'studentLastName',            a.student_last_name
      )
      ORDER BY a.student_first_name
    ) FILTER (WHERE a.id IS NOT NULL),
    '[]'::jsonb
  )::jsonb AS team_members
FROM
  team t
LEFT JOIN
  allocations a
  ON t.id = a.team_id
WHERE
  t.course_phase_id = $1
GROUP BY
  t.id, t.name
ORDER BY
  t.name;

-- name: GetTeamWithStudentNamesByTeamID :one
SELECT
  t.id,
  t.name,
  COALESCE(
    jsonb_agg(
      jsonb_build_object(
        'courseParticipationID', a.course_participation_id,
        'studentFirstName',           a.student_first_name,
        'studentLastName',            a.student_last_name
      )
      ORDER BY a.student_first_name
    ) FILTER (WHERE a.id IS NOT NULL),
    '[]'::jsonb
  )::jsonb AS team_members
FROM
  team t
LEFT JOIN
  allocations a
  ON t.id = a.team_id
WHERE
  t.course_phase_id = $1
  AND t.id = $2
GROUP BY
  t.id, t.name
ORDER BY
  t.name;

-- name: UpdateStudentNameForAllocation :exec
UPDATE allocations
SET student_first_name = $1,
    student_last_name = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE course_participation_id = $3
  AND course_phase_id = $4;

-- name: UpsertAllocationForParticipant :execrows
-- The COALESCE guard scopes the write to an expected source team: a NULL
-- expected_team_id writes unconditionally, a set one only updates a row that is
-- still in that team, so authorization and mutation stay a single statement.
INSERT INTO allocations AS a (
  id,
  course_participation_id,
  team_id,
  course_phase_id,
  student_first_name,
  student_last_name,
  created_at,
  updated_at
) VALUES (
  @id,
  @course_participation_id,
  @team_id,
  @course_phase_id,
  @student_first_name,
  @student_last_name,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
ON CONFLICT ON CONSTRAINT allocations_participation_phase_uk
DO UPDATE
SET team_id = EXCLUDED.team_id,
    student_first_name = EXCLUDED.student_first_name,
    student_last_name = EXCLUDED.student_last_name,
    updated_at = CURRENT_TIMESTAMP
WHERE a.team_id = COALESCE(sqlc.narg('expected_team_id')::uuid, a.team_id);

-- name: DeleteAllocationForParticipant :execrows
DELETE
FROM allocations a
WHERE a.course_participation_id = @course_participation_id
  AND a.course_phase_id = @course_phase_id
  AND a.team_id = COALESCE(sqlc.narg('expected_team_id')::uuid, a.team_id);
