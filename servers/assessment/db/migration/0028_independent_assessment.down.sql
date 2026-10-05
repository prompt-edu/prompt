BEGIN;

DROP TABLE IF EXISTS independent_assessment;

ALTER TABLE course_phase_config
    DROP COLUMN IF EXISTS independent_assessment_enabled;

COMMIT;
