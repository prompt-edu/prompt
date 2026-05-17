CREATE EXTENSION IF NOT EXISTS pgcrypto;

DROP TABLE IF EXISTS resource_instance CASCADE;
DROP TABLE IF EXISTS resource_config CASCADE;
DROP TABLE IF EXISTS provider_config CASCADE;
DROP TABLE IF EXISTS course_phase_config CASCADE;
DROP TYPE IF EXISTS resource_status CASCADE;
DROP TYPE IF EXISTS resource_scope CASCADE;
DROP TYPE IF EXISTS provider_type CASCADE;

CREATE TYPE provider_type AS ENUM ('gitlab', 'slack', 'outline', 'rancher', 'keycloak');
CREATE TYPE resource_scope AS ENUM ('per_team', 'per_student');
CREATE TYPE resource_status AS ENUM ('pending', 'in_progress', 'created', 'failed');

CREATE TABLE course_phase_config (
    course_phase_id                uuid PRIMARY KEY,
    team_source_course_phase_id    uuid,
    student_source_course_phase_id uuid,
    semester_tag                   text NOT NULL DEFAULT ''
);

CREATE TABLE provider_config (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    course_phase_id uuid NOT NULL,
    provider_type   provider_type NOT NULL,
    credentials     bytea NOT NULL DEFAULT ''::bytea,
    CONSTRAINT uq_provider_config_phase_type UNIQUE (course_phase_id, provider_type)
);

CREATE TABLE resource_config (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    course_phase_id       uuid NOT NULL,
    provider_type         provider_type NOT NULL,
    resource_type         text NOT NULL,
    scope                 resource_scope NOT NULL,
    name_template         text NOT NULL,
    permission_mapping    jsonb NOT NULL DEFAULT '{}',
    resource_extra_config jsonb NOT NULL DEFAULT '{}',
    created_at            timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_resource_config_provider
        FOREIGN KEY (course_phase_id, provider_type)
        REFERENCES provider_config (course_phase_id, provider_type) ON DELETE CASCADE
);

CREATE TABLE resource_instance (
    id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_config_id      uuid NOT NULL REFERENCES resource_config(id) ON DELETE CASCADE,
    course_phase_id         uuid NOT NULL,
    team_id                 uuid,
    course_participation_id uuid,
    status                  resource_status NOT NULL DEFAULT 'pending',
    external_id             text,
    external_url            text,
    error_message           text,
    retry_count             integer NOT NULL DEFAULT 0,
    created_at              timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uq_resource_instance_team
    ON resource_instance (resource_config_id, team_id)
    WHERE team_id IS NOT NULL AND status != 'failed';

CREATE UNIQUE INDEX uq_resource_instance_student
    ON resource_instance (resource_config_id, course_participation_id)
    WHERE course_participation_id IS NOT NULL AND status != 'failed';
