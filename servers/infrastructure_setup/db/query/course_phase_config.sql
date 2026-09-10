-- name: GetCoursePhaseConfig :one
SELECT course_phase_id, semester_tag
FROM course_phase_config
WHERE course_phase_id = $1;

-- name: CopyCoursePhaseConfig :exec
-- Creates the target phase's config row, carrying the source semester tag when there is
-- one and falling back to ''. Never overwrites a config already edited on the target.
INSERT INTO course_phase_config (course_phase_id, semester_tag)
VALUES (
    sqlc.arg(target_course_phase_id),
    COALESCE((
        SELECT src.semester_tag FROM course_phase_config AS src
        WHERE src.course_phase_id = sqlc.arg(source_course_phase_id)
    ), '')
)
ON CONFLICT (course_phase_id) DO NOTHING;

-- name: UpsertCoursePhaseConfig :one
INSERT INTO course_phase_config (
    course_phase_id,
    semester_tag
)
VALUES ($1, $2)
ON CONFLICT (course_phase_id)
DO UPDATE SET
    semester_tag = EXCLUDED.semester_tag
RETURNING course_phase_id, semester_tag;
