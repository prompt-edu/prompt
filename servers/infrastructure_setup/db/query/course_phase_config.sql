-- name: GetCoursePhaseConfig :one
SELECT course_phase_id, semester_tag
FROM course_phase_config
WHERE course_phase_id = $1;

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
