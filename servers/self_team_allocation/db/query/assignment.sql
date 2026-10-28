-- name: CreateOrUpdateAssignment :exec
INSERT INTO assignments AS a (
  id,
  course_participation_id,
  student_first_name,
  student_last_name,
  team_id,
  course_phase_id,
  created_at,
  updated_at
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
ON CONFLICT ON CONSTRAINT assignments_participation_phase_uk
  DO UPDATE
SET team_id = EXCLUDED.team_id,
  updated_at = CURRENT_TIMESTAMP;


-- name: DeleteAssignment :exec
DELETE FROM assignments
WHERE course_participation_id = $1
  AND course_phase_id = $2
  AND team_id = $3;

-- name: GetAssignmentsByCoursePhase :many
SELECT a.*
FROM assignments a
WHERE a.course_phase_id = $1;

-- name: GetAssignmentForStudent :one
SELECT id,
       course_participation_id,
       student_first_name,
       student_last_name,
       team_id,
       course_phase_id,
       created_at,
       updated_at
FROM assignments
WHERE course_participation_id = $1
  AND course_phase_id = $2;

-- name: LockTeamForAssignment :many
-- Locks the team row so concurrent joins are counted one after another. Returns no row when
-- the team does not exist in the course phase.
SELECT id
FROM team
WHERE id = $1
  AND course_phase_id = $2
FOR UPDATE;

-- name: CountTeamMembers :one
SELECT COUNT(*)
FROM assignments
WHERE team_id = $1;
