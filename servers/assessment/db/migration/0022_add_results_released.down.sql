BEGIN;

ALTER TABLE course_phase_config
    DROP COLUMN IF EXISTS grading_sheet_visible,
    DROP COLUMN IF EXISTS results_released;

COMMIT;
