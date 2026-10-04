BEGIN;

ALTER TABLE course_phase_config
    DROP COLUMN IF EXISTS tutor_display_name;

COMMIT;
