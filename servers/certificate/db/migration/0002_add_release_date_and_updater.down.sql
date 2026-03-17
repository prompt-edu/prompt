BEGIN;

ALTER TABLE course_phase_config
DROP COLUMN IF EXISTS updated_by,
DROP COLUMN IF EXISTS release_date;

COMMIT;