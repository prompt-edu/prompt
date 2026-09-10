BEGIN;

DROP TABLE IF EXISTS resource_instance;
DROP TABLE IF EXISTS resource_config;
DROP TABLE IF EXISTS provider_config;
DROP TABLE IF EXISTS course_phase_config;

DROP TYPE IF EXISTS resource_status;
DROP TYPE IF EXISTS resource_scope;
DROP TYPE IF EXISTS provider_type;

COMMIT;
