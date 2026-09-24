-- name: DeleteCoursePhaseConfig :exec
-- Deletes the phase's config row. feedback_category (course_phase_id) is removed by its
-- ON DELETE CASCADE constraint. feedback_answer restricts deleting a category it answers,
-- so the phase's presentations, whose feedback forms cascade to the answers, go first.
DELETE
FROM course_phase_config
WHERE course_phase_id = $1;
