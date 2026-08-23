BEGIN;

ALTER TABLE course_phase_config
    DROP COLUMN IF EXISTS assessment_enabled;

COMMIT;
