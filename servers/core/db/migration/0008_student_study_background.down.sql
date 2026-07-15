BEGIN;

DROP TRIGGER IF EXISTS set_last_modified_course_phase_participation ON course_phase_participation;
DROP TRIGGER IF EXISTS set_last_modified_student ON student;
DROP FUNCTION IF EXISTS update_last_modified_column();

ALTER TABLE course_phase_participation DROP COLUMN IF EXISTS last_modified;

ALTER TABLE student
    DROP COLUMN IF EXISTS study_program,
    DROP COLUMN IF EXISTS study_degree,
    DROP COLUMN IF EXISTS current_semester,
    DROP COLUMN IF EXISTS last_modified;

DROP TYPE IF EXISTS study_degree;

COMMIT;
