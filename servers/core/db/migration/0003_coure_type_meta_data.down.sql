BEGIN;

ALTER TABLE course_phase_type
    DROP COLUMN IF EXISTS required_input_meta_data,
    DROP COLUMN IF EXISTS provided_output_meta_data,
    DROP COLUMN IF EXISTS initial_phase;

COMMIT;
