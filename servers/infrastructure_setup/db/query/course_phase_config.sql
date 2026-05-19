-- name: GetCoursePhaseConfig :one
SELECT course_phase_id, team_source_course_phase_id, student_source_course_phase_id, semester_tag
FROM course_phase_config
WHERE course_phase_id = $1;

-- name: UpsertCoursePhaseConfig :one
INSERT INTO course_phase_config (
    course_phase_id,
    team_source_course_phase_id,
    student_source_course_phase_id,
    semester_tag
)
VALUES ($1, $2, $3, $4)
ON CONFLICT (course_phase_id)
DO UPDATE SET
    team_source_course_phase_id = EXCLUDED.team_source_course_phase_id,
    student_source_course_phase_id = EXCLUDED.student_source_course_phase_id,
    semester_tag = EXCLUDED.semester_tag
RETURNING course_phase_id, team_source_course_phase_id, student_source_course_phase_id, semester_tag;
