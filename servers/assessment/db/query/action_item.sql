-- name: GetActionItem :one
SELECT *
FROM action_item
WHERE id = $1;

-- name: ListActionItemsForCoursePhase :many
SELECT *
FROM action_item
WHERE course_phase_id = $1
ORDER BY created_at;

-- name: GetAllActionItemsForCoursePhaseCommunication :many
SELECT course_participation_id, ARRAY_AGG(action ORDER BY created_at)::TEXT[] AS action_items
FROM action_item
WHERE course_phase_id = $1
GROUP BY course_participation_id;

-- name: GetStudentActionItemsForCoursePhaseCommunication :many
SELECT action
FROM action_item
WHERE course_participation_id = $1
  AND course_phase_id = $2
ORDER BY created_at;

-- name: CreateActionItem :exec
INSERT INTO action_item (id,
                         course_phase_id,
                         course_participation_id,
                         action,
                         author)
VALUES ($1, $2, $3, $4, $5);

-- name: UpdateActionItem :exec
UPDATE action_item
SET course_phase_id         = $2,
    course_participation_id = $3,
    action                  = $4,
    author                  = $5
WHERE id = $1;

-- name: DeleteActionItem :exec
DELETE
FROM action_item
WHERE id = $1;

-- name: ListActionItemsForStudentInPhase :many
SELECT *
FROM action_item
WHERE course_participation_id = $1
  AND course_phase_id = $2
ORDER BY created_at;

-- name: CountActionItemsForStudentInPhase :one
SELECT COUNT(*) AS action_item_count
FROM action_item
WHERE course_participation_id = $1
  AND course_phase_id = $2;

