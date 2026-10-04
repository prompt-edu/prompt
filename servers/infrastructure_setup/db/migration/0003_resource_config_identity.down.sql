BEGIN;

ALTER TABLE resource_config
    DROP CONSTRAINT IF EXISTS uq_resource_config_identity;

COMMIT;
