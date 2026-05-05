BEGIN;

CREATE TABLE tease_workspace (
    course_phase_id        uuid PRIMARY KEY,
    constraints            jsonb NOT NULL DEFAULT '[]'::jsonb,
    locked_students        jsonb NOT NULL DEFAULT '[]'::jsonb,
    allocations_draft      jsonb NOT NULL DEFAULT '[]'::jsonb,
    algorithm_type         varchar(64),
    updated_by             uuid,
    last_saved_at          timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_exported_at       timestamptz
);

COMMIT;
