BEGIN;

ALTER TABLE assessment_schema
    DROP COLUMN source_phase_id;

COMMIT;
