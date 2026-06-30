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

-- name: UpdateFeedbackItem :exec
UPDATE feedback_items
SET feedback_type                  = $2,
    feedback_text                  = $3,
    course_participation_id        = $4,
    course_phase_id                = $5,
    author_course_participation_id = $6,
    type                           = $7
WHERE id = $1;

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
