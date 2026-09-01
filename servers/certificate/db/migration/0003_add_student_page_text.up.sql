BEGIN;

ALTER TABLE course_phase_config
ADD COLUMN student_page_text TEXT;

COMMIT;
