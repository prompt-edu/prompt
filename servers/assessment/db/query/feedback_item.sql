-- name: GetFeedbackItem :one
SELECT *
FROM feedback_items
WHERE id = $1;

-- name: CreateFeedbackItem :exec
INSERT INTO feedback_items (id,
                            feedback_type,
                            feedback_text,
                            course_participation_id,
                            course_phase_id,
                            author_course_participation_id,
                            type)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: UpdateFeedbackItem :execrows
UPDATE feedback_items
SET feedback_type = $4,
    feedback_text = $5
WHERE id = $1
  AND course_phase_id = $2
  AND author_course_participation_id = $3;

-- name: DeleteFeedbackItem :exec
DELETE
FROM feedback_items
WHERE id = $1;

-- name: ListFeedbackItemsForParticipantInPhase :many
SELECT *
FROM feedback_items
WHERE course_participation_id = $1
  AND course_phase_id = $2
  AND type != 'tutor'
ORDER BY created_at;

-- name: ListFeedbackItemsForTutorInPhase :many
SELECT *
FROM feedback_items
WHERE feedback_items.course_participation_id = $1
  AND course_phase_id = $2
  AND type = 'tutor'
ORDER BY created_at;

-- name: ListFeedbackItemsForCoursePhase :many
SELECT *
FROM feedback_items
WHERE course_phase_id = $1
ORDER BY created_at;

-- name: ListFeedbackItemsByAuthorInPhase :many
SELECT *
FROM feedback_items
WHERE author_course_participation_id = $1
  AND course_phase_id = $2
ORDER BY created_at;
