BEGIN;

ALTER TABLE resource_config
    ADD CONSTRAINT uq_resource_config_identity
    UNIQUE (course_phase_id, provider_type, resource_type, scope, name_template);

COMMIT;
