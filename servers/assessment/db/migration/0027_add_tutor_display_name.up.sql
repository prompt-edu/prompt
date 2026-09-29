BEGIN;

ALTER TABLE course_phase_config
    ADD COLUMN IF NOT EXISTS tutor_display_name TEXT;

COMMIT;
