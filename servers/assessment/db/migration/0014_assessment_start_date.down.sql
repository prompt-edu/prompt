BEGIN;

ALTER TABLE course_phase_config
    DROP COLUMN start,
    DROP COLUMN self_evaluation_start,
    DROP COLUMN peer_evaluation_start;

COMMIT;
