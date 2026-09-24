-- name: CopyCoursePhaseConfig :exec
-- Copies the template and the student page text of a certificate phase into another phase. The
-- release date is deliberately left out: it is printed as the certificate date, and without one
-- the copied phase stays unreleased until a lecturer sets it. Download records belong to the
-- source phase's students and are never copied. A source phase without a config row copies
-- nothing. On an existing target row the copied template replaces the target's, so its updater is
-- cleared.
INSERT INTO
    course_phase_config (
        course_phase_id,
        template_content,
        student_page_text
    )
SELECT
    sqlc.arg(target_course_phase_id)::uuid,
    template_content,
    student_page_text
FROM course_phase_config
WHERE
    course_phase_id = sqlc.arg(source_course_phase_id)::uuid
ON CONFLICT (course_phase_id) DO
UPDATE
SET
    template_content = EXCLUDED.template_content,
    student_page_text = EXCLUDED.student_page_text,
    updated_at = NOW(),
    updated_by = NULL;
