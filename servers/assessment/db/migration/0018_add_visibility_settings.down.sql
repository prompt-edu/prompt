BEGIN;

ALTER TABLE course_phase_config
    DROP COLUMN IF EXISTS grade_suggestion_visible,
    DROP COLUMN IF EXISTS action_items_visible;

COMMIT;
