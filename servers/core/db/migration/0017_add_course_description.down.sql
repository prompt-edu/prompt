BEGIN;

ALTER TABLE course
    DROP COLUMN IF EXISTS short_description,
    DROP COLUMN IF EXISTS long_description;

COMMIT;
