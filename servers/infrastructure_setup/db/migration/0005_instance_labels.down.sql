BEGIN;

ALTER TABLE resource_instance
    DROP COLUMN IF EXISTS target_name,
    DROP COLUMN IF EXISTS resolved_name;

COMMIT;
